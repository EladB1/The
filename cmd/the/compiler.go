package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/EladB1/The/internal/config"
	"github.com/EladB1/The/internal/diagnostic"
	"github.com/EladB1/The/internal/lexer"
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

func getSourceCode(filename string) ([]string, error) {
	if !strings.HasSuffix(filename, ".the") {
		return nil, fmt.Errorf("only '.the' files accepted")
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(contents), "\n"), nil
}

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
	tokens, lexerDiagnostics := lexer.Lex(source)
	compilerDiagnostics = append(compilerDiagnostics, lexerDiagnostics...)
	if lexerDiagnostics.HasError() {
		reportStatus(compilerDiagnostics)
		os.Exit(1)
	}
	//fmt.Println(tokens)
	lexer.PrintTokens(tokens)
	errors, warnings := reportStatus(compilerDiagnostics)
	if (conf.Strict && warnings != 0) || errors != 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

func reportStatus(messages diagnostic.PhaseDiagnostics) (int, int) {
	var warningCnt int = 0
	var errorCnt int = 0
	for _, message := range messages {
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
			errorCnt++
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
		diagnostic.ReportFatalStringPositionless(diagnostic.Error, "Cannot use strict and suppress-warnings flags together", 2)
	}
	if *colorOff {
		color.NoColor = true
	}
	args := os.Args
	if len(args) == 1 {
		flag.Usage() // show help message
		fmt.Fprintln(os.Stderr)
		diagnostic.ReportFatalStringPositionless(diagnostic.Error, "no input file", 1)
		os.Exit(1)
	}
	filename := os.Args[len(args)-1]
	src, err := getSourceCode(filename)
	if err != nil {
		diagnostic.ReportFatalPositionless(diagnostic.Error, err, 1)
	}
	compile(src)
}
