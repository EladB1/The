package diagnostic

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Severity string

const (
	Warning     Severity = "Warning"
	Error       Severity = "Error"
	SyntaxError Severity = "SyntaxError"
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

var (
	// custom colors
	BoldRed    func(...interface{}) string = color.New(color.FgHiRed, color.Bold).SprintFunc()
	BoldYellow func(...interface{}) string = color.New(color.FgYellow, color.Bold).SprintFunc()
)

func (diagnostic Diagnostic) String() string {
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

func ReportFatalPositionless(level Severity, err error, status int) {
	ReportFatal(level, err, -1, -1, status)

}

func ReportFatalStringPositionless(level Severity, message string, status int) {
	ReportFatalString(level, message, -1, -1, status)
}

func ReportFatal(level Severity, err error, line int, col int, status int) {
	fmt.Fprintln(os.Stderr, Complain(level, err.Error(), line, col))
	os.Exit(status)
}

func ReportFatalString(level Severity, message string, line int, col int, status int) {
	fmt.Fprintln(os.Stderr, Complain(level, message, line, col))
	os.Exit(status)
}

func Complain(level Severity, message string, line int, col int) Diagnostic {
	return Diagnostic{
		Level:   level,
		Message: message,
		Line:    line,
		Column:  col,
	}
}

func ComplainPositionless(level Severity, message string) Diagnostic {
	return Complain(level, message, -1, -1)
}

func Warn(message string, line int, col int) Diagnostic {
	return Diagnostic{
		Level:   Warning,
		Message: message,
		Line:    line,
		Column:  col,
	}
}

func WarnPositionless(message string) Diagnostic {
	return Warn(message, -1, -1)
}
