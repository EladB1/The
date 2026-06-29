package semantic

import (
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/parser"
)

var messages diagnostic.PhaseDiagnostics = diagnostic.PhaseDiagnostics{}

func Analyze(ast parser.AST) (parser.AST, diagnostic.PhaseDiagnostics) {
	return parser.AST{}, messages
}
