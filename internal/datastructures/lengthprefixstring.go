package datastructures

import (
	"fmt"
	"strings"
)

type LengthPrefixString struct {
	Length int
	Str    string
}

func (lps LengthPrefixString) String() string {
	return fmt.Sprintf("{Length: %d, String: %s}", lps.Length, lps.Str)
}

func (lps LengthPrefixString) WasmString() string {
	little_endian := []int{
		lps.Length & 0xff,
		(lps.Length >> 8) & 0xff,
		(lps.Length >> 16) & 0xff,
		(lps.Length >> 24) & 0xff,
	}
	lengthStr := strings.Builder{}
	for _, l_byte := range little_endian {
		lengthStr.WriteString(fmt.Sprintf("\\%02x", l_byte))
	}
	lengthStr.WriteString(lps.Str)
	return lengthStr.String()
}
