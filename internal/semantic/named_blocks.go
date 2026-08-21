package semantic

import (
	"fmt"
	"slices"
	"strings"

	dt "github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/parser"
)

func analyzeNamedBlock(nbNode *parser.AST, structName string, impl []string, offset *uint32) (*NamedBlockSymbol, int, []string) {
	extraSize := 0
	extraProps := []string{}
	details := nbNode.Children
	name := details[0].Token.Value
	if strings.HasPrefix(name, "__") {
		messages.Complain(diagnostic.NameError, nbNode.Location, "Name '%s' uses reserved prefix '__'", name)
		return nil, extraSize, extraProps
	}
	if !slices.Contains(specialBlocks, name) && !slices.Contains(impl, name) {
		messages.Complain(diagnostic.NameError, nbNode.Location, "Block '%s' not supported", name)
		return nil, extraSize, extraProps
	}
	body := details[1].Children
	scope := currentScope
	var newScope *Scope = nil
	if name != "private" {
		newScope = currentScope.addChild(fmt.Sprintf("%s@%s", name, currentScope.Id), NamedBlock)
		nbNode.IRName = newScope.Id
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
				if !symbol.hasDefaultImplementation {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Compare block function '%s' must be defined with a function body", symbol.getSignature())
				}
				node.IRName = symbol.getIRName(false)
			case "cast":
				if len(symbol.parameters) > 0 || symbol.returnType.Equals(dt.NoneType) || symbol.returnType.Equals(dt.NewDynamicType(structName)) {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Functions in cast block must take no parameters and return a different type")
				} else if intf := globalScope.LookupInterface(symbol.returnType.String()); intf != nil {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Functions in cast block cannot return an interface")
				} else if matches := newScope.LookupFunctionsByReturnType(symbol.returnType); len(matches) > 0 {
					messages.Complain(diagnostic.AmbiguityError, node.Location, "Cannot have more than one cast function that returns %s", symbol.returnType)
				} else if !symbol.hasDefaultImplementation {
					messages.Complain(diagnostic.NamedBlockError, node.Location, "Cast block function '%s' must be defined with a function body", symbol.getSignature())
				} else {
					node.IRName = symbol.getIRName(false)
				}
			case "private":
				symbol.isPrivate = true
				currentScope = scope
				// private is not a real named block; it is only a shortcut to mark everything in it as private
				if overload, err := create(currentScope.Functions, symbol); err != nil {
					messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
				} else {
					node.IRName = overload.IRName
				}
				continue
			}
			if overload, err := create(currentScope.Functions, symbol); err != nil {
				messages.Complain(diagnostic.IllegalStatementError, node.Location, "%s", err.Error())
			} else {
				node.IRName = overload.IRName
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
					symbol.Ctx = StructProp
					extraProps = append(extraProps, symbol.Name)
					size := symbol.Type.GetSizeInBytes()
					symbol.Offset.IsSet = true
					symbol.Offset.Value = *offset
					*offset = *offset + uint32(size)
					extraSize += size
					currentScope.Variables.Add(*symbol, symbol.Name)
				}
				continue
			}
		}
	}
	currentScope = scope
	if newScope == nil {
		return nil, extraSize, extraProps
	}
	return &NamedBlockSymbol{
		Name:           name,
		isSpecialBlock: slices.Contains(specialBlocks, name),
		Def:            nbNode,
		InnerScope:     newScope,
	}, extraSize, extraProps
}
