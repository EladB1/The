//go:build execution

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	SOURCE_DIR    string = "testdata/execution/source"
	GENERATED_DIR string = "testdata/execution/generated"
)

type ExecutionResult struct {
	Status int
	Stdout string
	Stderr string
}

func (result ExecutionResult) String() string {
	return fmt.Sprintf("{\n\tstatus: %d,\n\tstdout: %q,\n\tstderr: %q\n}", result.Status, result.Stdout, result.Stderr)
}

var executionTests = []struct {
	filename       string
	expectedResult ExecutionResult
}{
	{"helloworld.the", ExecutionResult{
		Status: 0,
		Stdout: "hello, world!\n",
		Stderr: "",
	}},
	{"nonzeroreturn.the", ExecutionResult{
		Status: 1,
		Stdout: "",
		Stderr: "Something went wrong\n",
	}},
	{"divisionbyzero.the", ExecutionResult{
		Status: 1,
		Stdout: "",
		Stderr: "RuntimeError: integer divide by zero \n",
	}},
	{"bounds.the", ExecutionResult{
		Status: 1,
		Stdout: "hlo",
		Stderr: "\x1b[1;31mRuntimeError:\x1b[0m index 6 out of range 5\n",
	}},
	{"slices.the", ExecutionResult{
		Status: 1,
		Stdout: "hello, world!\nello, world!\nhell\ntrue\nl\nld!\n",
		Stderr: "\x1b[1;31mRuntimeError:\x1b[0m slice start 10 cannot be greater than slice end 1\n",
	}},
}

func TestExecution(t *testing.T) {
	if err := os.MkdirAll(GENERATED_DIR, 0755); err != nil { // create directory if it doesn't exist
		t.Errorf("Failed to create directory %s with error %v\n", GENERATED_DIR, err)
	}
	for _, sub := range executionTests {
		t.Run(sub.filename, func(t *testing.T) {
			source := filepath.Join(SOURCE_DIR, sub.filename)
			var stdoutBuffer, stderrBuffer bytes.Buffer
			args := []string{"the", "-nowasm", "run", source}
			env := []string{}
			result := RunCompiler(args, &stdoutBuffer, &stderrBuffer, env)
			actualResult := ExecutionResult{
				Status: result,
				Stdout: stdoutBuffer.String(),
				Stderr: stderrBuffer.String(),
				// Extra new line gets inserted before the output
			}
			if !cmp.Equal(actualResult, sub.expectedResult) {
				diff := cmp.Diff(actualResult, sub.expectedResult)
				t.Errorf("\nActual result(-) did not match expected result(+):\n%v\n", diff)
			}
		})
	}
}
