package modules

import (
	"fmt"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

const (
	// ModuleConfig is the name of the config module.
	ModuleConfig = "config"

	// ActionConfigSet is the set config action.
	ActionConfigSet = "config.set"

	// ActionConfigGet is the get config action.
	ActionConfigGet = "config.get"

	//ActionConfig is the list config values action.
	ActionConfig = "config"
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
	return b.Sayf(m.Channel, "> %s: `%s` = %s", ActionConfigSet, key, value)
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
		configText = configText + fmt.Sprintf("> `%s` = %s\n", key, value)
	}

	return b.Say(m.Channel, configText)
}
