package parser

import (
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
)

type (
	AST struct {
	}
)

// program     =   { declaration } ;
func Parse(tokens []lexer.Token) (AST, diagnostic.PhaseDiagnostics) {
	root := AST{}
	report := diagnostic.PhaseDiagnostics{}
	// TODO
	return root, report
}
