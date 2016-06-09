package jarvis

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/collections"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/modules"
)

const (
	// EnvironmentSlackAPIToken is the slack api token environment variable.
	EnvironmentSlackAPIToken = "SLACK_API_TOKEN"
)

// NewBotFromEnvironment creates a new bot from environment variables.
func NewBotFromEnvironment() *Bot {
	envToken := os.Getenv(EnvironmentSlackAPIToken)
	if len(envToken) == 0 {
		return nil
	}
	b := NewBot(envToken)

	envModules := os.Getenv(modules.EnvironmentModules)
	if len(envModules) != 0 {
		b.Configuration()[modules.ConfigModules] = envModules
	} else {
		b.Configuration()[modules.ConfigModules] = "all"
	}
	return b
}

// NewBot returns a new Bot instance.
func NewBot(token string) *Bot {
	return &Bot{
		token:          token,
		jobManager:     chronometer.NewJobManager(),
		state:          map[string]interface{}{},
		configuration:  map[string]string{},
		actionLookup:   map[string]core.Action{},
		modules:        map[string]core.BotModule{},
		loadedModules:  collections.SetOfString{},
		mentionActions: []core.Action{},
		passiveActions: []core.Action{},
	}
}

// Bot is the main primitive.
type Bot struct {
	id    string
	token string

	organizationName string
	configuration    map[string]string
	state            map[string]interface{}
	jobManager       *chronometer.JobManager
	client           *slack.Client

	modules       map[string]core.BotModule
	loadedModules collections.SetOfString

	mentionActions []core.Action
	passiveActions []core.Action
	actionLookup   map[string]core.Action
	UsersLookup    map[string]slack.User
	ChannelsLookup map[string]slack.Channel
}

// ID returns the id.
func (b *Bot) ID() string {
	return b.id
}

// Token returns the token.
func (b *Bot) Token() string {
	return b.token
}

// OrganizationName returns the organization name.
func (b *Bot) OrganizationName() string {
	return b.organizationName
}

// JobManager returns the job manager.
func (b *Bot) JobManager() *chronometer.JobManager {
	return b.jobManager
}

//Configuration returns the current bot configuration.
func (b *Bot) Configuration() map[string]string {
	return b.configuration
}

//State returns the current bot state.
func (b *Bot) State() map[string]interface{} {
	return b.state
}

// Client returns the Slack client.
func (b *Bot) Client() *slack.Client {
	return b.client
}

// Actions returns the actions loaded for a bot
func (b *Bot) Actions() []core.Action {
	allActions := []core.Action{}
	allActions = append(allActions, b.mentionActions...)
	allActions = append(allActions, b.passiveActions...)
	sort.Sort(core.ActionsByPriority(allActions))
	return allActions
}

// AddAction adds an action for the bot.
func (b *Bot) AddAction(action core.Action) {
	if action.Priority == 0 {
		action.Priority = core.PriorityNormal
	}
	if action.Passive {
		b.passiveActions = append(b.passiveActions, action)

		sortable := core.ActionsByPriority(b.passiveActions)
		sort.Sort(sortable)
		b.passiveActions = sortable
	} else {
		b.mentionActions = append(b.mentionActions, action)

		sortable := core.ActionsByPriority(b.mentionActions)
		sort.Sort(sortable)
		b.mentionActions = sortable
	}
	b.actionLookup[action.ID] = action
}

// RemoveAction removes an action from the bot.
func (b *Bot) RemoveAction(id string) {
	action, hasAction := b.actionLookup[id]
	if !hasAction {
		return
	}

	if action.Passive {
		b.passiveActions = filterActions(b.passiveActions, id)
	} else {
		b.mentionActions = filterActions(b.mentionActions, id)
	}
	delete(b.actionLookup, id)
}

func filterActions(actions []core.Action, id string) []core.Action {
	newActions := []core.Action{}
	for _, action := range actions {
		if action.ID != id {
			newActions = append(newActions, action)
		}
	}
	return newActions
}

// TriggerAction triggers and action with a given message.
func (b *Bot) TriggerAction(id string, m *slack.Message) error {
	if action, hasAction := b.actionLookup[id]; hasAction {
		return action.Handler(b, m)
	}
	return exception.Newf("action %s is not loaded.", id)
}

