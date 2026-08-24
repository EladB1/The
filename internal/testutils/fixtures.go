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
		t.Fatalf("Failed to read directory with error: %v\n", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".the") {
			fmt.Printf("Updating fixture for '%s'...\n", file.Name())
			path := filepath.Join(dir, file.Name())
			src, err := filehandler.GetSourceCode(path)
			if err != nil {
				t.Fatalf("Failed to read source file with error: %v\n", err)
			}
			results = append(results, FixtureFile{
				Source: src,
				File:   file,
			})
		}
	}
	return results
}

func WriteResultToFile(t *testing.T, result any, dir string, source os.DirEntry) {
	output, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to mashal json with error: %v\n", err)
	}
	path := filepath.Join(dir, strings.ReplaceAll(source.Name(), ".the", ".json"))
	err = os.WriteFile(path, output, 0664)
	if err != nil {
		t.Fatalf("Failed to write file with error: %v\n", err)
	}
	fmt.Println("Fixture updated")
}
