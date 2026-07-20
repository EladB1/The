package semantic

import (
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/parser"
)

func analyzeVariable(varNode *parser.AST) *VariableSymbol {
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
			rhs = details[3]
		}
	} else {
		if len(details) == 3 {
			rhs = details[2]
		}
	}
	varType := nodeToType(typeNode)
	if currentScope.LookupVariable(name.Value) != nil {
		messages.Complain(diagnostic.NameError, name.Location, "Name: '%s' already defined", name.Value)
		return nil
	}
	if rhs != nil {
		if rType, hasError := evalType(rhs, varType); !hasError && (rType != varType && !isCompatibleType(varType, rType)) {
			messages.Complain(diagnostic.TypeError, rhs.Location, "Cannot assign type %s to variable type %s", rType, varType)
		}
	}
	varNode.Type = varType
	return &VariableSymbol{
		Name:        name.Value,
		Type:        varType,
		isPrivate:   isPrivate,
		isMutable:   isMutable,
		Def:         varNode,
		Initialized: rhs != nil,
	}
}
