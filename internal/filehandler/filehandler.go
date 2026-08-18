package filehandler

import (
	"embed"
	"fmt"
	"io"
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

func CopyFile(sourceFile embed.FS, sourceName string, destFile string) error {
	source, err := sourceFile.Open(sourceName)
	if err != nil {
		return fmt.Errorf("Failed to open source file %v", err)
	}
	defer source.Close()
	dest, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("Failed to create file: %v", err)
	}
	defer dest.Close()
	_, err = io.Copy(dest, source)
	if err != nil {
		return fmt.Errorf("Failed to copy file: %v", err)
	}
	err = dest.Sync() // write to disk
	if err != nil {
		return err
	}
	return nil
}

func CombineFiles(filename string, inName string, inFile embed.FS) error {
	content, err := inFile.ReadFile(inName)
	if err != nil {
		return err
	}
	return AppendToFile(filename, string(content))
}

func AppendToFile(filename string, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func WriteWasmToFile(filename string, wasm []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(wasm)
	return err
}
