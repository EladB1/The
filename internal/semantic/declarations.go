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

func processFunctionSignature(fnNode parser.AST) FnCreateSymbol {
	details := fnNode.Children
	length := len(details)
	name := details[0].Token.Value
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
		scopeId := fmt.Sprintf("@%s", name)
		if currentScope.id != "@global" {
			scopeId = fmt.Sprintf("%s@%s", name, currentScope.id)
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
		returnStr := ""
		if fn.returnType != datatypes.None {
			returnStr = fmt.Sprintf("->%s", fn.returnType)
		}
		params := datatypes.Join(overload.parameters)
		sig := fmt.Sprintf("%s(%s)%s", fn.name, params, returnStr)
		if !overload.hasDefaultImplementation {
			continue
		}
		currentScope = overload.innerScope
		length := len(overload.Body.Children)
		for i, stmt := range overload.Body.Children {
			if stmt.Label == "Variable" {
				symbol := analyzeVariable(stmt)
				if symbol != nil {
					currentScope.variables[symbol.name] = *symbol
				}
			} else if stmt.Token.Kind == lexer.OPERATOR_ASSIGN {
				// TODO
			} else if stmt.Label == "control-flow" {
				if i != length-1 {
					messages.Warn(stmt.Location, "Unreachable code found after statement")
				}
				// TODO: continue and break
				if len(stmt.Children) == 1 && stmt.Children[0].Token.Value == "return" {
					if fn.returnType != datatypes.None {
						messages.Complain(diagnostic.TypeError, stmt.Location, "Function '%s' missing return value, expected: %s", sig, fn.returnType)

					} else {
						stmt.Type = datatypes.None
					}
				} else if len(stmt.Children) == 1 { // continue and break

				} else { // return something
					rhs, rHasErr := evalType(&stmt.Children[1], fn.returnType)
					if !rHasErr && rhs != fn.returnType {
						messages.Complain(diagnostic.TypeError, stmt.Location, "Function '%s' expected return type %s but got %s", sig, fn.returnType, rhs)
					} else {
						stmt.Type = rhs
					}
				}
			} else if stmt.Label == "if-block" {

			} else if stmt.Label == "for" {

			} else if stmt.Label == "while" {

			} else {
				stmt.Type, _ = evalType(&stmt, datatypes.None) // expressions
			}
		}
	}
	currentScope = scope
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
		if rType, hasError := evalType(rhs, varType); !hasError && rType != varType {
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
