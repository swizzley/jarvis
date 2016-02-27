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

type Util struct{}

func (u Util) Name() string {
	return ModuleUtil
}

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
