package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/fatih/color"

	"github.com/EladB1/The/internal/codegen"
	"github.com/EladB1/The/internal/config"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/filehandler"
	"github.com/EladB1/The/internal/irgen"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/runner"
	"github.com/EladB1/The/internal/semantic"
)

// func init() {
// 	// Override the default help message
// 	flag.Usage = config.FlagUsageMessage
// }

func main() {
	result := RunCompiler(os.Args, os.Stdout, os.Stderr)
	os.Exit(result)
}

func RunCompiler(args []string, stdout, stderr io.Writer) int {
	buildOnly := false
	envconf := config.LoadEnvConfig(stdout, stderr)
	conf, err := config.LoadAndValidateConfig(args[1:], stderr)
	if err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 2
	}
	if conf.ColorOff {
		color.NoColor = true
	}
	if len(args) == 1 {
		conf.Flags.Usage() // show help message
		fmt.Fprintln(stderr)
		diagnostic.ReportFatal(envconf, "no input file", false)
		return 1
	}
	if len(args) == 2 {
		switch args[1] {
		case "version":
			status := 0
			if buildinfo, ok := debug.ReadBuildInfo(); ok {
				fmt.Fprintln(envconf.Stdout, buildinfo.Main.Version)
			} else {
				fmt.Fprintln(envconf.Stderr, "Unknown version")
				status = 1
			}
			return status
		case "help":
			conf.Flags.Usage()
			return 0
		}
	}
	filename := args[len(args)-1]
	if len(args) >= 3 {
		mode := args[len(args)-2]
		if mode == "run" {
			buildOnly = false
		}
	}
	src, err := filehandler.GetSourceCode(filename)
	if err != nil {
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 1
	}
	return Compile(filename, src, conf, envconf, buildOnly)
}

func Compile(filename string, source []string, conf *config.Config, envconf *config.EnvConfig, buildOnly bool) int {
	compilerDiagnostics := diagnostic.PhaseDiagnostics{}
	tokens, literals, lexerDiagnostics := lexer.Lex(source, false)
	compilerDiagnostics.Combine(lexerDiagnostics)
	if envconf.DevMode_Lexer {
		lexer.PrintTokens(envconf, tokens, literals)
	}
	if compilerDiagnostics.ExitOnError(conf, envconf) {
		return 1
	}

	ast, parserDiagnostics := parser.Parse(tokens, literals)
	compilerDiagnostics.Combine(parserDiagnostics)
	if envconf.DevMode_AST && !envconf.DevMode_AnnotatedAST {
		fmt.Fprintln(envconf.Stdout, ast.String(literals))
	}
	if compilerDiagnostics.ExitOnError(conf, envconf) {
		return 1
	}

	scopeTree, semanticDiagnostics := semantic.Analyze(&ast)
	compilerDiagnostics.Combine(semanticDiagnostics)
	if envconf.DevMode_ScopeTree {
		fmt.Fprintln(envconf.Stdout, scopeTree)
	}
	if envconf.DevMode_AnnotatedAST {
		fmt.Fprintln(envconf.Stdout, ast.String(literals))
	}
	if compilerDiagnostics.ExitOnError(conf, envconf) {
		return 1
	}

	ir, irDiagnostics := irgen.Generate(ast, scopeTree)
	compilerDiagnostics.Combine(irDiagnostics)
	if envconf.DevMode_irgen {
		fmt.Fprintln(envconf.Stdout, ir.String())
	}
	if compilerDiagnostics.ExitOnError(conf, envconf) {
		return 1
	}

	target := codegen.Generate(filename, ir, literals)
	if envconf.DevMode_codegen {
		fmt.Fprintln(envconf.Stdout, target)
	}
	errors, warnings := compilerDiagnostics.ReportStatus(conf, envconf)
	if (conf.Strict && warnings != 0) || errors != 0 {
		envconf.PrintDebugLogs()
		return 1
	}

	wasm, err := codegen.BuildExecutable(target, conf.PreserveWatFile, conf.WatFile, conf.OutFile, conf.NoWASM)
	if err != nil {
		envconf.PrintDebugLogs()
		diagnostic.ReportFatal(envconf, err.Error(), false)
		return 1
	}
	status := 0
	if !buildOnly {
		status = runner.Run(wasm, conf.EnableTraces, envconf)
	}
	envconf.PrintDebugLogs()
	return status
}
