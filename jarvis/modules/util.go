package modules

import (
	"fmt"

	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	// ModuleUtil is the util module.
	ModuleUtil = "util"

	// ActionUtilUserID is the action that fetches a userid for a mention.
	ActionUtilUserID = "util.user_id"
)

// Util is a set of slack specific utility commands.
type Util struct{}

// Init does nothing right now.
func (u *Util) Init(b core.Bot) error { return nil }

// Name is the name of the module.
func (u Util) Name() string {
	return ModuleUtil
}

// Actions are the actions for the module.
func (u Util) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionUtilUserID, MessagePattern: "^user", Description: "Get the Slack user_id for a given user.", Handler: u.handleUserID},
	}
}

func (u Util) handleUserID(b core.Bot, m *slack.Message) error {
	messageText := core.LessSpecificMention(m.Text, b.ID())
	mentionedUserIDs := core.Mentions(messageText)

	outputText := "I looked up the following users:\n"
	for _, userID := range mentionedUserIDs {
		user := b.FindUser(userID)
		outputText = outputText + fmt.Sprintf("> %s : %s %s", userID, user.Profile.FirstName, user.Profile.LastName)
	}

	return b.Say(m.Channel, outputText)
}
