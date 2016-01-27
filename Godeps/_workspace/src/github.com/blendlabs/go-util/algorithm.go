package util

import "reflect"

// Combinations returns the "power set" of values less the empty set.
// Use "combinations" when the order of the resulting sets do not matter.
func CombinationsOfInt(values []int) [][]int {
	possibleValues := PowOfInt(2, uint(len(values))) //less the empty entry
	output := make([][]int, possibleValues-1)

	for x := 0; x < possibleValues-1; x++ {
		row := []int{}
		for i := 0; i < len(values); i++ {
			y := 1 << uint(i)
			if y&x == 0 && y != x {
				row = append(row, values[i])
			}
		}
		if len(row) > 0 {
			output[x] = row
		}
	}
	return output
}

// Combinations returns the "power set" of values less the empty set.
// Use "combinations" when the order of the resulting sets do not matter.
func CombinationsOfFloat(values []float64) [][]float64 {
	possibleValues := PowOfInt(2, uint(len(values))) //less the empty entry
	output := make([][]float64, possibleValues-1)

	for x := 0; x < possibleValues-1; x++ {
		row := []float64{}
		for i := 0; i < len(values); i++ {
			y := 1 << uint(i)
			if y&x == 0 && y != x {
				row = append(row, values[i])
			}
		}
		if len(row) > 0 {
			output[x] = row
		}
	}
	return output
}

// Combinations returns the "power set" of values less the empty set.
// Use "combinations" when the order of the resulting sets do not matter.
func CombinationsOfString(values []string) [][]string {
	possibleValues := PowOfInt(2, uint(len(values))) //less the empty entry
	output := make([][]string, possibleValues-1)

	for x := 0; x < possibleValues-1; x++ {
		row := []string{}
		for i := 0; i < len(values); i++ {
			y := 1 << uint(i)
			if y&x == 0 && y != x {
				row = append(row, values[i])
			}
		}
		if len(row) > 0 {
			output[x] = row
		}
	}
	return output
}

// Permutations returns the possible orderings of the values array.
// Use "permutations" when order matters.
func PermutationsOfInt(values []int) [][]int {
	if len(values) == 1 {
		return [][]int{values}
	}

	output := [][]int{}
	for x := 0; x < len(values); x++ {
		workingValues := make([]int, len(values))
		copy(workingValues, values)
		value := workingValues[x]
		pre := workingValues[0:x]
		post := workingValues[x+1 : len(values)]

		joined := append(pre, post...)

		for _, inner := range PermutationsOfInt(joined) {
			output = append(output, append([]int{value}, inner...))
		}
	}

	return output
}

// Permutations returns the possible orderings of the values array.
// Use "permutations" when order matters.
func PermutationsOfFloat(values []float64) [][]float64 {
	if len(values) == 1 {
		return [][]float64{values}
	}

	output := [][]float64{}
	for x := 0; x < len(values); x++ {
		workingValues := make([]float64, len(values))
		copy(workingValues, values)
		value := workingValues[x]
		pre := workingValues[0:x]
		post := workingValues[x+1 : len(values)]

		joined := append(pre, post...)

		for _, inner := range PermutationsOfFloat(joined) {
			output = append(output, append([]float64{value}, inner...))
		}
	}

	return output
}

// Permutations returns the possible orderings of the values array.
// Use "permutations" when order matters.
func PermutationsOfString(values []string) [][]string {
	if len(values) == 1 {
		return [][]string{values}
	}

	output := [][]string{}
	for x := 0; x < len(values); x++ {
		workingValues := make([]string, len(values))
		copy(workingValues, values)
		value := workingValues[x]
		pre := workingValues[0:x]
		post := workingValues[x+1 : len(values)]
		joined := append(pre, post...)
		for _, inner := range PermutationsOfString(joined) {
			output = append(output, append([]string{value}, inner...))
		}
	}

	return output
}

// PermuteDistributions returns all the possible ways you can split a total among buckets completely.
func PermuteDistributions(total, buckets int) [][]int {
	return PermuteDistributionsFromExisting(total, buckets, []int{})
}

// PermuteDistributionsFromExisting returns all the possible ways you can split the total among additional buckets
// given an existing distribution
func PermuteDistributionsFromExisting(total, buckets int, existing []int) [][]int {
	output := [][]int{}
	existingLength := len(existing)
	existingSum := SumOfInt(existing)
	remainder := total - existingSum

	if buckets == 1 {
		newExisting := make([]int, existingLength+1)
		copy(newExisting, existing)
		newExisting[existingLength] = remainder
		output = append(output, newExisting)
		return output
	}

	for x := 0; x <= remainder; x++ {
		newExisting := make([]int, existingLength+1)
		copy(newExisting, existing)
		newExisting[existingLength] = x

		results := PermuteDistributionsFromExisting(total, buckets-1, newExisting)
		output = append(output, results...)
	}

	return output
}

type Predicate func(item interface{}) bool
type PredicateOfInt func(item int) bool
type PredicateOfFloat func(item float64) bool
type PredicateOfString func(item string) bool

func Any(target interface{}, predicate Predicate) bool {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return false
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfInt(target []int, predicate PredicateOfInt) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfFloat(target []float64, predicate PredicateOfFloat) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfString(target []string, predicate PredicateOfString) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if predicate(obj) {
			return true
		}
	}
	return false
}

func All(target interface{}, predicate Predicate) bool {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return false
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfInt(target []int, predicate PredicateOfInt) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfFloat(target []float64, predicate PredicateOfFloat) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfString(target []string, predicate PredicateOfString) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if !predicate(obj) {
			return false
		}
	}
	return true
}
