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
		newScope = currentScope.addChild(scopeId)
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
	// scope := currentScope
	// currentScope = fn.innerScope
	// for _, node := range fn.Body.Children {
	// 	if node.Label == "Variable" {
	// 		symbol := analyzeVariable(node)
	// 		if symbol != nil {
	// 			currentScope.variables[symbol.name] = *symbol
	// 		}
	// 	}
	// }
	// currentScope = scope
}

func analyzeNamedBlock(nbNode parser.AST, structName string, impl []string) *NamedBlockSymbol {
	details := nbNode.Children
	name := details[0].Token.Value
	if !slices.Contains(specialBlocks, name) && !slices.Contains(impl, name) {
		messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Block '%s' not supported", name), nbNode.Location)
		return nil
	}
	body := details[1].Children
	scope := currentScope
	newScope := currentScope.addChild(fmt.Sprintf("%s@%s", name, currentScope.id))
	currentScope = newScope
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
					messages = messages.Complain(diagnostic.NamedBlockError, fmt.Sprintf("Function signature '%s' not supported; only '%s' supported", symbol.getSignature(), strings.Join(supported, ",")), node.Location)
				}
			case "cast":
				if len(symbol.parameters) > 0 || symbol.returnType == datatypes.None || symbol.returnType == datatypes.DynamicType(structName) {
					messages = messages.Complain(diagnostic.NamedBlockError, "Functions in cast block must take no parameters and return a different type", node.Location)
				}
			case "private":
				symbol.isPrivate = true
			}
			if err := currentScope.functions.add(symbol); err != nil {
				messages = messages.Complain(diagnostic.IllegalStatementError, err.Error(), node.Location)
			}
		case "Variable":
			if name != "private" {
				messages = messages.Complain(diagnostic.IllegalStatementError, "Variable declaration only allowed in struct or private block", node.Location)
			} else {
				symbol := analyzeVariable(node)
				if symbol.isPrivate {
					messages = messages.Complain(diagnostic.Warning, "Redundant use of private in private block", node.Location)
				}
				currentScope.variables[symbol.name] = *symbol

			}
		}
	}
	currentScope = scope
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
		messages = messages.Complain(diagnostic.NameError, fmt.Sprintf("Name: '%s' already defined", name.Value), name.Location)
		return nil
	}
	if rhs != nil {
		if evalType(rhs, varType) != varType {
			// TODO
			fmt.Println("HI", name.Value, rhs.Type)
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
