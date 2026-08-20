package filehandler

import (
	"embed"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

const MAX_SIZE int64 = 1024 * 1024 // 1 MB

func GetSourceCode(filename string) ([]string, error) {
	if !strings.HasSuffix(filename, ".the") {
		return nil, fmt.Errorf("only '.the' files accepted")
	}
	actualSize, err := getFileSize(filename)
	if err != nil {
		return nil, err
	}
	if actualSize > MAX_SIZE {
		return nil, fmt.Errorf("File size %s exceeds limit of %s", getHumanReadableFileSize(actualSize), getHumanReadableFileSize(MAX_SIZE))
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(contents), "\n"), nil
}

func getFileSize(filename string) (int64, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func getHumanReadableFileSize(size int64) string {
	var base float64 = 1024
	suffixes := []string{"B", "KB", "MB", "GB", "TB"}
	index := math.Floor(math.Log(float64(size)) / math.Log(base))
	if int(index) >= len(suffixes) {
		return "INF"
	}
	value := float64(size) / math.Pow(base, index)
	return fmt.Sprintf("%.2f %s", value, suffixes[int(index)])
}

func ReadAllAndCombine(names []string, files []embed.FS) ([]string, error) {
	combined := []string{}
	for i := range len(names) {
		content, err := files[i].ReadFile(names[i])
		if err != nil {
			return combined, err
		}
		combined = append(combined, string(content))
	}
	return combined, nil
}

func ClearFileIfExists(filename string) error {
	_, err := os.Stat(filename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return os.Truncate(filename, 0)
}

func WriteWatToFile(filename string, content string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	err = file.Sync()
	return err
}

func WriteWasmToFile(filename string, wasm []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(wasm)
	if err != nil {
		return err
	}
	err = file.Sync()
	return err
}
