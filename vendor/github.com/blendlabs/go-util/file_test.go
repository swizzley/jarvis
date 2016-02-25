package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestReadFileByLines(t *testing.T) {
	assert := assert.New(t)

	called := false
	ReadFileByLines("README.md", func(line string) {
		called = true
	})

	assert.True(called, "We should have called the handler for `README.md`")
}

func TestReadFileByChunks(t *testing.T) {
	assert := assert.New(t)

	called := false
	ReadFileByChunks("README.md", 32, func(chunk []byte) {
		called = true
	})

	assert.True(called, "We should have called the handler for `README.md`")
}
