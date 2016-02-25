package collections

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestIntSet(t *testing.T) {
	a := assert.New(t)

	set := IntSet{}
	set.Add(1)
	a.True(set.Contains(1))
	a.False(set.Contains(2))
	set.Remove(1)
	a.False(set.Contains(1))
}
