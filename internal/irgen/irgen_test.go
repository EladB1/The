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

type ScopeMap map[string]*semantic.Scope

func (registry ScopeMap) get(id string) (*semantic.Scope, error) {
	if scope, ok := registry[id]; ok {
		return scope, nil
	} else {
		return nil, fmt.Errorf("Could not find %s in ScopeMap", id)
	}
}

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") != "true" {
		t.Skip()
	}
	fixtures := testutils.GetSourceFromDirectory(t, dir)
	for _, fixture := range fixtures {
		tokens, pool, _ := lexer.Lex(fixture.Source, false)
		ast, _ := parser.Parse(tokens, pool)
		scopes, _ := semantic.Analyze(&ast)
		fix := Fixture{
			ScopeTree: scopes,
			AST:       ast,
			Literals:  pool,
		}
		testutils.WriteResultToFile(fix, dir, fixture.File)
	}

}

func loadFixture(t *testing.T, testdir string, filename string) Fixture {
	path := filepath.Join(testdir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file %s\n%v", filename, err)
	}
	var fixture Fixture
	err = json.Unmarshal(content, &fixture)
	if err != nil {
		t.Errorf("Failed to unmarshal json %v\n", err)
	}
	return fixture
}

func repairScopeTree(scopeTree *semantic.Scope) error {
	smap := ScopeMap{}
	repairScope(scopeTree, nil, smap)
	err := repairSymbols(scopeTree, smap)
	return err
}

// after deserialization need to be point the child scopes back at their parents
func repairScope(scopeTree *semantic.Scope, parent *semantic.Scope, smap ScopeMap) {
	smap[scopeTree.Id] = scopeTree
	scopeTree.Parent = parent
	for _, child := range scopeTree.Children {
		repairScope(child, scopeTree, smap)
	}
}

// after deserialization need to be point the inner scopes back at scope tree nodes
func repairSymbols(scopeTree *semantic.Scope, smap ScopeMap) error {
	var err error
	// structs
	for str := range scopeTree.Structs.All() {
		if str.InnerScope != nil {
			str.InnerScope, err = smap.get(str.InnerScope.Id)
			if err != nil {
				return err
			}
			scopeTree.Structs.Update(str, str.Name)
		}
	}
	// named blocks
	for nb := range scopeTree.NamedBlocks.All() {
		if nb.InnerScope != nil {
			nb.InnerScope, err = smap.get(nb.InnerScope.Id)
			if err != nil {
				return err
			}
			scopeTree.NamedBlocks.Update(nb, nb.Name)
		}
	}
	// functions
	for fn := range scopeTree.Functions.All() {
		for i, overload := range fn.Overloads {
			if overload.InnerScope != nil {
				overload.InnerScope, err = smap.get(overload.InnerScope.Id)
				if err != nil {
					return err
				}
				fn.Overloads[i] = overload
			}
		}
		scopeTree.Functions.Update(fn, fn.Name)
	}
	for _, child := range scopeTree.Children {
		if err := repairSymbols(child, smap); err != nil {
			return err
		}
	}
	return nil
}

func snapshotTestIRGenerator(t *testing.T, filename string) {
	snapshots := snaps.WithConfig(
		snaps.Dir(snapsDir),
	)
	fixture := loadFixture(t, dir, filename)
	err := repairScopeTree(fixture.ScopeTree)
	if err != nil {
		t.Fatalf("Failed to re-build scopes from '%s' with error: %v", filename, err)
	}
	prog, messages := Generate(fixture.AST, fixture.ScopeTree)
	var msgBuilder strings.Builder
	delim := ","
	for i, msg := range messages.Messages {
		if i == len(messages.Messages)-1 {
			delim = ""
		}
		msgBuilder.WriteString(fmt.Sprintf("\n\t\"%v\"%s", msg, delim))
	}
	results := fmt.Sprintf("IR:\n%v\nCompiler messages:\n[%s\n]", prog.String(), msgBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

func TestIRGeneration(t *testing.T) {
	t.Run("should run expressions and produce IR", func(t *testing.T) {
		snapshotTestIRGenerator(t, "expressions.json")
	})

	t.Run("should run functions and produce IR", func(t *testing.T) {
		snapshotTestIRGenerator(t, "functions.json")
	})

	t.Run("should run ifBlocks and produce IR", func(t *testing.T) {
		snapshotTestIRGenerator(t, "ifBlocks.json")
	})

	t.Run("should run loops and produce IR", func(t *testing.T) {
		snapshotTestIRGenerator(t, "loops.json")
	})

	t.Run("should run structs and produce IR", func(t *testing.T) {
		snapshotTestIRGenerator(t, "structs.json")
	})
}
