//go:build integration

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

const (
	SNAPS_DIR string = "testdata/integration"
	FIX_DIR   string = "testdata/integration/fixtures"
)

var (
	executable []string = []string{"the"}
)

func snapshotTestCompilerWithArgs(t *testing.T, snapshots *snaps.Config, args ...string) {
	combined := &bytes.Buffer{}
	args = append(executable, args...) // prepend the program so it simulates the CLI
	env := []string{"THE_DEV_CODEGEN=true"}
	exitCode := RunCompiler(args, combined, combined, env)
	out := combined.String()
	re := regexp.MustCompile(`/.*/the.test`)
	out = re.ReplaceAllString(out, "the")
	results := fmt.Sprintf("Exit code: %d\n===\n\nOutput:\n\n%s", exitCode, out)
	snapshots.MatchSnapshot(t, results)

}

func TestCommandLineArgs(t *testing.T) {
	snapshots := snaps.WithConfig(
		snaps.Dir(SNAPS_DIR),
		snaps.Filename("cli"),
	)
	t.Run("should fail when no arguments or files provided", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots)
	})
	t.Run("should fail when given improper file extension", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "file.txt")
	})
	t.Run("should fail when conflicting flags provided", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "-strict", "-suppress-warnings", "examples/src/loops.the")
	})
	t.Run("should fail when file does not exist", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "something.the")
	})
	t.Run("should fail when invalid WAT path provided", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "-preserve-wat-file", "-wat", "something.txt", "build", "something.the")
	})
	t.Run("should ignore invalid WAT path when no preserve flag", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "-wat", "something.txt", "build", "something.the")
	})
	t.Run("should fail when invalid WASM path provided", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "-o", "out.txt", "build", "something.the")
	})
	t.Run("should ignore invalid WASM when nowasm flag present", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "-nowasm", "-o", "out.txt", "build", "something.the")
	})
	t.Run("should pass and show help message on help flag", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "help")
	})
}

func TestValidPrograms(t *testing.T) {
	snapshots := snaps.WithConfig(
		snaps.Dir(SNAPS_DIR),
		snaps.Filename("valid"),
	)
	path := filepath.Join(FIX_DIR, "valid")
	t.Run("should compile program.the with no issues", func(t *testing.T) {
		os.Setenv("THE_DEV_CODEGEN", "true")
		snapshotTestCompilerWithArgs(t, snapshots, "-nowasm", "build", filepath.Join(path, "program.the"))
		os.Setenv("THE_DEV_CODEGEN", "")
	})
}

func TestInvalidPrograms(t *testing.T) {
	snapshots := snaps.WithConfig(
		snaps.Dir(SNAPS_DIR),
		snaps.Filename("invalid"),
	)
	path := filepath.Join(FIX_DIR, "invalid")
	t.Run("should try to compile lexer_errors.the and report lexer errors", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "build", filepath.Join(path, "lexer_errors.the"))
	})
	t.Run("should try to compile parser_errors.the and report parser errors", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "build", filepath.Join(path, "parser_errors.the"))
	})
	t.Run("should try to compile semantic_errors.the and report semantic errors", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "build", filepath.Join(path, "semantic_errors.the"))
	})
}
