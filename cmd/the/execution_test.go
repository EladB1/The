//go:build execution

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	SOURCE_DIR    string = "testdata/execution/source"
	GENERATED_DIR string = "testdata/execution/generated"
)

type ExecutionResult struct {
	status int
	stdout string
	stderr string
}

func (result ExecutionResult) String() string {
	return fmt.Sprintf("{\n\tstatus: %d,\n\tstdout: %q,\n\tstderr: %q\n}", result.status, result.stdout, result.stderr)
}

var executionTests = []struct {
	filename       string
	expectedResult ExecutionResult
}{
	{"helloworld.the", ExecutionResult{
		status: 0,
		stdout: "hello, world!\n",
		stderr: "",
	}},
	{"nonzeroreturn.the", ExecutionResult{
		status: 1,
		stdout: "",
		stderr: "Something went wrong\n",
	}},
	{"divisionbyzero.the", ExecutionResult{
		status: 1,
		stdout: "",
		stderr: "RuntimeError: integer divide by zero \n",
	}},
	{"bounds.the", ExecutionResult{
		status: 1,
		stdout: "hlo",
		stderr: "\x1b[1;31mRuntimeError:\x1b[0m  index 6 out of range 5\n",
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
				status: result,
				stdout: stdoutBuffer.String(),
				stderr: stderrBuffer.String(),
				// Extra new line gets inserted before the output
			}
			if actualResult != sub.expectedResult {
				t.Errorf("\nExpected %v, but got %v\n", sub.expectedResult, actualResult)
			}
		})
	}
}
