//go:build execution

package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	SOURCE_DIR    string = "testdata/execution/source"
	GENERATED_DIR string = "testdata/execution/generated"
	targetBinary  string = "../../the"
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
}

func TestExecution(t *testing.T) {
	if err := os.MkdirAll(GENERATED_DIR, 0755); err != nil { // create directory if it doesn't exist
		t.Errorf("Failed to create directory %s with error %v\n", GENERATED_DIR, err)
	}
	for _, sub := range executionTests {
		t.Run(sub.filename, func(t *testing.T) {
			source := filepath.Join(SOURCE_DIR, sub.filename)
			wasm := filepath.Join(GENERATED_DIR, strings.Replace(sub.filename, ".the", ".wasm", 1))
			cmd := exec.Command(targetBinary, "-o", wasm, "run", source)
			cmd.Env = []string{}
			var stdoutBuffer, stderrBuffer bytes.Buffer
			cmd.Stdout = &stdoutBuffer
			cmd.Stderr = &stderrBuffer
			_ = cmd.Run()
			actualResult := ExecutionResult{
				status: cmd.ProcessState.ExitCode(),
				stdout: strings.TrimPrefix(stdoutBuffer.String(), "\n"),
				stderr: strings.TrimPrefix(stderrBuffer.String(), "\n"),
				// Extra new line gets inserted before the output
			}
			if actualResult != sub.expectedResult {
				t.Errorf("\nExpected %v, but got %v\n", actualResult, sub.expectedResult)
			}
		})
	}
}