// ActiveChannels returns a list of active channel ids.
func (b *Bot) ActiveChannels() []string {
	return b.client.ActiveChannels
}

// RegisterModule loads a given bot module
func (b *Bot) RegisterModule(m core.BotModule) {
	b.modules[m.Name()] = m
}

// LoadModule loads a registered module.
func (b *Bot) LoadModule(moduleName string) error {
	var err error
	var actions []core.Action
	if m, hasModule := b.modules[moduleName]; hasModule {
		err = m.Init(b)
		if err != nil {
			return err
		}

		actions = m.Actions()
		for _, action := range actions {
			b.AddAction(action)
		}
		b.loadedModules.Add(moduleName)
	}
	return nil
}

// UnloadModule unloads a module and its actions.
func (b *Bot) UnloadModule(moduleName string) {
	if m, hasModule := b.modules[moduleName]; hasModule {
		actions := m.Actions()
		for _, action := range actions {
			b.RemoveAction(action.ID)
		}
		b.loadedModules.Remove(moduleName)
	}
}

// LoadedModules returns the currently loaded modules.
func (b *Bot) LoadedModules() collections.SetOfString {
	return b.loadedModules
}

// RegisteredModules returns the registered modules.
func (b *Bot) RegisteredModules() collections.SetOfString {
	registered := collections.SetOfString{}
	for key := range b.modules {
		registered.Add(key)
	}

	return registered
}

func (b *Bot) loadAllRegisteredModules() {
	for name := range b.modules {
		loadErr := b.LoadModule(name)
		if loadErr != nil {
			b.Logf("Error loading module `%s`: %v", name, loadErr)
		}
	}
}

func (b *Bot) loadConfiguredModules() {
	configEntry, hasEntry := b.configuration[modules.ConfigModules]
	if !hasEntry || strings.ToLower(configEntry) == "all" {
		b.loadAllRegisteredModules()
		return
	}

	moduleNames := strings.Split(configEntry, ",")
	for _, name := range moduleNames {
		nameLower := strings.ToLower(name)
		loadErr := b.LoadModule(nameLower)
		if loadErr != nil {
			b.Logf("Error loading module `%s`: %v", name, loadErr)
		}
	}
}

// Init connects the bot to Slack.
func (b *Bot) Init() error {

	b.RegisterModule(&modules.ConsoleRunner{})
	b.RegisterModule(&modules.Jira{})
	b.RegisterModule(&modules.Stocks{})
	b.RegisterModule(&modules.Jobs{})
	b.RegisterModule(&modules.Config{})
	b.RegisterModule(&modules.Util{})
	b.RegisterModule(&modules.Core{})

	b.loadConfiguredModules()

	client := slack.NewClient(b.token)
	client.SetDebug(true)
	b.client = client
	b.client.AddEventListener(slack.EventHello, func(c *slack.Client, m *slack.Message) {
		b.Log("slack is connected")
	})
	b.client.AddEventListener(slack.EventPing, func(c *slack.Client, m *slack.Message) {
		b.Log("ping!")
	})
	b.client.AddEventListener(slack.EventPong, func(c *slack.Client, m *slack.Message) {
		b.Log("pong!")
	})
	b.client.AddEventListener(slack.EventMessage, func(c *slack.Client, m *slack.Message) {
		resErr := b.dispatchResponse(m)
		if resErr != nil {
			c.Sayf(m.Channel, "there was an error handling the message:\n> %s", resErr.Error())
			b.Log(resErr)
		}
	})

	return nil
}

// Start starts the bot and connects to Slack.
func (b *Bot) Start() error {
	session, err := b.client.Connect()
	if err != nil {
		return err
	}

	b.id = session.Self.ID
	b.organizationName = session.Team.Name
	b.ChannelsLookup = b.createChannelLookup(session)
	b.UsersLookup = b.createUsersLookup(session)
	b.jobManager.Start()
	return nil
}

