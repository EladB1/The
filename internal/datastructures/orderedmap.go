package datastructures

import (
	"fmt"
	"iter"
	"maps"
	"strings"
)

type (
	OrderedMap[T any] struct {
		OrderedKeys []string
		Values      map[string]T
	}
)

func NewOrderedMap[T any]() *OrderedMap[T] {
	return &OrderedMap[T]{
		OrderedKeys: []string{},
		Values:      map[string]T{},
	}
}

func (oMap *OrderedMap[T]) Add(value T, name string) {
	oMap.OrderedKeys = append(oMap.OrderedKeys, name)
	oMap.Values[name] = value
}

func (oMap *OrderedMap[T]) AddAll(other *OrderedMap[T]) {
	oMap.OrderedKeys = append(oMap.OrderedKeys, other.OrderedKeys...)
	maps.Copy(oMap.Values, other.Values)
}

func (oMap *OrderedMap[T]) Update(value T, name string) {
	oMap.Values[name] = value
}

func (oMap *OrderedMap[T]) IsEmpty() bool {
	return len(oMap.Values) == 0
}

func (oMap *OrderedMap[T]) Lookup(name string) (T, bool) {
	value, ok := oMap.Values[name]
	return value, ok
}

func (oMap *OrderedMap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < len(oMap.OrderedKeys); i++ {
			value := oMap.GetByIndex(i)
			if value == nil {
				continue
			}
			if !yield(*value) {
				return
			}
		}
	}
}

func (oMap *OrderedMap[T]) AllWithIndex() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		counter := 0
		for i := 0; i < len(oMap.OrderedKeys); i++ {
			value := oMap.GetByIndex(i)
			if value == nil {
				continue
			}
			if !yield(counter, *value) {
				return
			}
			counter++
		}
	}
}

func (oMap *OrderedMap[T]) GetByIndex(index int) *T {
	name := oMap.OrderedKeys[index]
	if value, ok := oMap.Values[name]; ok {
		copy := value
		return &copy
	}
	return nil
}

func (oMap *OrderedMap[T]) String() string {
	if oMap == nil {
		return ""
	}
	output := strings.Builder{}
	output.WriteString("map[")
	for i, value := range oMap.AllWithIndex() {
		output.WriteString(oMap.OrderedKeys[i])
		output.WriteString(":")
		output.WriteString(fmt.Sprintf("%v", value))
		if i != len(oMap.OrderedKeys)-1 {
			output.WriteString(" ")
		}
	}
	output.WriteString("]")
	return output.String()
}
