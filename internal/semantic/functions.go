package semantic

import (
	"fmt"

	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
)

var (
	ifBlockCounter int
	whileCounter   int
	forCounter     int
)

func processFunctionSignature(fnNode parser.AST) FnCreateSymbol {
	details := fnNode.Children
	length := len(details)
	name := details[0].Token.Value
	ifBlockCounter = 0
	whileCounter = 0
	forCounter = 0
	var paramNode *parser.AST = nil
	var returnTypeNode *parser.AST = nil
	var bodyNode *parser.AST = nil
	var newScope *Scope = nil
	switch length {
	case 2:
		if details[1].Label == "params" { // fn name(type pname);
			paramNode = &details[1]
		} else if details[1].Token.Kind == lexer.ID || details[1].Token.Kind == lexer.KW_TYPE { // fn name() -> type;
			returnTypeNode = &details[1]
		} else { // fn name() {}
			bodyNode = &details[1]
		}
	case 3:
		if details[1].Token.Kind == lexer.ID || details[1].Token.Kind == lexer.KW_TYPE { // fn name() -> type {}
			returnTypeNode = &details[1]
			bodyNode = &details[2]
		} else {
			paramNode = &details[1]
			if details[2].Token.Kind == lexer.ID || details[2].Token.Kind == lexer.KW_TYPE { // fn name(type pname) -> type;
				returnTypeNode = &details[2]
			} else { // fn name(type pname) {}
				bodyNode = &details[2]
			}

		}
	case 4: // fn name(type pname) -> type {}
		paramNode = &details[1]
		returnTypeNode = &details[2]
		bodyNode = &details[3]
	}
	var paramNames []string
	var paramTypes []datatypes.DataType
	var returnType datatypes.DataType = datatypes.None
	if returnTypeNode != nil {
		returnType = nodeToType(*returnTypeNode)
	}
	if paramNode != nil {
		for _, param := range paramNode.Children {
			paramTypes = append(paramTypes, nodeToType(param.Children[0]))
			paramNames = append(paramNames, param.Children[1].Token.Value)
		}
	}
	if bodyNode != nil {
		paramList := datatypes.Join(paramTypes)
		scopeId := fmt.Sprintf("@%s(%s)", name, paramList)
		if currentScope.id != "@global" {
			scopeId = fmt.Sprintf("%s(%s)@%s", name, paramList, currentScope.id)
		}
		newScope = currentScope.addChild(scopeId, Function)
		for i := range len(paramNames) {
			newScope.variables[paramNames[i]] = VariableSymbol{
				name: paramNames[i],
				Type: paramTypes[i],
			}
		}
	}
	symbol := FnCreateSymbol{
		name:                     name,
		parameters:               paramTypes,
		returnType:               returnType,
		hasDefaultImplementation: bodyNode != nil,
		Body:                     bodyNode,
		innerScope:               newScope,
	}
	return symbol
}

func analyzeFunctionBody(fn FunctionSymbol) {
	scope := currentScope
	for _, overload := range fn.overloads {
		if !overload.hasDefaultImplementation {
			continue
		}
		currentScope = overload.innerScope
		params := datatypes.Join(overload.parameters)
		returnStr := ""
		if fn.returnType != datatypes.None {
			returnStr = fmt.Sprintf("->%s", fn.returnType)
		}
		sig := fmt.Sprintf("%s(%s)%s", fn.name, params, returnStr)
		hasReturn := analyzeBlockAndCheckForReturn(overload.Body.Children, fn, sig)
		if !hasReturn && fn.returnType != datatypes.None {
			messages.Complain(diagnostic.TypeError, overload.Body.Location, "Function '%s' may not return a value", sig)
		}
	}
	currentScope = scope
}

