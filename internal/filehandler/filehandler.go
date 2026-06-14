package filehandler

import (
	"fmt"
	"os"
	"strings"
)

func GetSourceCode(filename string) ([]string, error) {
	if !strings.HasSuffix(filename, ".the") {
		return nil, fmt.Errorf("only '.the' files accepted")
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(contents), "\n"), nil
}
