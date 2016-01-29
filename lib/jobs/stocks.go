package jobs

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/lib"
)

type Stocks struct {
	Client  *slack.Client
	Tickers []string
}

func (t *Stocks) Track(ticker string) {
	t.Tickers = append(t.Tickers, ticker)
}

func (t *Stocks) StopTracking(ticker string) {
	newTickers := []string{}
	for _, v := range t.Tickers {
		if v != ticker {
			newTickers = append(newTickers, v)
		}
	}
	t.Tickers = newTickers
}

func (t Stocks) Name() string {
	return "stocks"
}

func (t Stocks) Execute(ct *chronometer.CancellationToken) error {
	stockInfo, stockErr := lib.StockPrice(t.Tickers, lib.STOCK_DEFAULT_FORMAT)
	if stockErr != nil {
		return stockErr
	}

	for x := 0; x < len(t.Client.ActiveChannels); x++ {
		channelId := t.Client.ActiveChannels[x]
		return lib.AnnounceStocks(t.Client, channelId, stockInfo)
	}
	return nil
}

func (t Stocks) Schedule() chronometer.Schedule {
	return OnTheHour{}
}
