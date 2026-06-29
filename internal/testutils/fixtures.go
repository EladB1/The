package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EladB1/The/internal/filehandler"
)

type FixtureFile struct {
	File   os.DirEntry
	Source []string
}

func GetSourceFromDirectory(t *testing.T, dir string) []FixtureFile {
	var results []FixtureFile
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read directory\n", err)
		os.Exit(1)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".the") {
			fmt.Printf("Updating fixture for '%s'... ", file.Name())
			path := filepath.Join(dir, file.Name())
			src, err := filehandler.GetSourceCode(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to read source file\n", err)
				os.Exit(1)
			}
			results = append(results, FixtureFile{
				Source: src,
				File:   file,
			})
		}
	}
	return results
}

func WriteResultToFile(result any, dir string, source os.DirEntry) {
	output, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to mashal json\n", err)
		os.Exit(1)
	}
	path := filepath.Join(dir, strings.ReplaceAll(source.Name(), ".the", ".json"))
	err = os.WriteFile(path, output, 0664)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write file\n", err)
		os.Exit(1)
	}
	fmt.Println("Fixture updated")
}
