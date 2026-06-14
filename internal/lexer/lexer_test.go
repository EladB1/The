package lexer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"

	"github.com/EladB1/The/internal/filehandler"
)

func snapshotTestLexer(t *testing.T, filename string) {
	src, err := filehandler.GetSourceCode(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	snapshots := snaps.WithConfig(
		snaps.Dir("testdata/lexer-snaps"),
	)
	tokens, messages := Lex(src)
	var tokenBuilder strings.Builder
	var messagesBuilder strings.Builder
	var formatStr string
	for i, token := range tokens {
		if i != len(tokens)-1 {
			formatStr = fmt.Sprintf("\n\t%v,", token)
		} else {
			formatStr = fmt.Sprintf("\n\t%v\n", token)
		}
		tokenBuilder.WriteString(formatStr)
	}
	for i, msg := range messages {
		if i != len(messages)-1 {
			formatStr = fmt.Sprintf("\n\t\"%v\",", msg)
		} else {
			formatStr = fmt.Sprintf("\n\t\"%v\"\n", msg)
		}
		messagesBuilder.WriteString(formatStr)
	}
	results := fmt.Sprintf("Tokens:\n[%s]\nCompiler messages:\n[%s]\n", tokenBuilder.String(), messagesBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

func TestLexerNonFatal(t *testing.T) {
	t.Run("should run chars.the and have mixed results", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/chars.the")
	})
	t.Run("should run strings.the and have mixed results", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/strings.the")
	})
	t.Run("should run numbers.the and have mixed results", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/numbers.the")
	})

	t.Run("should run symbols.the and have one error", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/symbols.the")
	})
}

/* TODO: Make fatal errors more testable
func TestLexerFatal(t *testing.T) {
	t.Run("Unterminated multiline comment should cause fatal error", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/fatal.the")
	})
}
*/
