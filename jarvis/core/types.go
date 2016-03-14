package core

import (
	"fmt"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util/collections"
	"github.com/wcharczuk/go-slack"
)

const (
	// OptionPassive is the config entry that governs whether or not to process passive responses.
	OptionPassive = "option.passive"

	// PriorityHigh is for actions that have to be processed / checked first.
	PriorityHigh = 500

	// PriorityNormal is for typical actions (this is the default).
	PriorityNormal = 100

	// PriorityCatchAll is for actions that should be processed / checked last.
	// CatchAll actions typically have this priority.
	PriorityCatchAll = 1
)

// Action represents an action that can be handled by Jarvis for a given message pattern.
type Action struct {
	ID             string
	MessagePattern string
	Description    string
	Passive        bool
	Handler        MessageHandler
	Priority       int
}

// ActionsByPriority sorts an action slice by the priority desc.
type ActionsByPriority []Action

// Len returns the slice length.
func (a ActionsByPriority) Len() int {
	return len(a)
}

// Swap swaps two indexes.
func (a ActionsByPriority) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ActionsByPriority) Less(i, j int) bool {
	return a[i].Priority > a[j].Priority
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

	LoadModule(moduleName string)
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

// NewMockBot returns a new Bot instance.
func NewMockBot(token string) *MockBot {
	return &MockBot{id: slack.UUIDv4().ToShortString(), organizationName: "Test Organization", token: token, jobManager: chronometer.NewJobManager(), state: map[string]interface{}{}, configuration: map[string]string{OptionPassive: "false"}, actions: map[string]Action{}}
}

// MockMessage returns a mock message.
func MockMessage(messageText string) *slack.Message {
	return &slack.Message{Channel: "CTESTCHANNEL", Text: messageText}
}

// MockBot is a testing bot.
type MockBot struct {
	id               string
	token            string
	organizationName string
	configuration    map[string]string
	state            map[string]interface{}
	jobManager       *chronometer.JobManager
	actions          map[string]Action

	modules       map[string]BotModule
	loadedModules collections.StringSet

	mockMessageHandler MessageHandler
}

// MockMessageHandler sets a handler for any call to Say or Sayf
func (mb *MockBot) MockMessageHandler(handler MessageHandler) {
	mb.mockMessageHandler = handler
}

// ID returns the id.
func (mb *MockBot) ID() string {
	return mb.id
}

// Token returns the token.
func (mb *MockBot) Token() string {
	return mb.token
}

// OrganizationName returns the organizationName.
func (mb *MockBot) OrganizationName() string {
	return mb.organizationName
}

// Configuration returns configuration.
func (mb *MockBot) Configuration() map[string]string {
	return mb.configuration
}

// State returns state.
func (mb *MockBot) State() map[string]interface{} {
	return mb.state
}

// JobManager returns the jobManager.
func (mb *MockBot) JobManager() *chronometer.JobManager {
	return mb.jobManager
}

// Client returns a bare slack client.
func (mb *MockBot) Client() *slack.Client {
	return &slack.Client{ActiveChannels: []string{"CTESTCHANNEL"}}
}

// Actions returns the actions loaded for a bot
func (mb *MockBot) Actions() []Action {
	actions := []Action{}
	for _, action := range mb.actions {
		actions = append(actions, action)
	}

	return actions
}

// AddAction adds an action for the bot.
func (mb *MockBot) AddAction(action Action) {
	mb.actions[action.ID] = action
}

// RemoveAction removes an action from the bot.
func (mb *MockBot) RemoveAction(id string) {
	delete(mb.actions, id)
}

// TriggerAction triggers and action with a given message.
func (mb *MockBot) TriggerAction(id string, m *slack.Message) error {
	if action, hasAction := mb.actions[id]; hasAction {
		return action.Handler(mb, m)
	}
	return exception.Newf("action %s is not loaded.", id)
}

// ActiveChannels returns a list of active channel ids.
func (mb *MockBot) ActiveChannels() []string {
	return []string{"CTESTCHANNEL"}
}

// FindUser returns the user object for a given userID.
func (mb *MockBot) FindUser(userID string) *slack.User {
	return &slack.User{
		ID:   slack.UUIDv4().ToShortString(),
		Name: "test_user",
		Profile: &slack.UserProfile{
			FirstName: "Test",
			LastName:  "User",
			Email:     "test_user@test.com",
			RealName:  "Mr. Test User",
		},
	}
}

// FindChannel returns the channel object for a given channelID.
func (mb *MockBot) FindChannel(channelID string) *slack.Channel {
	return &slack.Channel{
		ID:   "CTESTCHANNEL",
		Name: "test-channel",
	}
}

// Say routes messages to a mock handler if there is one.
func (mb *MockBot) Say(destinationID string, components ...interface{}) error {
	messageText := fmt.Sprint(components...)
	mb.dispatchToMockHandler(MockMessage(messageText))
	return nil
}

// Sayf routes messages to a mock handler if there is one.
func (mb *MockBot) Sayf(destinationID, format string, components ...interface{}) error {
	messageText := fmt.Sprintf(format, components...)
	mb.dispatchToMockHandler(MockMessage(messageText))
	return nil
}

// Log writes to the log.
func (mb *MockBot) Log(components ...interface{}) {}

// Logf writes to the log in a given format.
func (mb *MockBot) Logf(format string, components ...interface{}) {}

func (mb *MockBot) dispatchToMockHandler(m *slack.Message) {
	if mb.mockMessageHandler != nil {
		mb.mockMessageHandler(mb, m)
	}
}

// RegisterModule loads a given bot module
func (mb *MockBot) RegisterModule(m BotModule) {
	mb.modules[m.Name()] = m
}

// LoadModule loads a registered module.
func (mb *MockBot) LoadModule(moduleName string) {
	if _, hasModule := mb.modules[moduleName]; hasModule {
		mb.loadedModules.Add(moduleName)
	}
}

// UnloadModule unloads a module and its actions.
func (mb *MockBot) UnloadModule(moduleName string) {
	if _, hasModule := mb.modules[moduleName]; hasModule {
		mb.loadedModules.Remove(moduleName)
	}
}

// LoadedModules returns the currently loaded modules.
func (mb *MockBot) LoadedModules() collections.StringSet {
	return mb.loadedModules
}

// RegisteredModules returns the registered modules.
func (mb *MockBot) RegisteredModules() collections.StringSet {
	registered := collections.StringSet{}
	for key := range mb.modules {
		registered.Add(key)
	}

	return registered
}
