package lib

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestIsSalutation(t *testing.T) {
	a := assert.New(t)

	salutations := []string{"hey", "Hey", "hi", "hey jarvis", "hi jarvis", "yo", "yo jarvis"}
	notSalutations := []string{"stuff", "things", "hell no"}

	for _, message := range salutations {
		a.True(isSalutation(message))
	}

	for _, message := range notSalutations {
		a.False(isSalutation(message))
	}
}
