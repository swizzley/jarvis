package core

import (
	"fmt"
	"time"

	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

type Core struct {
}

func (c *Core) MentionCommands() []core.Action {
	return []core.Action{
		core.Action{ID: "help", MessagePattern: "^help", Description: "Prints help info.", Handler: c.HandleHelp},
		core.Action{ID: "time", MessagePattern: "^time", Description: "Prints the current time.", Handler: j.HandleTime},
		core.Action{"^tell", "Tell people things.", j.DoTell},
		core.Action{"^channels", "Prints the channels I'm currently listening to.", j.DoChannels},
		core.Action{"mention.catch_all", "(.*)", "I'll do the best I can.", j.DoOtherResponse},
	}
}

func (c *Core) PassiveCommands() []Action {
	return []core.Action{
		core.Action{"passive.catch_all", "(.*)", "I'll do the best I can.", j.DoOtherPassiveResponse},
	}
}

func (c *Core) HandleHelp(b core.Bot, m *slack.Message) error {
	responseText := "Here are the commands that are currently configured:"
	for _, actionHandler := range b.MentionCommands() {
		responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.MessagePattern, actionHandler.Description)
	}
	responseText = responseText + "\nWith the following passive commands:"
	for _, actionHandler := range b.PassiveCommands() {
		responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.MessagePattern, actionHandler.Description)
	}
	return b.Say(m.Channel, responseText)
}

func (c *Core) announceTime(b core.Bot, destinationID string, currentTime time.Time) error {
	timeText := fmt.Sprintf("%s UTC", currentTime.Format(time.Kitchen))
	message := slack.NewChatMessage(destinationID, "")
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	message.UnfurlMedia = slack.OptionalBool(false)
	message.Attachments = []slack.ChatMessageAttachment{
		slack.ChatMessageAttachment{
			Fallback: fmt.Sprintf("The time is now:\n>%s", timeText),
			Color:    slack.OptionalString("#4099FF"),
			Pretext:  slack.OptionalString("The time is now:"),
			Text:     slack.OptionalString(timeText),
		},
	}

	_, messageErr := b.Client.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}
