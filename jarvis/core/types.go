package core

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/go-slack"
)

const (
	// OptionPassive is the config entry that governs whether or not to process passive responses.
	OptionPassive = "option.passive"
)

// Action represents an action that can be handled by Jarvis for a given message pattern.
type Action struct {
	ID             string
	MessagePattern string
	Description    string
	Passive        bool
	Handler        MessageHandler
}

// MessageHandler is a function that takes a slack message and acts on it.
type MessageHandler func(b Bot, m *slack.Message) error

// BotModule is a suite of actions (either Mention driven or Passive).
type BotModule interface {
	Name() string
	Actions() []Action
}

// Bot interface is the interop interface used between modules.
type Bot interface {
	ID() string
	Token() string
	OrganizationName() string

	Configuration() map[string]string
	State() map[string]interface{}
	JobManager() *chronometer.JobManager

	Actions() []Action
	AddAction(action Action)
	RemoveAction(id string)
	TriggerAction(id string, m *slack.Message) error

	Client() *slack.Client
	ActiveChannels() []string

	FindUser(userID string) *slack.User
	FindChannel(channelID string) *slack.Channel

	Say(destinationID string, components ...interface{}) error
	Sayf(destinationID string, format string, components ...interface{}) error

	Log(components ...interface{})
	Logf(format string, components ...interface{})
}
