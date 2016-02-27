package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestAwaitAll(t *testing.T) {
	assert := assert.New(t)
	first_did_run := false
	second_did_run := false
	third_did_run := false
	AwaitAll(func() {
		first_did_run = true
	}, func() {
		second_did_run = true
	}, func() {
		third_did_run = true
	})

	assert.True(first_did_run)
	assert.True(second_did_run)
	assert.True(third_did_run)
}
