package modules

import (
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/external"
)

const (
	// ModuleGifs is the gifs module.
	ModuleGifs = "gifs"

	// ActionGifsSearch is the gifs search action.
	ActionGifsSearch = "gifs.search"
)

// Gifs is the google image search for gifs module
type Gifs struct{}

// Name returns the module name.
func (g *Gifs) Name() string {
	return ModuleGifs
}

// Actions returns the module actions.
func (g *Gifs) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionGifsSearch, MessagePattern: "^gif(s?)", Description: "Searches google images for a given pattern.", Handler: g.handleGifsSearch},
	}
}

func (g *Gifs) handleGifsSearch(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	pieces := core.ExtractSubMatches(messageWithoutMentions, "^gif(s?) (.*)")

	if len(pieces) < 2 {
		return exception.Newf("invalid input for %s", ActionGifsSearch)
	}

	query := pieces[1]
	images, imagesErr := external.GoogleImageSearch(query)

	if imagesErr != nil {
		return imagesErr
	}

	if len(images) == 0 {
		return b.Sayf(m.Channel, "No image results for `%s`", query)
	}

	return nil
}
