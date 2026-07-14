package semantic

import (
	"fmt"
	"slices"
	"strings"

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

func analyzeForCondition(condition []parser.AST) {
	parts := len(condition)
	if condition[0].Label == "Variable" {
		symbol := analyzeVariable(condition[0])
		if symbol != nil {
			currentScope.variables[symbol.name] = *symbol
		}
		if condition[parts-2].Token.Value == "in" {
			if parts == 4 { // int i, char c in string
				symbol2 := analyzeVariable(condition[1])
				if symbol2 != nil {
					currentScope.variables[symbol2.name] = *symbol2
				}
				if symbol != nil && symbol2 != nil && !((slices.Contains(datatypes.IntTypes, symbol.Type) && symbol2.Type == datatypes.Char) || (symbol.Type == datatypes.Char && slices.Contains(datatypes.IntTypes, symbol2.Type))) {
					messages.Complain(diagnostic.TypeError, condition[2].Location, "Cannot use %s and %s as loop variables", symbol.Type.String(), symbol2.Type.String())
				}

			} else {
				if condition[parts-1].Label == "range" {
					if symbol != nil {
						expr, hasErr := evalType(&condition[parts-1], symbol.Type)
						if !hasErr && expr != symbol.Type {
							messages.Complain(diagnostic.TypeError, condition[parts-1].Location, "Variable of type %s not compatible with range expression of type %s", symbol.Type, expr)
						}
					}
				} else { // char c in string
					if symbol != nil && symbol.Type != datatypes.Char {
						messages.Complain(diagnostic.TypeError, condition[0].Location, "Cannot use %s as loop variables", symbol.Type.String())
					}
					rhs, hasErr := evalType(&condition[parts-1], datatypes.String)
					if !hasErr && rhs != datatypes.String {
						messages.Complain(diagnostic.TypeError, condition[parts-1].Location, "Cannot loop over %s", rhs.String())
					}
				}
			}
		} else {
			if symbol != nil {
				cond, hasErr := evalType(&condition[1], datatypes.Bool)
				if !hasErr && cond != datatypes.Bool {
					messages.Complain(diagnostic.TypeError, condition[1].Location, "Expected bool as loop condition but got %s", cond.String())
				}
				// TODO: Add support for assignment as end part
				expr, hasErr := evalType(&condition[2], symbol.Type)
				if !hasErr && symbol.Type != expr {
					messages.Complain(diagnostic.TypeError, condition[2].Location, "Expected %s as loop expression but got %s", symbol.Type.String(), cond.String())
				}
			}
		}
	} else {
		iter := analyzeAssignment(&condition[0])
		if iter != datatypes.None {
			cond, hasErr := evalType(&condition[1], datatypes.Bool)
			if !hasErr && cond != datatypes.Bool {
				messages.Complain(diagnostic.TypeError, condition[1].Location, "Expected bool as loop condition but got %s", cond.String())
			}
			expr, hasErr := evalType(&condition[2], iter)
			if !hasErr && iter != expr {
				messages.Complain(diagnostic.TypeError, condition[2].Location, "Expected %s as loop expression but got %s", iter.String(), cond.String())
			}
		}

	}
}

func analyzeAssignment(stmt *parser.AST) datatypes.DataType {
	left := stmt.Children[0]
	var lhs datatypes.DataType
	var lHasErr bool
	if left.Label == "dot" {
		lhs, lHasErr = handleDot(left.Children[0], left.Children[1], false)
		if lHasErr {
			return datatypes.None
		}
		// TODO: Fix mutability for this case
	} else {
		symbol := currentScope.lookupVariable(left.Token.Value)
		if symbol == nil {
			messages.Complain(diagnostic.NameError, left.Location, "Variable '%s' is not defined in this scope", left.Token.Value)
			return datatypes.None
		}
		if !symbol.isMutable {
			messages.Complain(diagnostic.AccessError, left.Location, "Cannot change value of immutable variable '%s'", left.Token.Value)
			return datatypes.None
		}
		lhs = symbol.Type
	}
	rhs, hasError := evalType(&stmt.Children[1], lhs)
	if hasError {
		return datatypes.None
	}
	stmt.Children[0].Type = lhs
	operator := stmt.Token.Value
	switch operator {
	case "+=":
		if lhs == datatypes.String {
			if rhs != datatypes.Char && rhs != datatypes.String {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot concatenate %s to String", rhs.String())
			}
		} else if slices.Contains(datatypes.NumericTypes, lhs) {
			result, err := decideNumberType(lhs, rhs, operator)
			if err != nil {
				messages.Complain(diagnostic.TypeError, stmt.Location, "%s", err.Error())
			} else if result != lhs {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
			}
		} else {
			messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Invalid operation between %s to %s", rhs.String(), lhs)
		}
	case "-=", "*=", "/=":
		if slices.Contains(datatypes.NumericTypes, lhs) {
			result, err := decideNumberType(lhs, rhs, operator)
			if err != nil {
				messages.Complain(diagnostic.TypeError, stmt.Location, "%s", err.Error())
			} else if result != lhs {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
			}
		} else {
			messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Invalid operation between %s to %s", rhs.String(), lhs)
		}
	default:
		if lhs != rhs {
			if !ImplementsInterface(lhs, rhs) {
				messages.Complain(diagnostic.TypeError, stmt.Children[1].Location, "Cannot assign %s to %s", rhs, lhs)
			}
		} else if intf := globalScope.lookupInterface(rhs.String()); intf != nil {
			messages.Complain(diagnostic.IllegalStatementError, stmt.Children[1].Location, "Cannot use interface type as value")
		}

	}
	stmt.Children[1].Type = rhs
	return lhs
}

func analyzeNamedBlock(nbNode parser.AST, structName string, impl []string) *NamedBlockSymbol {
	details := nbNode.Children
	name := details[0].Token.Value
	if !slices.Contains(specialBlocks, name) && !slices.Contains(impl, name) {
		messages.Complain(diagnostic.NameError, nbNode.Location, "Block '%s' not supported", name)
		return nil
	}
	body := details[1].Children
	scope := currentScope
	var newScope *Scope = nil
	if name != "private" {
		newScope = currentScope.addChild(fmt.Sprintf("%s@%s", name, currentScope.id), NamedBlock)
		currentScope = newScope
	}
	for _, node := range body {
		switch node.Label {
		case "fn":
			symbol := processFunctionSignature(node)
			switch name {
			case "compare":
				supported := []string{
					fmt.Sprintf("fn equals(%s)->bool", structName),
					fmt.Sprintf("fn lessThan(%s)->bool", structName),
					fmt.Sprintf("fn greaterThan(%s)->bool", structName),
				}
				if !slices.Contains(supported, symbol.getSignature()) {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Function signature '%s' not supported; only '%s' supported", symbol.getSignature(), strings.Join(supported, ","))
				}
			case "cast":
				if len(symbol.parameters) > 0 || symbol.returnType == datatypes.None || symbol.returnType == datatypes.DynamicType(structName) {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Functions in cast block must take no parameters and return a different type")
				} else if intf := globalScope.lookupInterface(symbol.returnType.String()); intf != nil {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Functions in cast block cannot return an interface")
				} else if matches := newScope.lookupFunctionsByReturnType(symbol.returnType); len(matches) > 0 {
					messages.Complain(diagnostic.AmbiguityError, node.Location, "Cannot have more than one cast function that returns %s", symbol.returnType)
				}
			case "private":
				symbol.isPrivate = true
				currentScope = scope
				// private is not a real named block; it is only a shortcut to mark everything in it as private
				if err := currentScope.functions.add(symbol); err != nil {
					messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
				}
				continue
			}
			if err := currentScope.functions.add(symbol); err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
			}
		case "Variable":
			if name != "private" {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "Variable declaration only allowed in struct or private block")
			} else {
				currentScope = scope
				symbol := analyzeVariable(node)
				if symbol != nil {
					if symbol.isPrivate {
						messages.Complain(diagnostic.Warning, node.Location, "Redundant use of private in private block")
					}
					symbol.isPrivate = true
					currentScope.variables[symbol.name] = *symbol
				}
				continue
			}
		}
	}
	currentScope = scope
	if newScope == nil {
		return nil
	}
	return &NamedBlockSymbol{
		name:           name,
		isSpecialBlock: slices.Contains(specialBlocks, name),
		Def:            &nbNode,
		innerScope:     newScope,
	}
}

