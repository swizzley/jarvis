package linq

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestCombinationsOfInt(t *testing.T) {
	assert := assert.New(t)

	combinations := CombinationsOfInt([]int{1, 2, 3, 4})
	assert.Len(combinations, 15)
	assert.Len(combinations[0], 4)
}

func TestCombinationsOfFloat(t *testing.T) {
	assert := assert.New(t)

	combinations := CombinationsOfFloat([]float64{1.0, 2.0, 3.0, 4.0})
	assert.Len(combinations, 15)
	assert.Len(combinations[0], 4)
}

func TestCombinationsOfString(t *testing.T) {
	assert := assert.New(t)

	combinations := CombinationsOfString([]string{"a", "b", "c", "d"})
	assert.Len(combinations, 15)
	assert.Len(combinations[0], 4)
}

func TestPermutationsOfInt(t *testing.T) {
	assert := assert.New(t)

	permutations := PermutationsOfInt([]int{123, 216, 4, 11})
	assert.Len(permutations, 24)
	assert.Len(permutations[0], 4)
}

func TestPermutationsOfFloat(t *testing.T) {
	assert := assert.New(t)

	permutations := PermutationsOfFloat([]float64{3.14, 2.57, 1.0, 6.28})
	assert.Len(permutations, 24)
	assert.Len(permutations[0], 4)
}

func TestPermutationsOfString(t *testing.T) {
	assert := assert.New(t)

	permutations := PermutationsOfString([]string{"a", "b", "c", "d"})
	assert.Len(permutations, 24)
	assert.Len(permutations[0], 4)
}

func TestPermuteDistributions(t *testing.T) {
	assert := assert.New(t)

	permutations := PermuteDistributions(4, 2)
	assert.Len(permutations, 5)

	assert.Len(permutations[0], 2)
}
