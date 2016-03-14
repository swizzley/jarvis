package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestAwaitAll(t *testing.T) {
	assert := assert.New(t)
	firstDidRun := false
	secondDidRun := false
	thirdDidRun := false
	AwaitAll(func() {
		firstDidRun = true
	}, func() {
		secondDidRun = true
	}, func() {
		thirdDidRun = true
	})

	assert.True(firstDidRun)
	assert.True(secondDidRun)
	assert.True(thirdDidRun)
}
