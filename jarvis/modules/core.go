package modules

import (
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	// ModuleCore is the name of the core module.
	ModuleCore = "core"

	// ActionHelp is the name of the ActionHelp action.
	ActionHelp = "help"

	// ActionTime is the name of the ActionTime action.
	ActionTime = "time"

	// ActionTell is the name of the ActionTell action.
	ActionTell = "tell"

	// ActionChannels is the name of the ActionChannels action.
	ActionChannels = "channels"

	// ActionMentionCatchAll is the name of the ActionMentionCatchAll action.
	ActionMentionCatchAll = "mention.catch_all"

	// ActionPassiveCatchAll is the name of the ActionPassiveCatchAll action.
	ActionPassiveCatchAll = "passive.catch_all"

	// ActionSalutation is the name of the ActionSalutation action.
	ActionSalutation = "salutation"

	// ActionUnknown is the name of the ActionUnknown action.
	ActionUnknown = "unknown"
)

// Core is the module that handles basic methods
type Core struct{}

// Name returns the name of the module
func (c *Core) Name() string {
	return ModuleCore
}

// Actions returns mention commands for the core module.
func (c *Core) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionHelp, MessagePattern: "^help", Description: "Prints help info.", Handler: c.handleHelp},
		core.Action{ID: ActionTime, MessagePattern: "^time", Description: "Prints the current time.", Handler: c.handleTime},
		core.Action{ID: ActionTell, MessagePattern: "^tell", Description: "Tell people things.", Handler: c.handleTell},
		core.Action{ID: ActionChannels, MessagePattern: "^channels", Description: "Prints the channels I'm currently listening to.", Handler: c.handleChannels},

		core.Action{ID: ActionMentionCatchAll, MessagePattern: "(.*)", Description: "I'll do the best I can.", Handler: c.handleMentionCatchAll, Priority: core.PriorityCatchAll},
		core.Action{ID: ActionPassiveCatchAll, Passive: true, MessagePattern: "(.*)", Description: "I'll do the best I can (passively).", Handler: c.handlePassiveCatchAll, Priority: core.PriorityCatchAll},

		core.Action{ID: ActionSalutation, Description: "Salutation Response.", Handler: c.handleSalutation},
		core.Action{ID: ActionUnknown, Description: "Unknown Response.", Handler: c.handleUnknown},
	}
}

func (c *Core) handleHelp(b core.Bot, m *slack.Message) error {
	responseText := "Here are the commands that are currently configured:"
	for _, actionHandler := range b.Actions() {
		if !actionHandler.Passive {
			responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.MessagePattern, actionHandler.Description)
		}
	}
	responseText = responseText + "\nWith the following passive commands:"
	for _, actionHandler := range b.Actions() {
		if actionHandler.Passive {
			responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.MessagePattern, actionHandler.Description)
		}
	}
	return b.Say(m.Channel, responseText)
}

func (c *Core) handleTime(b core.Bot, m *slack.Message) error {
	timeText := fmt.Sprintf("%s UTC", time.Now().UTC().Format(time.Kitchen))
	message := slack.NewChatMessage(m.Channel, "")
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

	_, messageErr := b.Client().ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}

func (c *Core) handleTell(b core.Bot, m *slack.Message) error {
	messageText := core.LessSpecificMention(m.Text, b.ID())
	words := strings.Split(messageText, " ")

	destinationUser := ""
	tellMessage := ""

	for x := 0; x < len(words); x++ {
		word := words[x]
		if core.Like(word, "tell") {
			continue
		} else if core.IsMention(word) {
			destinationUser = word
			tellMessage = strings.Join(words[x+1:], " ")
		}
	}
	tellMessage = core.ReplaceAny(tellMessage, "you are", "shes", "she's", "she is", "hes", "he's", "he is", "theyre", "they're", "they are")
	resultMessage := fmt.Sprintf("%s %s", destinationUser, tellMessage)
	return b.Say(m.Channel, resultMessage)
}

func (c *Core) handleChannels(b core.Bot, m *slack.Message) error {
	if len(b.ActiveChannels()) == 0 {
		return b.Say(m.Channel, "currently listening to *no* channels.")
	}
	activeChannelsText := "currently listening to the following channels:\n"
	for _, channelID := range b.ActiveChannels() {
		if channel := b.FindChannel(channelID); channel != nil {
			activeChannelsText = activeChannelsText + fmt.Sprintf(">#%s (id:%s)\n", channel.Name, channel.ID)
		}
	}
	return b.Say(m.Channel, activeChannelsText)
}

func (c *Core) handleSalutation(b core.Bot, m *slack.Message) error {
	user := b.FindUser(m.User)
	salutation := []string{"hey %s", "hi %s", "hello %s", "ohayo gozaimasu %s", "salut %s", "bonjour %s", "yo %s", "sup %s"}
	return b.Sayf(m.Channel, core.Random(salutation), strings.ToLower(user.Profile.FirstName))
}

func (c *Core) handleMentionCatchAll(b core.Bot, m *slack.Message) error {
	message := util.TrimWhitespace(core.LessMentions(m.Text))
	if core.IsSalutation(message) {
		return c.handleSalutation(b, m)
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
