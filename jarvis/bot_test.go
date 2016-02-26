package jarvis

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/modules"
)

func TestAddAction(t *testing.T) {
	assert := assert.New(t)
	b := NewBot(slack.UUIDv4().ToShortString())
	b.AddAction(core.Action{ID: "test3", Priority: core.PriorityCatchAll})
	b.AddAction(core.Action{ID: "test1", Priority: core.PriorityHigh})
	b.AddAction(core.Action{ID: "test2", Priority: core.PriorityNormal})
	b.AddAction(core.Action{ID: "test2_passive", Priority: core.PriorityNormal, Passive: true})

	assert.Len(b.mentionActions, 3)
	assert.Len(b.passiveActions, 1)
	assert.Equal("test1", b.mentionActions[0].ID)
	assert.Equal("test2", b.mentionActions[1].ID)
	assert.Equal("test3", b.mentionActions[2].ID)

	allActions := b.Actions()
	assert.Len(allActions, 4)
	assert.Equal("test1", allActions[0].ID)
	assert.Equal("test3", allActions[3].ID)
}

func TestAddActionPriorityCoalesce(t *testing.T) {
	assert := assert.New(t)
	b := NewBot(slack.UUIDv4().ToShortString())
	b.AddAction(core.Action{ID: "test3", Priority: core.PriorityCatchAll})
	b.AddAction(core.Action{ID: "test1", Priority: core.PriorityHigh})
	b.AddAction(core.Action{ID: "test2"})

	assert.Equal("test1", b.mentionActions[0].ID)
	assert.Equal("test2", b.mentionActions[1].ID)
	assert.Equal("test3", b.mentionActions[2].ID)
}

func TestLoadModule(t *testing.T) {
	assert := assert.New(t)
	b := NewBot(slack.UUIDv4().ToShortString())
	b.LoadModule(&modules.Core{})

	assert.NotEmpty(b.Actions())
	assert.NotEmpty(b.mentionActions)
	assert.NotEmpty(b.passiveActions)
}
