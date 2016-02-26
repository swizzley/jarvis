package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	ModuleCore            = "core"
	ActionHelp            = "help"
	ActionTime            = "time"
	ActionTell            = "tell"
	ActionChannels        = "channels"
	ActionMentionCatchAll = "mention.catch_all"
	ActionPassiveCatchAll = "passive.catch_all"
)

// Core is the module that handles basic methods
type Core struct{}

// Name returns the name of the module
func (c *Core) Name() string {
	return ModuleCore
}

// MentionCommands returns mention commands for the core module.
func (c *Core) MentionCommands() []core.Action {
	return []core.Action{
		core.Action{ID: ActionHelp, MessagePattern: "^help", Description: "Prints help info.", Handler: c.handleHelp},
		core.Action{ID: ActionTime, MessagePattern: "^time", Description: "Prints the current time.", Handler: c.handleTime},
		core.Action{ID: ActionTell, MessagePattern: "^tell", Description: "Tell people things.", Handler: c.handleTell},
		core.Action{ID: ActionChannels, MessagePattern: "^channels", Description: "Prints the channels I'm currently listening to.", Handler: j.handleChannels},
		core.Action{ID: ActionMentionCatchAll, MessagePattern: "(.*)", Description: "I'll do the best I can.", Handler: c.handleMentionCatchAll},
	}
}

// PassiveCommands returns passive commands for the core module.
func (c *Core) PassiveCommands() []core.Action {
	return []core.Action{
		core.Action{ID: "passive.catch_all", MessagePattern: "(.*)", Description: "I'll do the best I can.", Handler: c.handlePassiveCatchAll},
	}
}

func (c *Core) handleHelp(b core.Bot, m *slack.Message) error {
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

func (c *Core) handleMentionCatchAll(b core.Bot, m *slack.Message) error {
	message := util.TrimWhitespace(core.LessMentions(m.Text))
	if core.IsSalutation(message) {
		user := b.FindUser(m.User)
		salutation := []string{"hey %s", "hi %s", "hello %s", "ohayo gozaimasu %s", "salut %s", "bonjour %s", "yo %s", "sup %s"}
		return b.Sayf(m.Channel, core.Random(salutation), strings.ToLower(user.Profile.FirstName))
	}
	return c.handleUnknown(b, m)
}

func (c *Core) handlePassiveCatchAll(b core.Bot, m *slack.Message) error {
	message := util.TrimWhitespace(core.LessMentions(m.Text))
	if core.IsAngry(message) {
		user := b.FindUser(m.User)
		response := []string{"slow down %s", "maybe calm down %s", "%s you should really relax", "chill %s", "it's ok %s, let it out"}
		return b.Sayf(m.Channel, core.Random(response), strings.ToLower(user.Profile.FirstName))
	}
	return nil
}

func (c *Core) handleUnknown(b core.Bot, m *slack.Message) error {
	return b.Sayf(m.Channel, "I don't know how to respond to this\n>%s", m.Text)
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
