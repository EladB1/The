package codegen

import (
	"embed"
	_ "embed"
	"os"

	"github.com/bytecodealliance/wasmtime-go"

	"github.com/EladB1/The/internal/filehandler"
)

//go:embed lib/runtime.wat
var runtimelib embed.FS

//go:embed lib/stdlib.wat
var stdlib embed.FS

//go:embed shell.wat
var shell embed.FS

func BuildExecutable(target CompileTarget, preserve bool, watfile string, outfile string, nowasm bool) error {
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
	var wasm_err error
	if !nowasm {
		wasm_err = wat2wasm(watpath, wasmpath)
	}
	if !preserve {
		err = os.Remove(watpath)
		if err != nil {
			return err
		}
	}
	if wasm_err != nil { // defer this for file cleanup
		return wasm_err
	}
	return nil
}

func wat2wasm(watpath, wasmpath string) error {
	wat, err := os.ReadFile(watpath)
	if err != nil {
		return err
	}
	wasm, err := wasmtime.Wat2Wasm(string(wat))
	if err != nil {
		return err // TODO: update
	}
	err = filehandler.WriteWasmToFile(wasmpath, wasm)
	return err
}
