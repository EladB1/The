package codegen

import (
	"embed"
	_ "embed"

	"github.com/EladB1/The/internal/filehandler"
)

//go:embed lib/runtime.wat
var runtimelib embed.FS

//go:embed lib/stdlib.wat
var stdlib embed.FS

//go:embed shell.wat
var shell embed.FS

func BuildExecutable(target CompileTarget) error {
	filepath := target.Filepath
	err := filehandler.CopyFile(runtimelib, "lib/runtime.wat", filepath)
	if err != nil {
		return err
	}
	if err = filehandler.CombineFiles(filepath, "lib/stdlib.wat", stdlib); err != nil {
		return err
	}
	if err = filehandler.CombineFiles(filepath, "shell.wat", shell); err != nil {
		return err
	}
	if err = filehandler.AppendToFile(filepath, target.String()); err != nil {
		return err
	}
	err = filehandler.AppendToFile(filepath, ")") // append the closing paren
	return err
}
