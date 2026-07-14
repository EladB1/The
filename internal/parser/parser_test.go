package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/testutils"
)

type Fixture struct {
	Tokens   []lexer.Token
	Literals ds.LiteralPool
}

var dir string = "./testdata/fixtures/"

func snapshotTestParser(t *testing.T, filename string, debug bool) {
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/parser-snaps"),
	)
	fixture := loadFixture(t, filename)
	ast, messages := Parse(fixture.Tokens, fixture.Literals)
	var msgBuilder strings.Builder
	var formatStr string
	for i, msg := range messages.Messages {
		if i != len(messages.Messages)-1 {
			formatStr = fmt.Sprintf("\n\t\"%v\",", msg)
		} else {
			formatStr = fmt.Sprintf("\n\t\"%v\"\n", msg)
		}
		msgBuilder.WriteString(formatStr)
	}
	results := fmt.Sprintf("AST:\n%v\n, Compiler messages:\n[%s]\n", ast.String(fixture.Literals), msgBuilder.String())
	if debug {
		lexer.PrintTokens(fixture.Tokens, fixture.Literals)
	}
	snapshots.MatchSnapshot(t, results)
}

func loadFixture(t *testing.T, filename string) Fixture {
	path := filepath.Join(dir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %s\n%v", filename, err)
		os.Exit(1)
	}
	var fixture Fixture
	err = json.Unmarshal(content, &fixture)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to unmarshal json", err)
		os.Exit(1)
	}
	return fixture
}

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") != "true" {
		t.Skip()
	}
	fixtures := testutils.GetSourceFromDirectory(t, dir)
	for _, fixture := range fixtures {
		tokens, pool, _ := lexer.Lex(fixture.Source, false)
		fix := Fixture{
			Tokens:   tokens,
			Literals: pool,
		}
		testutils.WriteResultToFile(fix, dir, fixture.File)
	}
}

func TestParser(t *testing.T) {
	t.Run("should run on empty file with no errors", func(t *testing.T) {
		token := lexer.Token{
			Kind: lexer.EOF,
			Location: ds.SourceLocation{
				Line:   0,
				Column: 0,
			},
		}
		ast, messages := Parse([]lexer.Token{token}, ds.LiteralPool{})
		if len(messages.Messages) != 0 {
			t.Errorf("Expected no warnings or errors but got %v\n", messages)
			os.Exit(1)
		}
		emptyAST := AST{Label: "program"}
		if !reflect.DeepEqual(ast, emptyAST) {
			t.Errorf("Expected empty AST but got %v\n", ast)
			os.Exit(1)
		}
	})
	t.Run("should run variables.the and have no errors", func(t *testing.T) {
		snapshotTestParser(t, "variables.json", false)
	})
	t.Run("should run variables_errors.the and have errors", func(t *testing.T) {
		snapshotTestParser(t, "variables_errors.json", false)
	})
	t.Run("should run functions.the and have no errors", func(t *testing.T) {
		snapshotTestParser(t, "functions.json", false)
	})
	t.Run("should run functions_errors.the and have errors", func(t *testing.T) {
		snapshotTestParser(t, "functions_errors.json", false)
	})
	t.Run("should run structs.the and have no errors", func(t *testing.T) {
		snapshotTestParser(t, "structs.json", false)
	})
	t.Run("should run structs_errors.the and have errors", func(t *testing.T) {
		snapshotTestParser(t, "structs_errors.json", false)
	})
	t.Run("should run branch.the and have no errors", func(t *testing.T) {
		snapshotTestParser(t, "branch.json", false)
	})

	t.Run("should run branch_errors.the and have errors", func(t *testing.T) {
		snapshotTestParser(t, "branch_errors.json", false)
	})

}
