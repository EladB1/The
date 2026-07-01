package diagnostic

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Severity string

const (
	Info                  Severity = "Info"
	Warning               Severity = "Warning"
	Error                 Severity = "Error"
	SyntaxError           Severity = "SyntaxError"
	NameError             Severity = "NameError"
	TypeError             Severity = "TypeError"
	AccessError           Severity = "AccessError"
	IllegalStatementError Severity = "IllegalStatementError"
	NamedBlockError       Severity = "NamedBlockError"
)

type Diagnostic struct {
	Level   Severity
	Message string
	Line    int
	Column  int
}

type PhaseDiagnostics []Diagnostic

func (diagnostics PhaseDiagnostics) HasError() bool {
	for _, diagnostic := range diagnostics {
		if strings.HasSuffix(string(diagnostic.Level), "Error") {
			return true
		}
	}
	return false
}

func (diagnostics PhaseDiagnostics) Complain(level Severity, message string, line int, column int) PhaseDiagnostics {
	diagnostic := Diagnostic{
		Level:   level,
		Message: message,
		Line:    line,
		Column:  column,
	}
	return append(diagnostics, diagnostic)
}

func (diagnostics PhaseDiagnostics) ComplainPositionless(level Severity, message string) PhaseDiagnostics {
	return diagnostics.Complain(level, message, -1, -1)
}

func (diagnostics PhaseDiagnostics) ProvideInfo(message string) PhaseDiagnostics {
	diagnostic := Diagnostic{
		Level:   Info,
		Message: message,
		Line:    -1,
		Column:  -1,
	}
	return append(diagnostics, diagnostic)
}

func (diagnostics PhaseDiagnostics) Warn(message string, line int, column int) PhaseDiagnostics {
	return diagnostics.Complain(Warning, message, line, column)
}

func (diagnostics PhaseDiagnostics) WarnPositionless(message string) PhaseDiagnostics {
	return diagnostics.Warn(message, -1, -1)
}

// Use for errors outside of source code
func ReportFatal(message string, status int) {
	fatal_err := Diagnostic{
		Level:   Error,
		Message: message,
		Line:    -1,
		Column:  -1,
	}
	fmt.Fprintln(os.Stderr, fatal_err)
	os.Exit(status)
}

var (
	// custom colors
	BoldRed    func(...interface{}) string = color.New(color.FgHiRed, color.Bold).SprintFunc()
	BoldYellow func(...interface{}) string = color.New(color.FgYellow, color.Bold).SprintFunc()
)

func (diagnostic Diagnostic) String() string {
	if diagnostic.Level == Info {
		return diagnostic.Message
	}
	var prefix string
	if diagnostic.Level == Warning {
		prefix = BoldYellow(diagnostic.Level)
	} else {
		prefix = BoldRed(diagnostic.Level)
	}
	var position string = ""
	if diagnostic.Line != -1 && diagnostic.Column != -1 {
		position = fmt.Sprintf("at line: %d, column: %d", diagnostic.Line+1, diagnostic.Column+1)
	}
	return fmt.Sprintf("%s: %s %s", prefix, diagnostic.Message, position)
}
