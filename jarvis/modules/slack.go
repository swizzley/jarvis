package modules

import (
	"bytes"
	"fmt"
	"strings"

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
		{ID: ActionSlackKeeping, MessagePattern: "^keeping", Description: "List kept users", Handler: s.handleKeeping},
		{ID: ActionSlackKeep, MessagePattern: "^keep", Description: "Keep a user in a channel", Handler: s.handleKeep},
		{ID: ActionSlackUnkeep, MessagePattern: "^unkeep", Description: "Dont keep a user in a channel", Handler: s.handleUnkeep},
		{ID: ActionSlackUnkeep, Passive: true, MessagePattern: "(.*)", Description: "Listen for channel events", Handler: s.handleSlackEvent, Priority: core.PriorityCatchAll},
	}
}

func (s *Slack) handleKeep(b core.Bot, m *slack.Message) error {
	mentions := core.Mentions(m.Text)
	if len(mentions) == 0 {
		return b.Sayf(m.Channel, "Need to mention (1) user")
	}

	channel := b.FindChannel(m.Channel)

	var users []string
	for _, user := range mentions {
		fmt.Printf("mentioned: %s\n", user)
		user := b.FindUser(user)

		if user != nil {
			s.keepUsers.Register(b.OrganizationName(), m.Channel, user.ID)
			users = append(users, user.Profile.FirstName)
		}
	}
	if len(users) == 0 {
		return b.Say(m.Channel, "Need to mention (1) valid user.")
	}
	return b.Sayf(m.Channel, "Keeping %s in %s", strings.Join(users, ", "), channel.Name)
}

func (s *Slack) handleUnkeep(b core.Bot, m *slack.Message) error {
	mentions := core.Mentions(m.Text)
	if len(mentions) == 0 {
		return b.Sayf(m.Channel, "Need to mention (1) user")
	}

	channel := b.FindChannel(m.Channel)

	var users []string
	for _, user := range mentions {
		user := b.FindUser(user)

		if user != nil {
			s.keepUsers.Unregister(b.OrganizationName(), m.Channel, user.ID)
			users = append(users, user.Profile.FirstName)
		}
	}
	if len(users) == 0 {
		return b.Say(m.Channel, "Need to mention (1) valid user.")
	}
	return b.Sayf(m.Channel, "No longer keeping %s in %s", strings.Join(users, ", "), channel.Name)
}

func (s *Slack) handleKeeping(b core.Bot, m *slack.Message) error {
	channel := b.FindChannel(m.Channel)
	users := s.keepUsers.UsersInChannel(b.OrganizationName(), m.Channel)
	if len(users) == 0 {
		return b.Sayf(m.Channel, "Not keeping any users in %s", channel.Name)
	}

	response := bytes.NewBuffer(nil)
	response.WriteString(fmt.Sprintf("Keeping (%d) users in %s\n", len(users), channel.Name))
	for _, u := range users {
		response.WriteString(fmt.Sprintf("\t - %s\n", u))
	}
	return b.Say(m.Channel, response.String())
}

func (s *Slack) handleSlackEvent(b core.Bot, m *slack.Message) error {
	if slack.Event(m.SubType) == slack.EventSubtypeChannelLeave {
		if s.keepUsers.Has(b.OrganizationName(), m.Channel, m.User) {
			b.Client().InviteUser(m.Channel, m.User)
		}
		return nil
	}

	return nil
}
