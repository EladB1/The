package datastructures

import (
	"fmt"

	"github.com/EladB1/The/internal/config"
)

// Store string, float, and int literals
type Literal string

type LiteralPool []Literal

// Append the literal and return the last index
func (pool LiteralPool) Add(value string) (LiteralPool, int) {
	index := len(pool)
	pool = append(pool, Literal(value))
	return pool, index
}

func (pool LiteralPool) Show(envconf *config.EnvConfig) {
	fmt.Fprintln(envconf.Stdout, "[")
	for i, literal := range pool {
		if i == len(pool)-1 {
			fmt.Fprintf(envconf.Stdout, "\t%s\n", literal)
		} else {
			fmt.Fprintf(envconf.Stdout, "\t%s,\n", literal)
		}
	}
	fmt.Fprintln(envconf.Stdout, "]")
}
