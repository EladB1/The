package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"runtime/debug"

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
	// cli flags
	colorOff         *bool   = flag.Bool("no-color", false, "Disable color output")
	suppressWarnings *bool   = flag.Bool("suppress-warnings", false, "Disable reporting of warnings")
	strict           *bool   = flag.Bool("strict", false, "Any warnings will cause compilation to fail")
	preserveWatFile  *bool   = flag.Bool("preserve-wat-file", false, "Prevents the compiler from deleting the generated WAT file")
	watfile          *string = flag.String("wat", "", "Path to the generated wat file")
	outfile          *string = flag.String("o", "", "Path to the generated wasm executable")

	// env flags used to show output from different parts of compiler
	devMode_lexer    bool = false
	devMode_parser   bool = false
	devMode_semantic bool = false
	devMode_irgen    bool = false
	devMode_codegen  bool = false

	// internal configurations
	conf config.Config = config.Config{
		Strict:           *strict,
		SuppressWarnings: *suppressWarnings,
	}
	compilerDiagnostics diagnostic.PhaseDiagnostics
	buildOnly           bool = true
)

func init() {
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
	err := codegen.BuildExecutable(target, *preserveWatFile, *watfile, *outfile)
	if err != nil {
		diagnostic.ReportFatal(err.Error(), 1)
	}
	status := 0
	if !buildOnly {
		status = run(target, outfile)
	}
	errors, warnings := compilerDiagnostics.ReportStatus(conf)
	if (conf.Strict && warnings != 0) || errors != 0 {
		os.Exit(1)
	}
	os.Exit(status)
}

func run(target codegen.CompileTarget, outfile *string) int {
	file := target.WasmFilepath
	if *outfile != "" {
		file = *outfile
	}
	cmd := exec.Command("wasmtime", file)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return cmd.ProcessState.ExitCode()
	}
	return 0
}
