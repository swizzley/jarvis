package modules

import (
	"bytes"
	"fmt"

	slack "github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	// ModuleSlack is a label.
	ModuleSlack = "slack"

	// ActionSlackKeeping is a label.
	ActionSlackKeeping = "slack.keeping"

	// ActionSlackKeep is a label.
	ActionSlackKeep = "slack.keep"

	// ActionSlackUnkeep is a label.
	ActionSlackUnkeep = "slack.unkeep"

	// ActionSlackListen is a label.
	ActionSlackListen = "slack.listen"
)

// Slack is a module for slack things.
type Slack struct {
	keepUsers core.ChannelRegistry
}

// Init does nothing for `Slack`.
func (s *Slack) Init(b core.Bot) error { return nil }

// Name returns the module name.
func (s *Slack) Name() string {
	return ModuleSlack
}

// Actions returns the actions for the module.
func (s *Slack) Actions() []core.Action {
	return []core.Action{
		{ID: ActionSlackKeep, MessagePattern: "^keep$", Description: "Keep a user in a channel", Handler: s.handleKeep},
		{ID: ActionSlackKeep, MessagePattern: "^keeping$", Description: "List kept users", Handler: s.handleKeeping},
		{ID: ActionSlackUnkeep, MessagePattern: "^unkeep$", Description: "Dont keep a user in a channel", Handler: s.handleUnkeep},
		{ID: ActionSlackUnkeep, Passive: true, MessagePattern: "(.*)", Description: "Listen for channel events", Handler: s.handleSlackEvent, Priority: core.PriorityCatchAll},
	}
}

func (s *Slack) handleKeep(b core.Bot, m *slack.Message) error {
	user := b.FindUser(m.User)
	channel := b.FindChannel(m.Channel)
	s.keepUsers.Register(b.OrganizationName(), m.Channel, m.User)
	return b.Sayf("Keeping %s in %s", user.Profile.FirstName, channel.Name)
}

func (s *Slack) handleUnkeep(b core.Bot, m *slack.Message) error {
	user := b.FindUser(m.User)
	channel := b.FindChannel(m.Channel)
	s.keepUsers.Unregister(b.OrganizationName(), m.Channel, m.User)
	return b.Sayf("No longer keeping %s in %s", user.Profile.FirstName, channel.Name)
}

func (s *Slack) handleKeeping(b core.Bot, m *slack.Message) error {
	channel := b.FindChannel(m.Channel)
	users := s.keepUsers.UsersInChannel(b.OrganizationName(), m.Channel)
	if len(users) == 0 {
		return b.Sayf("Not keeping any users in %s", channel.Name)
	}

	response := bytes.NewBuffer(nil)
	response.WriteString(fmt.Sprintf("Keeping (%d) users in %s\n", len(users), channel.Name))
	for _, u := range users {
		response.WriteString(fmt.Sprintf("\t - %s\n", u))
	}

	return b.Say(response.String())
}

func (s *Slack) handleSlackEvent(b core.Bot, m *slack.Message) error {
	return nil
}
