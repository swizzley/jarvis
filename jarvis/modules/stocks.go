package modules

import (
	"fmt"
	"strings"

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
type Stocks struct {
}

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
	stockInfo, stockErr := external.StockPrice(tickers, external.StockDefaultFormat)
	if stockErr != nil {
		return stockErr
	}
	return s.announceStocks(b, m.Channel, stockInfo)
}

func (s *Stocks) announceStocks(b core.Bot, destinationID string, stockInfo []external.StockInfo) error {
	tickersLabels := []string{}
	for _, stock := range stockInfo {
		tickersLabels = append(tickersLabels, fmt.Sprintf("`%s`", stock.Ticker))
	}
	tickersLabel := strings.Join(tickersLabels, " ")
	stockText := fmt.Sprintf("current equity price info for %s\n", tickersLabel)
	for _, stock := range stockInfo {
		if stock.Values != nil && len(stock.Values) > 3 {
			if floatValue, isFloat := stock.Values[2].(float64); isFloat {
				change := floatValue
				changeText := fmt.Sprintf("%.2f", change)
				changePct := stock.Values[3]
				stockText = stockText + fmt.Sprintf("> `%s` - last: *%.2f* vol: *%d* ch: *%s* *%s*\n", stock.Ticker, stock.Values[0], int(stock.Values[1].(float64)), changeText, util.StripQuotes(changePct.(string)))
			} else {
				return exception.Newf("There was an issue with `%s`", stock.Ticker)
			}
		}
	}
	return b.Say(destinationID, stockText)
}
