package codegen

import (
	"embed"
	_ "embed"
	"strings"

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
	watcode, err := filehandler.ReadAllAndCombine([]string{
		"lib/runtime.wat",
		"lib/stdlib.wat",
		"shell.wat",
	}, []embed.FS{
		runtimelib,
		stdlib,
		shell,
	})
	if err != nil {
		return err
	}
	watcode = append(watcode, target.String())
	watcode = append(watcode, ")") // close the top level module paren
	wat := strings.Join(watcode, "\n")
	if !nowasm {
		err = wat2wasm(wat, wasmpath)
		if err != nil {
			return err
		}
	}
	if preserve {
		err = filehandler.WriteToFile(watpath, []byte(wat))
		if err != nil {
			return err
		}
	}
	return nil
}

func wat2wasm(wat, wasmpath string) error {
	wasm, err := wasmtime.Wat2Wasm(wat)
	if err != nil {
		return err // TODO: update
	}
	err = filehandler.WriteToFile(wasmpath, wasm)
	return err
}
