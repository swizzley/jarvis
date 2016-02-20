package collections

import "strings"

type StringSet map[string]bool

func (ss StringSet) Add(entry string) {
	if _, hasEntry := ss[entry]; !hasEntry {
		ss[entry] = true
	}
}

func (ss StringSet) Contains(entry string) bool {
	if _, hasEntry := ss[entry]; hasEntry {
		return true
	} else {
		return false
	}
}

func (ss StringSet) Remove(entry string) bool {
	if _, hasEntry := ss[entry]; hasEntry {
		delete(ss, entry)
		return true
	}
	return false
}

func (ss StringSet) Len() int {
	return len(ss)
}

func (ss StringSet) Copy() StringSet {
	newSet := StringSet{}
	for key, _ := range ss {
		newSet.Add(key)
	}
	return newSet
}

func (ss StringSet) ToArray() []string {
	output := []string{}
	for key, _ := range ss {
		output = append(output, key)
	}
	return output
}

func (ss StringSet) String() string {
	return strings.Join(ss.ToArray(), ", ")
}
