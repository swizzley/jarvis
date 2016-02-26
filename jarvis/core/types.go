package core

import (
	"github.com/wcharczuk/go-slack"
)

// Action represents an action that can be handled by Jarvis for a given message pattern.
type Action struct {
	ID             string
	MessagePattern string
	Description    string
	Handler        MessageHandler
}

// MessageHandler is a function that takes a slack message and acts on it.
type MessageHandler func(b Bot, m *slack.Message) error

// BotModule is a suite of actions (either Mention driven or Passive).
type BotModule interface {
	Name() string
	MentionCommands() []Action
	PassiveCommands() []Action
}

// Bot interface is the interop interface used between modules.
type Bot interface {
	Token() string

	MentionCommands() []Action
	PassiveCommands() []Action

	TriggerMentionCommand(id string, m *slack.Message) error
	TriggerPassiveCommand(id string, m *slack.Message) error

	GetClient() *slack.Client
	GetActiveChannels() []slack.Channel

	FindUser(userID string) *slack.User
	FindChannel(channelID string) *slack.Channel

	Say(destinationID string, components ...interface{}) error
	Sayf(destinationID string, format string, components ...interface{}) error

	Log(components ...interface{})
	Logf(format string, components ...interface{})
}