func (b *Bot) passivesEnabled() bool {
	if value, hasKey := b.configuration[modules.ConfigOptionPassive]; hasKey {
		return strings.ToLower(value) == "true"
	}
	return false
}

func (b *Bot) dispatchResponse(m *slack.Message) error {
	defer func() {
		if r := recover(); r != nil {
			b.Sayf(m.Channel, "there was a panic handling the message:\n> %v", r)
		}
	}()

	//b.LogIncomingMessage(m)
	user := b.FindUser(m.User)
	if user != nil {
		if m.User != "slackbot" && m.User != b.id && !user.IsBot {
			messageText := util.TrimWhitespace(core.LessMentions(m.Text))
			if core.IsUserMention(m.Text, b.id) || core.IsDM(m.Channel) {
				for _, action := range b.mentionActions {
					if core.Like(messageText, action.MessagePattern) && len(action.MessagePattern) != 0 {
						return action.Handler(b, m)
					}
				}
			}
			if b.passivesEnabled() {
				for _, action := range b.passiveActions {
					if core.Like(messageText, action.MessagePattern) && len(action.MessagePattern) != 0 {
						return action.Handler(b, m)
					}
				}
			}
		}
	}
	return nil
}

// FindUser returns the user object for a given userID.
func (b *Bot) FindUser(userID string) *slack.User {
	if user, hasUser := b.UsersLookup[userID]; hasUser {
		return &user
	}
	return nil
}

// FindChannel returns the channel object for a given channelID.
func (b *Bot) FindChannel(channelID string) *slack.Channel {
	if channel, hasChannel := b.ChannelsLookup[channelID]; hasChannel {
		return &channel
	}
	return nil
}

func (b *Bot) createUsersLookup(session *slack.Session) map[string]slack.User {
	lookup := map[string]slack.User{}
	for x := 0; x < len(session.Users); x++ {
		user := session.Users[x]
		lookup[user.ID] = user
	}
	return lookup
}

func (b *Bot) createChannelLookup(session *slack.Session) map[string]slack.Channel {
	lookup := map[string]slack.Channel{}
	for x := 0; x < len(session.Channels); x++ {
		channel := session.Channels[x]
		lookup[channel.ID] = channel
	}
	return lookup
}

// Say calls the internal slack.Client.Say method.
func (b *Bot) Say(destinationID string, components ...interface{}) error {
	//b.LogOutgoingMessage(destinationID, components...)
	return b.client.Say(destinationID, components...)
}

// Sayf calls the internal slack.Client.Sayf method.
func (b *Bot) Sayf(destinationID string, format string, components ...interface{}) error {
	//message := fmt.Sprintf(format, components...)
	//b.LogOutgoingMessage(destinationID, message)
	return b.client.Sayf(destinationID, format, components...)
}

// LogIncomingMessage writes an incoming message to the log.
func (b *Bot) LogIncomingMessage(m *slack.Message) {
	user := b.FindUser(m.User)
	channel := b.FindChannel(m.Channel)

	userName := "system"
	if user != nil {
		userName = user.Name
	}

	if channel != nil {
		b.Logf("=> #%s (%s) - %s: %s", channel.Name, channel.ID, userName, m.Text)
	} else {
		b.Logf("=> PM - %s: %s", userName, m.Text)
	}
}

// LogOutgoingMessage logs an outgoing message.
func (b *Bot) LogOutgoingMessage(destinationID string, components ...interface{}) {
	if core.Like(destinationID, "^C") {
		channel := b.FindChannel(destinationID)
		b.Logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.ID, fmt.Sprint(components...))
	} else {
		b.Logf("<= PM - jarvis: %s", fmt.Sprint(components...))
	}
}

// Log writes to the log.
func (b *Bot) Log(components ...interface{}) {
	message := fmt.Sprint(components...)
	fmt.Printf("%s - %s - %s\n", b.OrganizationName(), time.Now().UTC().Format(time.RFC3339), message)
}

// Logf writes to the log in a given format.
func (b *Bot) Logf(format string, components ...interface{}) {
	message := fmt.Sprintf(format, components...)
	fmt.Printf("%s - %s - %s\n", b.OrganizationName(), time.Now().UTC().Format(time.RFC3339), message)
}
