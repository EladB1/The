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
var snapsDir string = "testdata/semantic-snaps"

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") != "true" {
		t.Skip()
	}
	subdirs, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	for _, pass := range subdirs {
		if !pass.IsDir() {
			continue
		}
		testpath := filepath.Join(dir, pass.Name())
		fixtures := testutils.GetSourceFromDirectory(t, testpath)
		for _, fixture := range fixtures {
			tokens, _ := lexer.Lex(fixture.Source, false)
			ast, _ := parser.Parse(tokens)
			testutils.WriteResultToFile(ast, testpath, fixture.File)
		}
	}

}

func loadAST(t *testing.T, testdir string, filename string) parser.AST {
	path := filepath.Join(testdir, filename)
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

func snapshotTestSemanticAnalyzer(t *testing.T, filename string, subdir string) {
	snapTarget := filepath.Join(snapsDir, subdir)
	testdir := filepath.Join(dir, subdir)
	snapshots := snaps.WithConfig(
		snaps.Dir(snapTarget),
	)
	ast := loadAST(t, testdir, filename)
	result, messages := Analyze(ast)
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
	results := fmt.Sprintf("AST:\n%v\n, Compiler messages:\n[%s]\n", result, msgBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

func TestPassOne(t *testing.T) {
	subdir := "pass-1"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should errors.the and have errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}

func TestPassTwo(t *testing.T) {
	subdir := "pass-2"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should errors.the and have errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}

func TestPassThree(t *testing.T) {
	subdir := "pass-3"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should run warnings.the and have warnings", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "warnings.json", subdir)
	})
	t.Run("should errors.the and have a mix of errors and warnings", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}

func TestPassFour(t *testing.T) {
	subdir := "pass-4"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should errors.the and have a mix of errors and warnings", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}

func TestPassFive(t *testing.T) {
	subdir := "pass-5"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should run warnings.the and have warnings", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "warnings.json", subdir)
	})
	t.Run("should errors.the and have a mix of errors and warnings", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}
