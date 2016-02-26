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

	regression1 := "<@U0K0N6L85>: run ping <http://google.com|google.com>"
	regression1LessMentions := LessSpecificMention(regression1, "U0K0N6L85")
	a.Equal("run ping <http://google.com|google.com>", regression1LessMentions)
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

func TestFixLinks(t *testing.T) {
	a := assert.New(t)

	messageWithLink := "this is a test <http://google.com|google.com> of links"
	messageWithLinkAndMention := "this is a test <http://google.com|google.com> of links and <@abc1234> of mentions"
	messageWithMention := "this is a test <@abc123> of mentions"
	messageWithoutAnything := "this is a test"
	regression1 := " run ping <http://google.com|google.com>"

	a.Equal("this is a test google.com of links", FixLinks(messageWithLink))
	a.Equal("this is a test google.com of links and <@abc1234> of mentions", FixLinks(messageWithLinkAndMention))
	a.Equal("this is a test <@abc123> of mentions", FixLinks(messageWithMention))
	a.Equal("this is a test", FixLinks(messageWithoutAnything))

	a.Equal(" run ping google.com", FixLinks(regression1))
}
