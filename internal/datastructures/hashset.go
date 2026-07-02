package datastructures

type HashSet map[string]struct{}

func BuildHashSet(items ...string) HashSet {
	set := make(map[string]struct{})
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func (hs HashSet) Append(item string) {
	hs[item] = struct{}{}
}
