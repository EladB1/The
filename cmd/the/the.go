package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	// compiler message collections
	warnings []string
	errors   []error

	// custom colors
	boldRed    func(...interface{}) string = color.New(color.FgHiRed, color.Bold).SprintFunc()
	boldYellow func(...interface{}) string = color.New(color.FgYellow, color.Bold).SprintFunc()

	// cli flags
	colorOff         *bool
	suppressWarnings *bool
	strict           *bool
)

func colorPrefix(colorFormatter func(...interface{}) string, prefix string, suffix string) string {
	return fmt.Sprintf("%s, %s", colorFormatter(prefix), suffix)
}

func Complain(message string) error {
	err := fmt.Errorf(colorPrefix(boldRed, "Error:", message))
	errors = append(errors, err)
	return err
}

func Warn(warning string) {
	if *suppressWarnings {
		return
	}
	warnStr := colorPrefix(boldYellow, "Warning:", warning)
	warnings = append(warnings, warnStr)
}

func ReportFatal(err error /*, status int*/) {
	fmt.Println(err)
	//os.Exit(status)
}

func ReportFatalString(message string /*, status int*/) {
	err := Complain(message)
	ReportFatal(err /*, status*/)
}

func reportStatus() {
	if !*suppressWarnings {
		for _, warning := range warnings {
			fmt.Println(warning)
		}
	}
	for _, err := range errors {
		fmt.Println(err)
	}
	errorsLen := len(errors)
	errorCount := colorPrefix(boldRed, "Errors", strconv.Itoa(errorsLen))
	if *suppressWarnings {
		if errorsLen != 0 {
			fmt.Println(errorCount)
		}
	} else {
		warningsLen := len(warnings)
		if warningsLen != 0 || errorsLen != 0 {
			warningCount := colorPrefix(boldYellow, "Warnings", strconv.Itoa(warningsLen))
			fmt.Printf("%s, %s\n", warningCount, errorCount)
		}
	}
}

func getSourceCode(filename string) ([]string, error) {
	if !strings.HasSuffix(filename, ".the") {
		return nil, Complain("only '.the' files accepted")
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, Complain(err.Error())
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
	colorOff = flag.Bool("no-color", false, "Disable color output")
	suppressWarnings = flag.Bool("suppress-warnings", false, "Disable reporting of warnings")
	strict = flag.Bool("strict", false, "Any warnings will cause compilation to fail")
}

func main() {
	flag.Parse()
	if *colorOff {
		color.NoColor = true
	}
	fatal := false
	args := os.Args
	if len(args) == 1 {
		flag.Usage() // show help message
		fmt.Println()
		ReportFatalString("no input file" /*, 1*/)
		os.Exit(1)
	}
	filename := os.Args[len(args)-1]
	src, err := getSourceCode(filename)
	if err != nil {
		ReportFatal(err /*, 1*/)
		fatal = true
	}
	if fatal {
		os.Exit(1)
	}
	fmt.Println(src) // TODO: replace with compiler pipeline steps (lexer, parser, etc.)
	reportStatus()
}
