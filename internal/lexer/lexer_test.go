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
	tokens, messages := Lex(src, false)
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
	for i, msg := range messages.Messages {
		if i != len(messages.Messages)-1 {
			formatStr = fmt.Sprintf("\n\t\"%v\",", msg)
		} else {
			formatStr = fmt.Sprintf("\n\t\"%v\"\n", msg)
		}
		messagesBuilder.WriteString(formatStr)
	}
	results := fmt.Sprintf("Tokens:\n[%s]\nCompiler messages:\n[%s]\n", tokenBuilder.String(), messagesBuilder.String())
	snapshots.MatchSnapshot(t, results)
}

func TestLexer(t *testing.T) {
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
	t.Run("should run endless_comment.the and result in an error", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/endless_comment.the")
	})
	t.Run("should run struct.the and get no errors", func(t *testing.T) {
		snapshotTestLexer(t, "testdata/fixtures/struct.the")
	})
	t.Run("token.HasValue() should return true when given matching value", func(t *testing.T) {
		token := Token{
			Kind:  ID,
			Value: "name",
		}
		if !token.HasValue("name") {
			t.Errorf("Token %v should have matched value %s\n", token, "name")
		}
	})
	t.Run("token.HasValue() should return false when given non-matching value", func(t *testing.T) {
		token := Token{
			Kind:  ID,
			Value: "name",
		}
		if token.HasValue("x") {
			t.Errorf("Token %v should not have matched value %s\n", token, "x")
		}
	})
}
