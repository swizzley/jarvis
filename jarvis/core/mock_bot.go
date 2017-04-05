package core

// NewMockBot returns a new Bot instance.
import (
	"fmt"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/collections"
	"github.com/wcharczuk/go-slack"
)

// NewMockBot creates a new mock bot.
func NewMockBot(token string) *MockBot {
	return &MockBot{
		id:               slack.UUIDv4().ToShortString(),
		organizationName: "Test Organization",
		token:            token,
		jobManager:       chronometer.NewJobManager(),
		state:            map[string]interface{}{},
		configuration:    map[string]string{"option.passive": "false"},
		actions:          map[string]Action{},
		agent:            logger.New(logger.NewEventFlagSetNone())}
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

	agent         *logger.Agent
	modules       map[string]BotModule
	loadedModules collections.SetOfString

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
func (mb *MockBot) LoadModule(moduleName string) error {
	if _, hasModule := mb.modules[moduleName]; hasModule {
		mb.loadedModules.Add(moduleName)
	}
	return nil
}

// UnloadModule unloads a module and its actions.
func (mb *MockBot) UnloadModule(moduleName string) {
	if _, hasModule := mb.modules[moduleName]; hasModule {
		mb.loadedModules.Remove(moduleName)
	}
}

// LoadedModules returns the currently loaded modules.
func (mb *MockBot) LoadedModules() collections.SetOfString {
	return mb.loadedModules
}

// RegisteredModules returns the registered modules.
func (mb *MockBot) RegisteredModules() collections.SetOfString {
	registered := collections.SetOfString{}
	for key := range mb.modules {
		registered.Add(key)
	}

	return registered
}

// Logger returns the logger agent.
func (mb *MockBot) Logger() *logger.Agent {
	return mb.agent
}
