package modules

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestExtractJiraIssues(t *testing.T) {
	a := assert.New(t)

	jb := &Jira{}

	text := "something DSP-1234 DSP-4321 BUGS-1234 not-an-issue"
	issueIds := jb.extractJiraIssues(text)
	a.Len(issueIds, 3)
	a.Equal("DSP-1234", issueIds[0])
	a.Equal("DSP-4321", issueIds[1])
	a.Equal("BUGS-1234", issueIds[2])
}
