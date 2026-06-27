//go:build integration

package main_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func snapshotTestCompilerWithArgs(t *testing.T, snapshots *snaps.Config, args ...string) {
	targetBinary := "../../the"
	cmd := exec.Command(targetBinary, args...)
	output, _ := cmd.CombinedOutput() // ignore errors since we'll be expecting errors from the compiler for some tests
	out := string(output)
	exitCode := cmd.ProcessState.ExitCode()
	out = strings.ReplaceAll(out, fmt.Sprintf("exit status %d", exitCode), "") // remove stderr line inserted by cmd.CombinedOutput
	results := fmt.Sprintf("Exit code: %d\n===\n\nOutput:\n\n%s", exitCode, out)
	snapshots.MatchSnapshot(t, results)

}

func TestNoCommandLineArgs(t *testing.T) {
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/"),
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
	t.Run("should pass and show help message on help flag", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "-h")
	})
}

func TestValidPrograms(t *testing.T) {
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/"),
		snaps.Filename("valid"),
	)
	t.Run("should compile program.the with no issues", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "testdata/fixtures/valid/program.the")
	})
}

func TestInvalidPrograms(t *testing.T) {
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/"),
		snaps.Filename("invalid"),
	)
	t.Run("should try to compile lexer_errors.the and report lexer errors", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "testdata/fixtures/invalid/lexer_errors.the")
	})
	t.Run("should try to compile parser_errors.the and report parser errors", func(t *testing.T) {
		snapshotTestCompilerWithArgs(t, snapshots, "testdata/fixtures/invalid/parser_errors.the")
	})
}
