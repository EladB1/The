package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	ds "github.com/EladB1/The/internal/datastructures"
	"github.com/EladB1/The/internal/irgen"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
	"github.com/EladB1/The/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
)

var dir string = "./testdata/fixtures"
var snapsDir string = "testdata/codegen-snaps"

type Fixture struct {
	Prog     irgen.Program
	Literals ds.LiteralPool
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
		prog, _ := irgen.Generate(ast, scopes)
		fix := Fixture{
			Prog:     prog,
			Literals: pool,
		}
		testutils.WriteResultToFile(fix, dir, fixture.File)
	}
}

func loadFixture(t *testing.T, testdir string, filename string) Fixture {
	path := filepath.Join(testdir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file %s\n%v\n", filename, err)
	}
	var fixture Fixture
	err = json.Unmarshal(content, &fixture)
	if err != nil {
		t.Errorf("Failed to unmarshal json %v\n", err)
	}
	return fixture
}

func snapshotTestCodeGenerator(t *testing.T, filename string) {
	snapshots := snaps.WithConfig(
		snaps.Dir(snapsDir),
	)
	fixture := loadFixture(t, dir, filename)
	target := Generate(filename, fixture.Prog, fixture.Literals)
	results := fmt.Sprintf("{Target:\n%v\n}", target)
	snapshots.MatchSnapshot(t, results)
}

func TestCodeGeneration(t *testing.T) {
	t.Run("should compile simple.the", func(t *testing.T) {
		snapshotTestCodeGenerator(t, "simple.json")
	})
}
