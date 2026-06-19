package parser

import (
	"fmt"
	"strings"

	"github.com/EladB1/The/internal/lexer"
)

type (
	AST struct {
		label    string
		token    lexer.Token
		children []AST
	}
)

func (ast AST) String() string {
	return ast.to_string(0)
}

func (ast AST) to_string(indentLevel int) string {
	prefix := strings.Repeat("\t", indentLevel)
	builder := strings.Builder{}
	builder.WriteString(prefix)
	builder.WriteString("Node: { ")
	if ast.label != "" {
		builder.WriteString(fmt.Sprintf("Label: \"%s\"", ast.label))
	} else {
		builder.WriteString(fmt.Sprintf("Token: %v", ast.token))
	}
	childCount := len(ast.children)
	if childCount > 0 {
		builder.WriteString("\n")
		builder.WriteString(prefix)
		builder.WriteString("children: [\n")
		for i, child := range ast.children {
			builder.WriteString(child.to_string(indentLevel + 1))
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

func (ast *AST) AddChildren(nodes ...AST) {
	ast.children = append(ast.children, nodes...)
}

func (ast *AST) AddChildToken(token lexer.Token) {
	ast.AddChildren(AST{token: token})
}
