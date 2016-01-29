package lib

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
)

type StockInfo struct {
	Ticker string
	Format string
	Values []interface{}
}

const (
	STOCK_DEFAULT_FORMAT = "l1vc1p2"
)

func StockPrice(tickers []string, format string) ([]StockInfo, error) {
	if len(tickers) == 0 {
		return []StockInfo{}, nil
	}

	rawResults, meta, resErr := request.NewRequest().AsGet().
		WithUrl("http://finance.yahoo.com/d/quotes.csv").
		WithQueryString("s", strings.Join(tickers, "+")).
		WithQueryString("f", format).
		FetchStringWithMeta()

	if resErr != nil {
		return []StockInfo{}, resErr
	}

	if meta.StatusCode != http.StatusOK {
		return []StockInfo{}, exception.New("Non (200) response from pricing provider.")
	}

	results := []StockInfo{}

	scanner := bufio.NewScanner(strings.NewReader(rawResults))
	scanner.Split(bufio.ScanLines)

	index := 0
	for scanner.Scan() {
		si := StockInfo{}
		si.Format = format

		line := scanner.Text()
		parts := strings.Split(line, ",")

		si.Ticker = tickers[index]

		values := []interface{}{}
		for _, v := range parts {
			f, fErr := strconv.ParseFloat(v, 64)
			if fErr == nil {
				values = append(values, f)
			} else {
				values = append(values, v)
			}
		}
		si.Values = values
		results = append(results, si)
		index++
	}
	return results, nil
}

func AnnounceStocks(c *slack.Client, destinationId string, stockInfo []StockInfo) error {
	tickersLabels := []string{}
	for _, stock := range stockInfo {
		tickersLabels = append(tickersLabels, fmt.Sprintf("`%s`", stock.Ticker))
	}
	tickersLabel := strings.Join(tickersLabels, " ")
	stockText := fmt.Sprintf("current equity price info for %s\n", tickersLabel)
	for _, stock := range stockInfo {
		if stock.Values != nil && len(stock.Values) > 3 {
			change := stock.Values[2].(float64)
			changeText := ""
			if change > 0 {
				changeText = fmt.Sprintf("+%.2f", change)
			} else if change < 0 {
				changeText = fmt.Sprintf("-%.2f", change)
			} else {
				changeText = "0.00"
			}

			changePct := stock.Values[3]
			stockText = stockText + fmt.Sprintf("> `%s` - last: *%.2f* vol: *%.2f* ch: *%s* *%s*\n", stock.Ticker, stock.Values[0], stock.Values[1], changeText, util.StripQuotes(changePct.(string)))
		}
	}
	return c.Say(destinationId, stockText)
}

func AnnounceTime(c *slack.Client, channelId string, currentTime time.Time) error {
	timeText := fmt.Sprintf("%s UTC", currentTime.Format(time.Kitchen))
	message := slack.NewChatMessage(channelId, "")
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	message.UnfurlMedia = slack.OptionalBool(false)
	message.Attachments = []slack.ChatMessageAttachment{
		slack.ChatMessageAttachment{
			Fallback: fmt.Sprintf("The time is now:\n>%s", timeText),
			Color:    slack.OptionalString("#4099FF"),
			Pretext:  slack.OptionalString("The time is now:"),
			Text:     slack.OptionalString(timeText),
		},
	}

	_, messageErr := c.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}
