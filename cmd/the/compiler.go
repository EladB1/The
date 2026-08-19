package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/bytecodealliance/wasmtime-go"
	"github.com/fatih/color"

	"github.com/EladB1/The/internal/codegen"
	"github.com/EladB1/The/internal/config"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/filehandler"
	"github.com/EladB1/The/internal/irgen"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
)

var (
	logBuffer bytes.Buffer
	// cli flags
	colorOff         *bool   = flag.Bool("no-color", false, "Disable color output")
	suppressWarnings *bool   = flag.Bool("suppress-warnings", false, "Disable reporting of warnings")
	strict           *bool   = flag.Bool("strict", false, "Any warnings will cause compilation to fail")
	preserveWatFile  *bool   = flag.Bool("preserve-wat-file", false, "Prevents the compiler from deleting the generated WAT file")
	watfile          *string = flag.String("wat", "", "Path to the generated wat file")
	outfile          *string = flag.String("o", "", "Path to the generated wasm executable")
	nowasm           *bool   = flag.Bool("nowasm", false, "Only produce a WAT file instead of compiling it down to WASM. Cannot be used with run command.")
	// env flags used to show output from different parts of compiler
	devMode_lexer    bool = false
	devMode_parser   bool = false
	devMode_semantic bool = false
	devMode_irgen    bool = false
	devMode_codegen  bool = false
	compilerDebug    bool = false

	// internal configurations
	conf config.Config = config.Config{
		Debug:            &compilerDebug,
		Strict:           *strict,
		SuppressWarnings: *suppressWarnings,
		LogBuffer:        &logBuffer,
	}
	compilerDiagnostics diagnostic.PhaseDiagnostics
	buildOnly           bool = true
)

func init() {
	log.SetOutput(&logBuffer)
	// Override the default help message
	flag.Usage = func() {
		output := flag.CommandLine.Output()

		fmt.Fprintf(output, "Usage: %s version\n", os.Args[0])
		fmt.Fprintf(output, "Usage: %s [options] [run | build] [file]\n", os.Args[0])
		fmt.Fprintln(output, "options:")
		flag.PrintDefaults()
	}
}

func checkEnvironment() {
	vars := map[string]*bool{
		"THE_DEV_LEXER":    &devMode_lexer,
		"THE_DEV_PARSER":   &devMode_parser,
		"THE_DEV_SEMANTIC": &devMode_semantic,
		"THE_DEV_IRGEN":    &devMode_irgen,
		"THE_DEV_CODEGEN":  &devMode_codegen,
		"THE_DEV_DEBUG":    &compilerDebug,
	}
	for key, flag := range vars {
		value, err := fetchEnvironmentVariableValue(key)
		if err != nil {
			*flag = false
		} else {
			*flag = value
		}
	}
}

func fetchEnvironmentVariableValue(key string) (bool, error) {
	envVar := os.Getenv(key)
	return strconv.ParseBool(envVar)
}

func main() {
	flag.Parse()
	checkEnvironment()
	if *strict && *suppressWarnings {
		diagnostic.ReportFatal("Cannot use strict and suppress-warnings flags together", 2)
	}
	if *colorOff {
		color.NoColor = true
	}
	args := os.Args
	if len(args) == 1 {
		flag.Usage() // show help message
		fmt.Fprintln(os.Stderr)
		diagnostic.ReportFatal("no input file", 1)
	}
	if len(args) == 2 && args[1] == "version" {
		if buildinfo, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(buildinfo.Main.Version)
		} else {
			fmt.Fprintln(os.Stderr, "Unknown version")
			os.Exit(1)
		}
		return
	}
	filename := args[len(args)-1]
	if len(args) >= 3 {
		switch args[len(args)-2] {
		case "run":
			if *nowasm {
				diagnostic.ReportFatal("Cannot use -nowasm with run command", 1)
			}
			buildOnly = false
		case "build":
			buildOnly = true
		}
	}
	src, err := filehandler.GetSourceCode(filename)
	if err != nil {
		diagnostic.ReportFatal(err.Error(), 1)
	}
	compile(filename, src)
}

func compile(filename string, source []string) {
	tokens, literals, lexerDiagnostics := lexer.Lex(source, false)
	compilerDiagnostics.Combine(lexerDiagnostics)
	if devMode_lexer {
		lexer.PrintTokens(tokens, literals)
	}
	lexerDiagnostics.ExitOnError(conf)

	ast, parserDiagnostics := parser.Parse(tokens, literals)
	compilerDiagnostics.Combine(parserDiagnostics)
	if devMode_parser && !devMode_semantic {
		fmt.Println(ast.String(literals))
	}
	parserDiagnostics.ExitOnError(conf)

	scopeTree, semanticDiagnostics := semantic.Analyze(&ast)
	compilerDiagnostics.Combine(semanticDiagnostics)
	if devMode_semantic {
		fmt.Println(scopeTree)
		fmt.Println(ast.String(literals))
	}
	semanticDiagnostics.ExitOnError(conf)

	ir, irDiagnostics := irgen.Generate(ast, scopeTree)
	compilerDiagnostics.Combine(irDiagnostics)
	if devMode_irgen {
		fmt.Println(ir.String())
	}
	irDiagnostics.ExitOnError(conf)

	target := codegen.Generate(filename, ir, literals)
	if devMode_codegen {
		fmt.Println(target)
	}
	errors, warnings := compilerDiagnostics.ReportStatus(conf)
	if (conf.Strict && warnings != 0) || errors != 0 {
		conf.PrintDebugLogs()
		os.Exit(1)
	}

	err := codegen.BuildExecutable(target, *preserveWatFile, *watfile, *outfile, *nowasm)
	if err != nil {
		conf.PrintDebugLogs()
		diagnostic.ReportFatal(err.Error(), 1)
	}
	status := 0
	if !buildOnly {
		status = run(target, outfile)
	}
	conf.PrintDebugLogs()
	os.Exit(status)
}

func run(target codegen.CompileTarget, outfile *string) int {
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
		panic(err)
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
			panic(err)
		}
	}
	return status
}
