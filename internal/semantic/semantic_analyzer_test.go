package semantic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
)

var dir string = "./testdata/fixtures"

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") != "true" {
		t.Skip()
	}
	fixtures := testutils.GetSourceFromDirectory(t, dir)
	for _, fixture := range fixtures {
		tokens, _ := lexer.Lex(fixture.Source, false)
		ast, _ := parser.Parse(tokens)
		testutils.WriteResultToFile(ast, dir, fixture.File)
	}
}

func loadAST(t *testing.T, filename string) parser.AST {
	path := filepath.Join(dir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %s\n%v", filename, err)
		os.Exit(1)
	}
	var ast parser.AST
	err = json.Unmarshal(content, &ast)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to unmarshal json", err)
		os.Exit(1)
	}
	return ast
}

func snapshotTestSemanticAnalyzer(t *testing.T, filename string) {
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/semantic-snaps"),
	)
	ast := loadAST(t, filename)
	result, messages := Analyze(ast)
	var msgBuilder strings.Builder
	var formatStr string
	for i, msg := range messages {
		if i != len(messages)-1 {
			formatStr = fmt.Sprintf("\n\t\"%v\",", msg)
		} else {
			formatStr = fmt.Sprintf("\n\t\"%v\"\n", msg)
		}
		msgBuilder.WriteString(formatStr)
	}
	results := fmt.Sprintf("AST:\n%v\n, Compiler messages:\n[%s]\n", result, msgBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

func TestSemanticAnalyzer(t *testing.T) {

}
