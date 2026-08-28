package datastructures

import (
	"fmt"

	"github.com/EladB1/The/internal/config"
)

type LiteralPool []LengthPrefixString

// Append the literal and return the last index
func (pool LiteralPool) Add(value string) (LiteralPool, int) {
	index := len(pool)
	pool = append(pool, LengthPrefixString{
		Length: len(value),
		Str:    value,
	})
	return pool, index
}

func (pool LiteralPool) Show(envconf *config.EnvConfig) {
	fmt.Fprintln(envconf.Stdout, "[")
	for i, literal := range pool {
		if i == len(pool)-1 {
			fmt.Fprintf(envconf.Stdout, "\t%v\n", literal)
		} else {
			fmt.Fprintf(envconf.Stdout, "\t%v,\n", literal)
		}
	}
	fmt.Fprintln(envconf.Stdout, "]")
}
