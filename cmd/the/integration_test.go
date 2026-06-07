//go:build integration

package main_test

/*
	Integration testing:
		1. Build up-to-date executable for compiler
		2. If no snapshot exists, fail the test
		3. Run source code through compiler and compare output with snapshot
		4. Report pass/fail

		Command: go test -tags=integration ./cmd/the

	Snapshot creation/updates:
		1. Write source code for new test file or make change to existing file (.the file)
		2. Build-up-to-date executable for compiler (using "--update" flag)
		3. Run command which will create/update the snapshot by running the executable through the compiler and saving the results (.golden file)

		Command: go test -tags=integration ./cmd/the -update="file.the"
*/

import (
	"flag"
	"testing"
)

var (
	update *string = flag.String("update", false, "create/update compiler snapshot file from a provided source code file")
)

func ValidPrograms(t *testing.T) {

}

func InvalidPrograms(t *testing.T)
