package jobs

import (
	"fmt"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/lib"
)

func marketStartUtc(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, time.UTC)
}

func marketEndUtc(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, time.UTC)
}

type MarketHours struct{}

func (o MarketHours) GetNextRunTime(after *time.Time) time.Time {
	var returnValue time.Time
	if after == nil {
		now := time.Now().UTC()
		marketStart := marketStartUtc(now)
		marketEnd := marketEndUtc(now)
		if now.After(marketStart) && now.Before(marketEnd) {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(1 * time.Hour)
		} else if now.Before(marketStart) {
			returnValue = marketStart
		} else if now.After(marketEnd) {
			returnValue = marketStart.AddDate(0, 0, 1)
		}
	} else {
		marketStart := marketStartUtc(*after)
		marketEnd := marketEndUtc(*after)
		if after.After(marketStart) && after.Before(marketEnd) {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 0, 0, 0, time.UTC).Add(1 * time.Hour)
		} else if after.Before(marketStart) {
			returnValue = marketStart
		} else if after.After(marketEnd) {
			returnValue = marketStart.AddDate(0, 0, 1)
		}
	}
	return returnValue
}

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
	if len(t.Tickers) == 0 {
		return nil
	}

	stockInfo, stockErr := lib.StockPrice(t.Tickers, lib.STOCK_DEFAULT_FORMAT)
	if stockErr != nil {
		return stockErr
	}

	for x := 0; x < len(t.Client.ActiveChannels); x++ {
		channelId := t.Client.ActiveChannels[x]
		announceErr := lib.AnnounceStocks(t.Client, channelId, stockInfo)
		if announceErr != nil {
			fmt.Printf("%s - error announcing stocks: %v\n", time.Now().UTC(), announceErr)
		}
	}
	return nil
}

func (t Stocks) Schedule() chronometer.Schedule {
	return MarketHours{}
}
