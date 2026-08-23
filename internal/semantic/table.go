package semantic

import "iter"

type (
	SymbolTable[T any] struct {
		OrderedNames []string
		Symbols      map[string]T
	}

	FunctionTable = SymbolTable[FunctionSymbol]
)

func NewTable[T any]() *SymbolTable[T] {
	return &SymbolTable[T]{
		OrderedNames: []string{},
		Symbols:      map[string]T{},
	}
}

func (table *SymbolTable[T]) Add(symbol T, name string) {
	table.OrderedNames = append(table.OrderedNames, name)
	table.Symbols[name] = symbol
}

func (table *SymbolTable[T]) update(symbol T, name string) {
	table.Symbols[name] = symbol
}

func (table *SymbolTable[T]) isEmpty() bool {
	return len(table.Symbols) == 0
}

func (table *SymbolTable[T]) Lookup(name string) (T, bool) {
	symbol, ok := table.Symbols[name]
	return symbol, ok
}

func (table *SymbolTable[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < len(table.OrderedNames); i++ {
			symbol := table.GetByIndex(i)
			if symbol == nil {
				continue
			}
			if !yield(*symbol) {
				return
			}
		}
	}
}

func (table *SymbolTable[T]) GetByIndex(index int) *T {
	name := table.OrderedNames[index]
	if symbol, ok := table.Symbols[name]; ok {
		copy := symbol
		return &copy
	}
	return nil
}
