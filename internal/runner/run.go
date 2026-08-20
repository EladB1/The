package runner

import (
	"errors"
	"strings"

	"github.com/bytecodealliance/wasmtime-go"

	"github.com/EladB1/The/internal/codegen"
	"github.com/EladB1/The/internal/diagnostic"
)

func Run(target codegen.CompileTarget, outfile *string, enableTraces bool) int {
	var status int = 0
	cleanExit := false
	file := target.WasmFilepath
	if *outfile != "" {
		file = *outfile
	}
	engine := wasmtime.NewEngine()
	linker := wasmtime.NewLinker(engine)
	store := wasmtime.NewStore(engine)
	err := linker.DefineWasi()
	if err != nil {
		panic(err)
	}
	linker.AllowShadowing(true)
	err = linker.DefineFunc(
		store,
		"wasi_snapshot_preview1",
		"proc_exit",
		func(exitCode int32) {
			status = int(exitCode)
			cleanExit = true
			wasmtime.NewTrap("") // terminate the wasmtime process silently
		},
	)
	if err != nil {
		panic(err)
	}
	wasiConf := wasmtime.NewWasiConfig()
	wasiConf.InheritStdout()
	wasiConf.InheritStderr()
	wasiConf.InheritStdin()
	wasiConf.InheritEnv()
	wasiConf.InheritArgv()
	mod, err := wasmtime.NewModuleFromFile(engine, file)
	if err != nil {
		diagnostic.ReportFatal(err.Error(), 1, false)
	}

	store.SetWasi(wasiConf)
	instance, err := linker.Instantiate(store, mod)
	if err != nil {
		panic(err)
	}
	run := instance.GetFunc(store, "_start")
	if run == nil {
		panic("entry point not found")
	}
	_, err = run.Call(store)
	if err != nil {
		if cleanExit {
			return status
		} else {
			if trap, ok := errors.AsType[*wasmtime.Trap](err); ok {
				message := strings.Replace(trap.Error(), "wasm trap: ", "", 1)
				if !enableTraces {
					message = strings.Split(message, "\n")[0] // cut off the backtrace
				}
				diagnostic.ReportFatal(message, 1, true)

			}
		}
	}
	return status
}
