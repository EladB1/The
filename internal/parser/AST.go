package parser

import (
	"fmt"
	"strings"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/datatypes"
	"github.com/EladB1/The/internal/lexer"
)

type (
	AST struct {
		Label    string
		Token    lexer.Token
		Location ds.SourceLocation
		Type     datatypes.DataType
		Children []AST
	}
)

func (ast AST) String(pool ds.LiteralPool) string {
	return ast.to_string(0, pool)
}

func (ast AST) to_string(indentLevel int, pool ds.LiteralPool) string {
	prefix := strings.Repeat("\t", indentLevel)
	builder := strings.Builder{}
	builder.WriteString(prefix)
	builder.WriteString("Node: { ")
	if ast.Label != "" {
		builder.WriteString(fmt.Sprintf("Label: \"%s\"", ast.Label))
	} else {
		builder.WriteString(fmt.Sprintf("Token: %v", ast.Token.String(pool)))
	}
	if ast.Type != nil {
		builder.WriteString(fmt.Sprintf(", Type: %v", ast.Type))
	}
	childCount := len(ast.Children)
	if childCount > 0 {
		builder.WriteString("\n")
		builder.WriteString(prefix)
		builder.WriteString("children: [\n")
		for i, child := range ast.Children {
			builder.WriteString(child.to_string(indentLevel+1, pool))
			if i != childCount-1 {
				builder.WriteString(",\n")
			}
		}
		builder.WriteString("\n")
		builder.WriteString(prefix)
		builder.WriteString("]")
	}
	builder.WriteString(" }")
	return builder.String()
}

func (ast *AST) PrependChildren(nodes ...AST) {
	ast.Children = append(nodes, ast.Children...)
}

func (ast *AST) AddChildren(nodes ...AST) {
	ast.Children = append(ast.Children, nodes...)
}

func (ast *AST) AddChildToken(token lexer.Token) {
	ast.AddChildren(AST{Token: token, Location: token.Location})
}

func nodeFromToken(token lexer.Token) AST {
	return AST{Token: token, Location: token.Location}
}

func (ast *AST) IsLiteral() bool {
	return (ast.Label == "struct_literal" ||
		ast.Token.Kind == lexer.LIT_CHAR ||
		ast.Token.Kind == lexer.LIT_STRING ||
		ast.Token.Kind == lexer.LIT_FLOAT ||
		ast.Token.Kind == lexer.LIT_INT ||
		ast.Token.Kind == lexer.LIT_HEX ||
		ast.Token.Kind == lexer.KW_BOOLVALUE)
}

func (ast *AST) HasChildren() bool {
	return len(ast.Children) != 0
}
