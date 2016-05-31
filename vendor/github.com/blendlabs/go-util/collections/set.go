package collections

import (
	"strconv"
	"strings"
)

// NewSetOfInt returns a new SetOfInt
func NewSetOfInt(values []int) SetOfInt {
	set := SetOfInt{}
	for _, v := range values {
		set.Add(v)
	}
	return set
}

// SetOfInt is a type alias for map[int]int
type SetOfInt map[int]bool

// Add adds an element to the set, replaceing a previous value.
func (si SetOfInt) Add(i int) {
	si[i] = true
}

// Remove removes an element from the set.
func (si SetOfInt) Remove(i int) {
	delete(si, i)
}

// Contains returns if the element is in the set.
func (si SetOfInt) Contains(i int) bool {
	_, ok := si[i]
	return ok
}

// Len returns the number of elements in the set.
func (si SetOfInt) Len() int {
	return len(si)
}

// AsSlice returns the set as a slice.
func (si SetOfInt) AsSlice() []int {
	output := []int{}
	for key := range si {
		output = append(output, key)
	}
	return output
}

// String returns the set as a csv string.
func (si SetOfInt) String() string {
	var values []string
	for _, i := range si.AsSlice() {
		values = append(values, strconv.Itoa(i))
	}

	return strings.Join(values, ", ")
}

// NewSetOfString returns a new SetOfString.
func NewSetOfString(values []string) SetOfString {
	set := SetOfString{}
	for _, v := range values {
		set.Add(v)
	}
	return set
}

// SetOfString is a set of strings
type SetOfString map[string]bool

// Add adds an element.
func (ss SetOfString) Add(entry string) {
	if _, hasEntry := ss[entry]; !hasEntry {
		ss[entry] = true
	}
}

// Remove deletes an element, returns if the element was in the set.
func (ss SetOfString) Remove(entry string) bool {
	if _, hasEntry := ss[entry]; hasEntry {
		delete(ss, entry)
		return true
	}
	return false
}

// Contains returns if an element is in the set.
func (ss SetOfString) Contains(entry string) bool {
	_, hasEntry := ss[entry]
	return hasEntry
}

// Len returns the length of the set.
func (ss SetOfString) Len() int {
	return len(ss)
}

// Copy returns a new copy of the set.
func (ss SetOfString) Copy() SetOfString {
	newSet := SetOfString{}
	for key := range ss {
		newSet.Add(key)
	}
	return newSet
}

// AsSlice returns the set as a slice.
func (ss SetOfString) AsSlice() []string {
	output := []string{}
	for key := range ss {
		output = append(output, key)
	}
	return output
}

// String returns the set as a csv string.
func (ss SetOfString) String() string {
	return strings.Join(ss.AsSlice(), ", ")
}
