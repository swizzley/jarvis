package collections

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestStringSet(t *testing.T) {
	assert := assert.New(t)

	set := StringSet{}
	assert.Equal(0, set.Len())

	set.Add("test")
	assert.Equal(1, set.Len())
	assert.True(set.Contains("test"))

	set.Add("test")
	assert.Equal(1, set.Len())
	assert.True(set.Contains("test"))

	set.Add("not test")
	assert.Equal(2, set.Len())
	assert.True(set.Contains("not test"))

	set.Remove("test")
	assert.Equal(1, set.Len())
	assert.False(set.Contains("test"))
	assert.True(set.Contains("not test"))

	set.Remove("not test")
	assert.Equal(0, set.Len())
	assert.False(set.Contains("test"))
	assert.False(set.Contains("not test"))
}
