package irgen

import (
	"os"
	"testing"
)

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") != "true" {
		t.Skip()
	}
	// TODO
}
