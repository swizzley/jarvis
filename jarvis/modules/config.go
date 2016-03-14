package modules

import (
	"fmt"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	// ConfigModules is the modules config entry.
	ConfigModules = "modules"

	// ModuleConfig is the name of the config module.
	ModuleConfig = "config"

	// ActionConfigSet is the set config action.
	ActionConfigSet = "config.set"

	// ActionConfigGet is the get config action.
	ActionConfigGet = "config.get"

	// ActionConfig is the list config values action.
	ActionConfig = "config"

	// ActionModuleLoad is the list config values action.
	ActionModuleLoad = "module.load"

	// ActionModuleUnload is the list config values action.
	ActionModuleUnload = "module.unload"

	// ActionModule is the list config values action.
	ActionModule = "module"
)

// Config is the module that governs configuration manipulation.
type Config struct{}

// Name returns the name for the module.
func (c *Config) Name() string {
	return ModuleConfig
}

// Actions returns the actions for the module.
func (c *Config) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionConfigSet, MessagePattern: "^config:(.+) (.+)", Description: "Set config values", Handler: c.handleConfigSet},
		core.Action{ID: ActionConfigGet, MessagePattern: "^config:(.+)", Description: "Get config values", Handler: c.handleConfigGet},
		core.Action{ID: ActionConfig, MessagePattern: "^config", Description: "Prints the current config", Handler: c.handleConfig},

		core.Action{ID: ActionModuleLoad, MessagePattern: "^module:load (.+)", Description: "Loads a module", Handler: c.handleLoadModule},
		core.Action{ID: ActionModuleUnload, MessagePattern: "^module:unload (.+)", Description: "Unloads a module", Handler: c.handleUnloadModule},
		core.Action{ID: ActionModule, MessagePattern: "^module", Description: "Prints the current loaded modules", Handler: c.handleModule},
	}
}

func (c *Config) handleConfigSet(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	parts := core.ExtractSubMatches(messageWithoutMentions, "^config:(.+) (.+)")

	if len(parts) < 3 {
		return exception.Newf("malformed message for `%s`", ActionConfigSet)
	}

	key := parts[1]
	value := parts[2]

	setting := value
	if core.LikeAny(value, "true", "yes", "on", "1") {
		setting = "true"
	} else if core.LikeAny(value, "false", "off", "0") {
		setting = "false"
	}
	b.Configuration()[key] = setting
	return b.Sayf(m.Channel, "> %s: `%s` = %s", ActionConfigSet, key, setting)
}

func (c *Config) handleConfigGet(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	parts := core.ExtractSubMatches(messageWithoutMentions, "^config:(.+)")

	if len(parts) < 2 {
		return exception.Newf("malformed message for `%s`", ActionConfigGet)
	}

	key := parts[1]
	value := b.Configuration()[key]
	return b.Sayf(m.Channel, "> %s: `%s` = %s", ActionConfigGet, key, value)
}

func (c *Config) handleConfig(b core.Bot, m *slack.Message) error {
	configText := "current config:\n"
	for key, value := range b.Configuration() {
		if strings.HasPrefix(key, "option.") {
			configText = configText + fmt.Sprintf("> `%s` = %s\n", key, value)
		}
	}

	return b.Say(m.Channel, configText)
}

func (c *Config) handleLoadModule(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	parts := core.ExtractSubMatches(messageWithoutMentions, "^module:load (.+)")
	if len(parts) < 2 {
		return exception.Newf("malformed message for `%s`", ActionModuleLoad)
	}

	key := parts[1]
	if b.LoadedModules().Contains(key) {
		return b.Sayf(m.Channel, "Module `%s` is already loaded.", key)
	}
	if !b.RegisteredModules().Contains(key) {
		return b.Sayf(m.Channel, "Module `%s` isn't registered.", key)
	}

	b.LoadModule(key)
	return b.Sayf(m.Channel, "Loaded Module `%s`.", key)
}

func (c *Config) handleUnloadModule(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	parts := core.ExtractSubMatches(messageWithoutMentions, "^module:unload (.+)")
	if len(parts) < 2 {
		return exception.Newf("malformed message for `%s`", ActionModuleUnload)
	}

	key := parts[1]
	if !b.LoadedModules().Contains(key) {
		return b.Sayf(m.Channel, "Module `%s` isn't loaded.", key)
	}
	if !b.RegisteredModules().Contains(key) {
		return b.Sayf(m.Channel, "Module `%s` isn't registered.", key)
	}

	b.UnloadModule(key)
	return b.Sayf(m.Channel, "Unloaded Module `%s`.", key)
}

func (c *Config) handleModule(b core.Bot, m *slack.Message) error {
	moduleText := "currently loaded modules:\n"
	for key := range b.LoadedModules() {
		moduleText = moduleText + fmt.Sprintf("> `%s`\n", key)
	}
	return b.Say(m.Channel, moduleText)
}