func analyzeVariable(varNode parser.AST) *VariableSymbol {
	details := varNode.Children
	typeNode := details[0]
	name := details[1].Token
	isPrivate := false
	isMutable := false
	var rhs *parser.AST = nil
	if details[0].Label == "modifiers" {
		typeNode = details[1]
		name = details[2].Token
		for _, modifier := range details[0].Children {
			if modifier.Token.Value == "private" {
				isPrivate = true
			}
			if modifier.Token.Value == "mut" {
				isMutable = true
			}
		}
		// handle value
		if len(details) == 4 {
			rhs = &details[3]
		}
	} else {
		if len(details) == 3 {
			rhs = &details[2]
		}
	}
	varType := nodeToType(typeNode)
	if currentScope.lookupVariable(name.Value) != nil {
		messages.Complain(diagnostic.NameError, name.Location, "Name: '%s' already defined", name.Value)
		return nil
	}
	if rhs != nil {
		if rType, hasError := evalType(rhs, varType); !hasError && (rType != varType && !ImplementsInterface(varType, rType)) {
			messages.Complain(diagnostic.TypeError, rhs.Location, "Cannot assign type %s to variable type %s", rType, varType)
		}
	}
	return &VariableSymbol{
		name:        name.Value,
		Type:        varType,
		isPrivate:   isPrivate,
		isMutable:   isMutable,
		Def:         &varNode,
		Initialized: rhs != nil,
	}
}
