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

	"github.com/EladB1/The/internal/filehandler"
	"github.com/EladB1/The/internal/lexer"
)

var dir string = "testdata/fixtures/"

func snapshotTestParser(t *testing.T, filename string, debug bool) {
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/parser-snaps"),
	)
	tokens := loadTokens(t, filename)
	ast, messages := Parse(tokens)
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
	results := fmt.Sprintf("AST:\n%v\n, Compiler messages:\n[%s]\n", ast, msgBuilder.String())
	if debug {
		lexer.PrintTokens(tokens)
	}
	snapshots.MatchSnapshot(t, results)
}

func loadTokens(t *testing.T, filename string) []lexer.Token {
	path := filepath.Join(dir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %s\n%v", filename, err)
		os.Exit(1)
	}
	var tokens []lexer.Token
	err = json.Unmarshal(content, &tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to unmarshal json", err)
		os.Exit(1)
	}
	return tokens
}

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") != "true" {
		t.Skip()
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read directory\n", err)
		os.Exit(1)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".the") {
			fmt.Printf("Updating fixture for '%s'... ", file.Name())
			path := filepath.Join(dir, file.Name())
			src, err := filehandler.GetSourceCode(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to read source file\n", err)
				os.Exit(1)
			}
			tokens, _ := lexer.Lex(src, false)
			result, err := json.Marshal(tokens)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to marshal json\n", err)
				os.Exit(1)
			}
			path = filepath.Join(dir, strings.ReplaceAll(file.Name(), ".the", ".json"))
			err = os.WriteFile(path, result, 0664)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to write file\n", err)
				os.Exit(1)
			}
			fmt.Println("Fixture updated")
		}
	}
}

func TestParser(t *testing.T) {
	t.Run("should run on empty file with no errors", func(t *testing.T) {
		token := lexer.Token{
			Kind:   lexer.EOF,
			Line:   0,
			Column: 0,
		}
		ast, messages := Parse([]lexer.Token{token})
		if len(messages) != 0 {
			t.Errorf("Expected no warnings or errors but got %v\n", messages)
			os.Exit(1)
		}
		emptyAST := AST{label: "program"}
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
