package semantic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/testutils"
)

var dir string = "./testdata/fixtures"
var snapsDir string = "testdata/semantic-snaps"

type Fixture struct {
	AST      parser.AST
	Literals ds.LiteralPool
}

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
			tokens, pool, _ := lexer.Lex(fixture.Source, false)
			ast, _ := parser.Parse(tokens, pool)
			fix := Fixture{
				AST:      ast,
				Literals: pool,
			}
			testutils.WriteResultToFile(fix, testpath, fixture.File)
		}
	}

}

func loadFixture(t *testing.T, testdir string, filename string) Fixture {
	path := filepath.Join(testdir, filename)
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

func snapshotTestSemanticAnalyzer(t *testing.T, filename string, subdir string) {
	snapTarget := filepath.Join(snapsDir, subdir)
	testdir := filepath.Join(dir, subdir)
	snapshots := snaps.WithConfig(
		snaps.Dir(snapTarget),
	)
	fixture := loadFixture(t, testdir, filename)
	scopeTree, messages := Analyze(&fixture.AST)
	var msgBuilder strings.Builder
	delim := ","
	for i, msg := range messages.Messages {
		if i == len(messages.Messages)-1 {
			delim = "\n"
		}
		msgBuilder.WriteString(fmt.Sprintf("\n\t\"%v\"%s", msg, delim))
	}
	results := fmt.Sprintf("AST:\n%v\nScopeTree:\n%v\nCompiler messages:\n[%s]\n", fixture.AST.String(fixture.Literals), scopeTree, msgBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

func TestInterfaces(t *testing.T) {
	subdir := "interfaces"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should errors.the and have errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}

func TestStructs(t *testing.T) {
	subdir := "structs"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should run warnings.the and have warnings", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "warnings.json", subdir)
	})
	t.Run("should errors.the and have errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}

func TestVariables(t *testing.T) {
	subdir := "variables"
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

func TestFunctions(t *testing.T) {
	subdir := "functions"
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

func TestBranch(t *testing.T) {
	subdir := "branch"
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

func TestTypeSystem(t *testing.T) {
	subdir := "typecheck"
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

func TestStrings(t *testing.T) {
	subdir := "strings"
	t.Run("should run valid.the and have no errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "valid.json", subdir)
	})
	t.Run("should errors.the and have errors", func(t *testing.T) {
		snapshotTestSemanticAnalyzer(t, "errors.json", subdir)
	})
}
