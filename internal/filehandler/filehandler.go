package filehandler

import (
	"embed"
	"fmt"
	"io"
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

func WriteToFile(filename string, wasm []byte) error {
	return os.WriteFile(filename, wasm, 0600)
}

func CreateTempFiles(dir string, names ...string) ([]*os.File, error) {
	tempFiles := []*os.File{}
	for _, name := range names {
		if temp, err := os.CreateTemp(dir, name); err != nil {
			return nil, err
		} else {
			tempFiles = append(tempFiles, temp)
		}
	}
	return tempFiles, nil
}

func CopyFromTempFiles(temps []*os.File, destFiles ...io.Writer) error {
	defer cleanTempFiles(temps)
	for i := range temps {
		if _, err := io.Copy(destFiles[i], temps[i]); err != nil {
			return err
		}
	}
	return nil
}

func cleanTempFiles(temps []*os.File) error {
	for _, temp := range temps {
		if err := os.Remove(temp.Name()); err != nil {
			return err
		}
	}
	return nil
}
