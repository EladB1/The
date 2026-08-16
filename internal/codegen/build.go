package codegen

import (
	"embed"
	_ "embed"
	"os"
	"os/exec"

	"github.com/EladB1/The/internal/filehandler"
)

//go:embed lib/runtime.wat
var runtimelib embed.FS

//go:embed lib/stdlib.wat
var stdlib embed.FS

//go:embed shell.wat
var shell embed.FS

func BuildExecutable(target CompileTarget, preserve bool, watfile string, outfile string) error {
	var wasmpath string
	var watpath string
	if watfile != "" {
		watpath = watfile
	} else {
		watpath = target.WatFilepath
	}
	if outfile != "" {
		wasmpath = outfile
	} else {
		wasmpath = target.WasmFilepath
	}
	err := filehandler.CopyFile(runtimelib, "lib/runtime.wat", watpath)
	if err != nil {
		return err
	}
	if err = filehandler.CombineFiles(watpath, "lib/stdlib.wat", stdlib); err != nil {
		return err
	}
	if err = filehandler.CombineFiles(watpath, "shell.wat", shell); err != nil {
		return err
	}
	if err = filehandler.AppendToFile(watpath, target.String()); err != nil {
		return err
	}
	if err = filehandler.AppendToFile(watpath, ")"); err != nil { // append the closing paren
		return err
	}
	cmd := exec.Command("wat2wasm", watpath, "-o", wasmpath)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if !preserve {
		err = os.Remove(watpath)
	}
	return err
}
