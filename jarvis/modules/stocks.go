package modules

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/external"
)

const (
	// ModuleStocks is the name of the stocks module.
	ModuleStocks = "stocks"

	// ActionStockPrice is the action that queries a stocks price.
	ActionStockPrice = "stock.price"

	// ActionStockChart is the action that queries a stocks historical chart.
	ActionStockChart = "stock.chart"
)

// Stocks is the module that does stocks things.
type Stocks struct{}

// Init does nothing right now.
func (s *Stocks) Init(b core.Bot) error { return nil }

// Name returns the name of the stocks module.
func (s *Stocks) Name() string {
	return ModuleStocks
}

// Actions returns the actions for the module.
func (s *Stocks) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionStockPrice, MessagePattern: "^stock:price", Description: "Fetches the current price and volume for a given ticker.", Handler: s.handleStockPrice},
		core.Action{ID: ActionStockChart, MessagePattern: "^stock:chart", Description: "Fetches the current price chart for a given ticker.", Handler: s.handleStockChart},
	}
}

func (s *Stocks) handleStockPrice(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	pieces := core.ExtractSubMatches(messageWithoutMentions, "^stock:price (.*)")

	if len(pieces) < 2 {
		return exception.Newf("invalid input for %s", ActionStockPrice)
	}

	rawTicker := pieces[1]
	tickers := []string{}
	if strings.Contains(rawTicker, ",") {
		tickers = strings.Split(rawTicker, ",")
	} else {
		tickers = []string{rawTicker}
	}
	stockInfo, err := external.StockPrice(tickers)
	if err != nil {
		return err
	}
	if len(stockInfo) == 0 {
		return b.Sayf(m.Channel, "No stock information returned for: `%s`", strings.Join(tickers, ", "))
	}
	return s.announceStocks(b, m.Channel, stockInfo)
}

func (s *Stocks) announceStocks(b core.Bot, destinationID string, stockInfo []external.StockInfo) error {
	tickersLabels := []string{}
	for _, stock := range stockInfo {
		tickersLabels = append(tickersLabels, fmt.Sprintf("`%s`", stock.Ticker))
	}
	tickersLabel := strings.Join(tickersLabels, " ")
	leadText := fmt.Sprintf("current equity price info for %s", tickersLabel)
	message := slack.NewChatMessage(destinationID, leadText)
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	message.Parse = util.OptionalString("full")
	for _, stock := range stockInfo {
		change := stock.Change
		changePct := stock.ChangePercent
		volume := stock.Volume
		tickerText := fmt.Sprintf("`%s`", stock.Ticker)
		nameText := fmt.Sprintf("%s", stock.Name)
		lastPriceText := fmt.Sprintf("%0.2f USD", stock.LastPrice)
		volumeText := humanize.Comma(volume)
		changeText := fmt.Sprintf("%.2f USD", change)
		changePctText := util.StripQuotes(changePct)

		var barColor = "#00FF00"
		if change < 0 {
			barColor = "#FF0000"
		}

		item := slack.ChatMessageAttachment{
			Color: slack.OptionalString(barColor),
			Fields: []slack.Field{
				slack.Field{Title: "Ticker", Value: tickerText, Short: true},
				slack.Field{Title: "Name", Value: nameText, Short: true},
				slack.Field{Title: "Last", Value: lastPriceText, Short: true},
				slack.Field{Title: "Volume", Value: volumeText, Short: true},
				slack.Field{Title: "Change ∆", Value: changeText, Short: true},
				slack.Field{Title: "Change %", Value: changePctText, Short: true},
			},
		}

		message.Attachments = append(message.Attachments, item)
	}
	_, err := b.Client().ChatPostMessage(message)
	return err
}

func (s *Stocks) handleStockChart(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	args := core.ExtractSubMatches(messageWithoutMentions, "^stock:chart (.*)")

	if len(args) < 2 {
		return exception.Newf("invalid input for %s", ActionStockPrice)
	}

	pieces := strings.Split(args[1], " ")
	ticker := pieces[0]
	timeframe := "1M"
	if len(pieces) > 1 {
		timeframe = pieces[1]
	}
	imageURL := fmt.Sprintf("https://chart-service.charczuk.com/stock/chart/%s/%s?width=400&height=150&format=png", ticker, timeframe)

	leadText := fmt.Sprintf("Historical Chart for `%s`", ticker)
	message := slack.NewChatMessage(m.Channel, leadText)
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	message.Parse = util.OptionalString("full")
	message.Attachments = []slack.ChatMessageAttachment{
		slack.ChatMessageAttachment{
			Title:    util.OptionalString("Chart"),
			ImageURL: util.OptionalString(imageURL),
		},
	}
	_, err := b.Client().ChatPostMessage(message)
	return err
}