func analyzeBlockAndCheckForReturn(body []parser.AST, fn FunctionSymbol, sig string) bool {
	hasValidReturn := false
	length := len(body)
	for i, stmt := range body {
		if stmt.Label == "Variable" {
			symbol := analyzeVariable(stmt)
			if symbol != nil {
				if symbol.isPrivate {
					messages.Complain(diagnostic.IllegalStatementError, stmt.Location, "Cannot set private variable in function body")
				} else {
					currentScope.variables[symbol.name] = *symbol
				}
			}
		} else if stmt.Token.Kind == lexer.OPERATOR_ASSIGN {
			analyzeAssignment(&stmt)
		} else if stmt.Label == "control-flow" {
			if i != length-1 {
				messages.Warn(stmt.Location, "Unreachable code found after statement")
			}
			if len(stmt.Children) == 1 && stmt.Children[0].Token.Value == "return" {
				if fn.returnType != datatypes.None {
					messages.Complain(diagnostic.TypeError, stmt.Location, "Function '%s' missing return value, expected: %s", sig, fn.returnType)
				} else {
					stmt.Type = datatypes.None
					hasValidReturn = true
				}
			} else if len(stmt.Children) == 1 { // continue and break
				if !currentScope.HasScopeTypeAncestor(Loop) {
					messages.Complain(diagnostic.IllegalStatementError, stmt.Location, "Cannot use %s outside of loop", stmt.Children[0].Token.Value)
				}
			} else { // return something
				rhs, rHasErr := evalType(&stmt.Children[1], fn.returnType)
				if !rHasErr && rhs != fn.returnType && !ImplementsInterface(fn.returnType, rhs) {
					messages.Complain(diagnostic.TypeError, stmt.Location, "Function '%s' expected return type %s but got %s", sig, fn.returnType, rhs)
				} else {
					stmt.Type = rhs
					hasValidReturn = true
				}
			}
		} else if stmt.Label == "if-block" {
			scope := currentScope
			elseIfCounter := 0
			for i, branch := range stmt.Children {
				currentScope = scope
				var id string
				if branch.Label == "else if" {
					id = fmt.Sprintf("%s#%d.%d@%s", branch.Label, ifBlockCounter, elseIfCounter, currentScope.id)
					elseIfCounter++
				} else {
					id = fmt.Sprintf("%s#%d@%s", branch.Label, ifBlockCounter, currentScope.id)
				}
				branchScope := currentScope.addChild(id, Branch)
				currentScope = branchScope
				block_index := 0
				if branch.Label != "else" {
					condition, hasErr := evalType(&branch.Children[0], datatypes.Bool)
					if !hasErr && condition != datatypes.Bool {
						messages.Complain(diagnostic.TypeError, branch.Children[0].Location, "Expected bool but got %s", condition)
					}
					block_index = 1
				}
				returns := analyzeBlockAndCheckForReturn(branch.Children[block_index].Children, fn, sig)
				if i == 0 {
					hasValidReturn = returns
				} else {
					hasValidReturn = hasValidReturn && returns
				}
			}
			currentScope = scope
			ifBlockCounter++

		} else if stmt.Label == "for" {
			scope := currentScope
			newScope := currentScope.addChild(fmt.Sprintf("for#%d@%s", forCounter, currentScope.id), Loop)
			forCounter++
			currentScope = newScope
			analyzeForCondition(stmt.Children[0].Children)
			analyzeBlockAndCheckForReturn(stmt.Children[1].Children, fn, sig)
			currentScope = scope
		} else if stmt.Label == "while" {
			scope := currentScope
			newScope := currentScope.addChild(fmt.Sprintf("while#%d@%s", whileCounter, currentScope.id), Loop)
			whileCounter++
			currentScope = newScope
			cond, hasError := evalType(&stmt.Children[0], datatypes.Bool)
			if !hasError && cond != datatypes.Bool {
				messages.Complain(diagnostic.TypeError, stmt.Children[0].Location, "Expected bool as loop condition but got %s", cond.String())
			}
			analyzeBlockAndCheckForReturn(stmt.Children[1].Children, fn, sig)
			currentScope = scope

		} else {
			stmt.Type, _ = evalType(&stmt, datatypes.None) // expressions
		}
	}
	return hasValidReturn
}
