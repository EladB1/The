package irgen

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
	"github.com/EladB1/The/internal/semantic"
	"github.com/EladB1/The/internal/testutils"
)

var dir string = "./testdata/fixtures"
var snapsDir string = "testdata/irgen-snaps"

type Fixture struct {
	ScopeTree *semantic.Scope
	AST       parser.AST
	Literals  ds.LiteralPool
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
	for _, subdir := range subdirs {
		if !subdir.IsDir() {
			continue
		}
		testpath := filepath.Join(dir, subdir.Name())
		fixtures := testutils.GetSourceFromDirectory(t, testpath)
		for _, fixture := range fixtures {
			tokens, pool, _ := lexer.Lex(fixture.Source, false)
			ast, _ := parser.Parse(tokens, pool)
			scopes, _ := semantic.Analyze(&ast)
			fix := Fixture{
				ScopeTree: scopes,
				AST:       ast,
				Literals:  pool,
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

func snapshotTestIRGenerator(t *testing.T, filename string, subdir string) {
	snapTarget := filepath.Join(snapsDir, subdir)
	testdir := filepath.Join(dir, subdir)
	snapshots := snaps.WithConfig(
		snaps.Dir(snapTarget),
	)
	fixture := loadFixture(t, testdir, filename)
	prog, messages := Generate(fixture.AST, fixture.ScopeTree)
	var msgBuilder strings.Builder
	delim := ","
	for i, msg := range messages.Messages {
		if i == len(messages.Messages)-1 {
			delim = ""
		}
		msgBuilder.WriteString(fmt.Sprintf("\n\t\"%v\"%s", msg, delim))
	}
	results := fmt.Sprintf("IR:\n%v\nCompiler messages:\n[%s\n]", prog, msgBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

// func TestLiteralsAndSimpleAssignments(t *testing.T) {
// 	subdir := "statements"
// 	t.Run("should run", func(t *testing.T) {
// 		snapshotTestIRGenerator(t, "", subdir)
// 	})
// }
