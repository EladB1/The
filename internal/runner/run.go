package runner

import (
	"errors"
	"strings"

	"github.com/bytecodealliance/wasmtime-go"

	"github.com/EladB1/The/internal/config"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/filehandler"
)

func Run(wasm []byte, enableTraces bool, envconf *config.EnvConfig) int {
	var status int = 0
	cleanExit := false
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
		func(exitCode int32) *wasmtime.Trap {
			status = int(exitCode)
			cleanExit = true
			return wasmtime.NewTrap("") // terminate the wasmtime process silently
		},
	)
	if err != nil {
		panic(err)
	}
	wasiConf := wasmtime.NewWasiConfig()
	//wasiConf.InheritStdout()
	//wasiConf.InheritStderr()
	temps, err := filehandler.CreateTempFiles(".", ".wasmtime_stdout", ".wasmtime_stderr")
	if err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 1
	}
	tempout := temps[0]
	temperr := temps[1]
	err = wasiConf.SetStdoutFile(tempout.Name())
	if err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 1
	}
	err = wasiConf.SetStderrFile(temperr.Name())
	if err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 1
	}
	wasiConf.InheritStdin()
	wasiConf.InheritEnv()
	wasiConf.InheritArgv()
	mod, err := wasmtime.NewModule(engine, wasm)
	// mod, err := wasmtime.NewModuleFromFile(engine, file)
	if err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 1
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
		if !cleanExit {
			if trap, ok := errors.AsType[*wasmtime.Trap](err); ok {
				message := formatTrap(trap, enableTraces)
				diagnostic.ReportFatal(envconf, message, true)
				status = 1

			}
		}
	}
	if err = filehandler.CopyFromTempFiles(temps, envconf.Stdout, envconf.Stderr); err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		status = 1
	}
	return status
}

func formatTrap(trap *wasmtime.Trap, enableTraces bool) string {
	message := strings.Replace(trap.Error(), "wasm trap:", "", 1)
	message = strings.TrimLeft(message, " ")
	if !enableTraces {
		message = strings.Split(message, "\n")[0] // cut off the backtrace
	}
	return message
}
