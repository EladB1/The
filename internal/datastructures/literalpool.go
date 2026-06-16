package datastructures

import "fmt"

// Store string, float, and int literals
type Literal string

type LiteralPool []Literal

// Append the literal and return the last index
func (pool LiteralPool) Add(value string) (LiteralPool, int) {
	index := len(pool)
	pool = append(pool, Literal(value))
	return pool, index
}

func (pool LiteralPool) Show() {
	fmt.Println("[")
	for i, literal := range pool {
		if i == len(pool)-1 {
			fmt.Printf("\t%s\n", literal)
		} else {
			fmt.Printf("\t%s,\n", literal)
		}
	}
	fmt.Println("]")
}

var LiteralStorage LiteralPool
