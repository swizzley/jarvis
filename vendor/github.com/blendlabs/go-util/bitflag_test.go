package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestBitFlagCombine(t *testing.T) {
	assert := assert.New(t)

	three := BitFlagCombine(1, 2)
	assert.Equal(3, three)
}

func TestBitFlagAny(t *testing.T) {
	assert := assert.New(t)

	one := 1 << 0
	two := 1 << 1
	four := 1 << 2
	eight := 1 << 3
	sixteen := 1 << 4
	invalid := 1 << 5

	masterFlag := BitFlagCombine(one, two, four, eight)
	checkFlag := BitFlagCombine(one, sixteen)
	assert.True(BitFlagAny(masterFlag, checkFlag))
	assert.False(BitFlagAny(masterFlag, invalid))
}

func TestBitFlagAll(t *testing.T) {
	assert := assert.New(t)

	one := 1 << 0
	two := 1 << 1
	four := 1 << 2
	eight := 1 << 3
	sixteen := 1 << 4

	masterFlag := BitFlagCombine(one, two, four, eight)
	checkValidFlag := BitFlagCombine(one, two)
	checkInvalidFlag := BitFlagCombine(one, sixteen)
	assert.True(BitFlagAll(masterFlag, checkValidFlag))
	assert.False(BitFlagAll(masterFlag, checkInvalidFlag))
}
