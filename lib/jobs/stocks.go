package jobs

import (
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/lib"
)

func NewStock(c *slack.Client) *Stocks {
	job := &Stocks{}
	job.Client = c
	job.Tickers = []string{}

	tickers := []lib.TrackedStock{}
	spiffy.DefaultDb().GetAll(&tickers)
	for _, ticker := range tickers {
		job.Tickers = append(job.Tickers, ticker.Ticker)
	}
	return job
}

type Stocks struct {
	Client  *slack.Client
	Tickers []string
}

func (t *Stocks) Track(ticker, createdBy string) {
	t.Tickers = append(t.Tickers, ticker)

	item := lib.TrackedStock{}
	item.Ticker = ticker
	item.CreatedBy = createdBy
	item.TimestampUTC = time.Now().UTC()
	spiffy.DefaultDb().Create(&item)
}

func (t *Stocks) StopTracking(ticker string) {
	newTickers := []string{}
	for _, v := range t.Tickers {
		if v != ticker {
			newTickers = append(newTickers, v)
		}
	}
	t.Tickers = newTickers

	spiffy.DefaultDb().Exec("DELETE FROM tracked_stock where ticker = $1", ticker)
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
