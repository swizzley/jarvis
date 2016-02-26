package core

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestIsSalutation(t *testing.T) {
	a := assert.New(t)

	salutations := []string{"hey", "Hey", "hi", "hey jarvis", "hi jarvis", "yo", "yo jarvis"}
	notSalutations := []string{"stuff", "things", "hell no"}

	for _, message := range salutations {
		a.True(IsSalutation(message))
	}

	for _, message := range notSalutations {
		a.False(IsSalutation(message))
	}
}

func TestReplaceAny(t *testing.T) {
	a := assert.New(t)

	text := "this is a test of their things that she likes"
	replaced := ReplaceAny(text, "it", "test", "she")
	a.Equal("this is a it of their things that it likes", replaced)
}

func TestLessMentions(t *testing.T) {
	a := assert.New(t)

	message := "this is a test <@abc123> of mentions <@bca321>"
	lessMentions := LessMentions(message)
	a.Equal("this is a test of mentions ", lessMentions)
}

func TestLessSpecificMention(t *testing.T) {
	a := assert.New(t)

	message := "this is a test <@abc123> of specific mentions <@bca321> etc."
	lessMentions := LessSpecificMention(message, "abc123")
	a.Equal("this is a test of specific mentions <@bca321> etc.", lessMentions)
}

func TestExtract(t *testing.T) {
	a := assert.New(t)

	text := " DSP-7916"

	results := Extract(text, "(DSP-[0-9]+)")
	a.NotEmpty(results)
	a.Equal("DSP-7916", results[0])

	noResults := Extract(text, "(BUGS-[0-9]+)")
	a.Empty(noResults)
}
