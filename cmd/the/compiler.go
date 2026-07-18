package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/EladB1/The/internal/config"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/filehandler"
	"github.com/EladB1/The/internal/irgen"
	"github.com/EladB1/The/internal/lexer"
	"github.com/EladB1/The/internal/parser"
	"github.com/EladB1/The/internal/semantic"
	"github.com/fatih/color"
)

var (
	// cli flags
	colorOff         *bool         = flag.Bool("no-color", false, "Disable color output")
	suppressWarnings *bool         = flag.Bool("suppress-warnings", false, "Disable reporting of warnings")
	strict           *bool         = flag.Bool("strict", false, "Any warnings will cause compilation to fail")
	conf             config.Config = config.Config{
		Strict:           *strict,
		SuppressWarnings: *suppressWarnings,
	}
	compilerDiagnostics diagnostic.PhaseDiagnostics
)

func init() {
	// Override the default help message
	flag.Usage = func() {
		output := flag.CommandLine.Output()

		fmt.Fprintf(output, "Usage: %s [options] [file]\n", os.Args[0])
		fmt.Fprintln(output, "options:")
		flag.PrintDefaults()
	}
}

func compile(source []string) {
	tokens, literals, lexerDiagnostics := lexer.Lex(source, false)
	compilerDiagnostics.Combine(lexerDiagnostics)
	//lexer.PrintTokens(tokens, literals)
	if lexerDiagnostics.HasError {
		reportStatus(compilerDiagnostics)
		os.Exit(1)
	}
	ast, parserDiagnostics := parser.Parse(tokens, literals)
	compilerDiagnostics.Combine(parserDiagnostics)
	// fmt.Println(ast.String(literals))
	if parserDiagnostics.HasError {
		reportStatus(compilerDiagnostics)
		os.Exit(1)
	}
	scopeTree, semanticDiagnostics := semantic.Analyze(&ast)
	fmt.Println(scopeTree)
	compilerDiagnostics.Combine(semanticDiagnostics)
	fmt.Println(ast.String(literals))
	if semanticDiagnostics.HasError {
		reportStatus(compilerDiagnostics)
		os.Exit(1)
	}
	ir, irDiagnostics := irgen.Generate(ast, scopeTree)
	compilerDiagnostics.Combine(irDiagnostics)
	fmt.Println(ir.String())
	if irDiagnostics.HasError {
		reportStatus(compilerDiagnostics)
		os.Exit(1)
	}
	errors, warnings := reportStatus(compilerDiagnostics)
	if (conf.Strict && warnings != 0) || errors != 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func reportStatus(messages diagnostic.PhaseDiagnostics) (int, int) {
	var warningCnt int = 0
	var errorCnt int = 0
	for _, message := range messages.Messages {
		if message.Level == diagnostic.Warning {
			if conf.SuppressWarnings {
				continue
			}
			warningCnt++
			if conf.Strict {
				fmt.Fprintln(os.Stderr, message)
			} else {
				fmt.Println(message)
			}
		} else {
			if message.Level != diagnostic.Info {
				errorCnt++
			}
			fmt.Fprintln(os.Stderr, message)
		}
	}
	var summary string = ""
	if warningCnt != 0 || errorCnt != 0 {
		if conf.SuppressWarnings {
			summary = fmt.Sprintf("\n%s:\n%s: %d", color.HiBlueString("Summary"), diagnostic.BoldRed("Errors"), errorCnt)
		}
		summary = fmt.Sprintf("\n%s:\n%s: %d, %s: %d", color.HiBlueString("Summary"), diagnostic.BoldRed("Errors"), errorCnt, diagnostic.BoldYellow("Warnings"), warningCnt)
	}
	fmt.Println(summary)
	return errorCnt, warningCnt
}

func main() {
	flag.Parse()
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
		os.Exit(1)
	}
	filename := os.Args[len(args)-1]
	src, err := filehandler.GetSourceCode(filename)
	if err != nil {
		diagnostic.ReportFatal(err.Error(), 1)
	}
	compile(src)
}
