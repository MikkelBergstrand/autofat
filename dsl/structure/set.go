package structure

import (
	"cmp"
	"sort"
)

// Set is a collection of unique elements
type Set[K cmp.Ordered] struct {
	elements map[K]struct{}
}

// NewSet creates a new set
func NewSet[K cmp.Ordered]() *Set[K] {
	return &Set[K]{
		elements: make(map[K]struct{}),
	}
}

func (s Set[K]) Copy() Set[K] {
	newSet := *NewSet[K]()
	for k := range s.elements {
		newSet.elements[k] = struct{}{}
	}
	return newSet
}

// Add inserts an element into the set
func (s Set[K]) Add(value K) Set[K] {
	s.elements[value] = struct{}{}
	return s
}

// Remove deletes an element from the set
func (s Set[K]) Remove(value K) Set[K] {
	delete(s.elements, value)
	return s
}

// Contains checks if an element is in the set
func (s Set[K]) Contains(value K) bool {
	_, found := s.elements[value]
	return found
}

// Size returns the number of elements in the set
func (s Set[K]) Size() int {
	return len(s.elements)
}

// List returns all elements in the set as a slice
func (s Set[K]) List() []K {
	keys := make([]K, 0, len(s.elements))
	for key := range s.elements {
		keys = append(keys, key)
	}
	return keys
}

func (s Set[K]) SortedList() []K {
	ret := s.List()
	sort.Slice(ret, func(i, j int) bool { return cmp.Compare(ret[i], ret[j]) > 0 })
	return ret
}

func (s Set[K]) Union(other Set[K]) Set[K] {
	for k := range other.elements {
		s.elements[k] = struct{}{}
	}
	return s
}

func (s Set[K]) Elements() map[K]struct{} {
	return s.elements
}
