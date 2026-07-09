package diagnostic

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"

	ds "github.com/EladB1/The/internal/datastructures"
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
	CastError             Severity = "CastError"
	CallError             Severity = "CallError"
	ImplementationError   Severity = "ImplementationError"
	AmbiguityError        Severity = "AmbiguityError"
	ReferenceError        Severity = "ReferenceError"
)

type Diagnostic struct {
	Level    Severity
	Message  string
	Position ds.SourceLocation
}

//type PhaseDiagnostics []Diagnostic

type PhaseDiagnostics struct {
	Messages []Diagnostic
	HasError bool
}

func (diagnostics *PhaseDiagnostics) Sort() {
	sort.Slice(diagnostics.Messages, func(i, j int) bool {
		if diagnostics.Messages[i].Position.Line != diagnostics.Messages[j].Position.Line {
			return diagnostics.Messages[i].Position.Line < diagnostics.Messages[j].Position.Line
		}
		return diagnostics.Messages[i].Position.Column != diagnostics.Messages[j].Position.Column
	})
}

func (diagnostics *PhaseDiagnostics) Complain(level Severity, pos ds.SourceLocation, formatStr string, args ...any) {
	diagnostic := Diagnostic{
		Level:    level,
		Message:  fmt.Sprintf(formatStr, args...),
		Position: pos,
	}
	if strings.HasSuffix(string(level), "Error") {
		diagnostics.HasError = true
	}
	diagnostics.Messages = append(diagnostics.Messages, diagnostic)
}

func (diagnostics *PhaseDiagnostics) ComplainPositionless(level Severity, message string, args ...any) {
	pos := ds.SourceLocation{
		Line:   -1,
		Column: -1,
	}
	diagnostics.Complain(level, pos, message, args...)
}

func (diagnostics *PhaseDiagnostics) ProvideInfo(message string, args ...any) {
	diagnostics.ComplainPositionless(Info, message, args...)
}

func (diagnostics *PhaseDiagnostics) Warn(pos ds.SourceLocation, message string, args ...any) {
	diagnostics.Complain(Warning, pos, message, args...)
}

func (diagnostics *PhaseDiagnostics) WarnPositionless(message string, args ...any) {
	pos := ds.SourceLocation{
		Line:   -1,
		Column: -1,
	}
	diagnostics.Warn(pos, message, args...)
}

func (diagnostics *PhaseDiagnostics) Combine(other PhaseDiagnostics) {
	if len(diagnostics.Messages) == 0 {
		diagnostics.Messages = other.Messages
		return
	}
	diagnostics.HasError = diagnostics.HasError && other.HasError
	diagnostics.Messages = append(diagnostics.Messages, other.Messages...)
}

// Use for errors outside of source code
func ReportFatal(message string, status int) {
	fatal_err := Diagnostic{
		Level:   Error,
		Message: message,
		Position: ds.SourceLocation{
			Line:   -1,
			Column: -1,
		},
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
	if diagnostic.Position.Line != -1 && diagnostic.Position.Column != -1 {
		position = fmt.Sprintf("at line: %d, column: %d", diagnostic.Position.Line+1, diagnostic.Position.Column+1)
	}
	return fmt.Sprintf("%s: %s %s", prefix, diagnostic.Message, position)
}
