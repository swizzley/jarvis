package core

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-util/collections"
	"github.com/wcharczuk/go-slack"
)

// MessageHandler is a function that takes a slack message and acts on it.
type MessageHandler func(b Bot, m *slack.Message) error

// BotModule is a suite of actions (either Mention driven or Passive).
type BotModule interface {
	//Init is a function that lets modules perform any initialization they might need.
	Init(b Bot) error

	//Name is the name of the module.
	Name() string

	//Actions are the actions the module provides.
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

	LoadModule(moduleName string) error
	UnloadModule(moduleName string)
	RegisteredModules() collections.StringSet
	LoadedModules() collections.StringSet

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
