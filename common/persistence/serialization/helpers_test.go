package serialization

func emptyMap[K comparable, V any]() map[K]V {
	return map[K]V(nil)
}

func emptySlice[K any]() []K {
	return []K(nil)
}
