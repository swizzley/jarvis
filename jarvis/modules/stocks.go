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
	stockInfo, err := external.StockPrice(tickers, external.StockDefaultFormat)
	if err != nil {
		return err
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
		if stock.Values != nil && len(stock.Values) > 3 {
			if floatValue, isFloat := stock.Values[2].(float64); isFloat {
				change := floatValue
				changePct := stock.Values[3]

				volume := int64(stock.Values[1].(float64))

				tickerText := fmt.Sprintf("`%s`", stock.Ticker)
				lastPriceText := fmt.Sprintf("%0.2f", stock.Values[0])
				volumeText := humanize.Comma(volume)
				changeText := fmt.Sprintf("%.2f", change)
				changePctText := util.StripQuotes(changePct.(string))

				var barColor = "#00FF00"
				if change < 0 {
					barColor = "#FF0000"
				}

				item := slack.ChatMessageAttachment{
					Color: slack.OptionalString(barColor),
					Fields: []slack.Field{
						slack.Field{Title: "Ticker", Value: tickerText, Short: true},
						slack.Field{Title: "Last", Value: lastPriceText, Short: true},
						slack.Field{Title: "Volume", Value: volumeText, Short: true},
						slack.Field{Title: "Change ∆", Value: changeText, Short: true},
						slack.Field{Title: "Change %", Value: changePctText, Short: true},
					},
				}

				message.Attachments = append(message.Attachments, item)
			}
		}
	}
	_, err := b.Client().ChatPostMessage(message)
	return err
}
