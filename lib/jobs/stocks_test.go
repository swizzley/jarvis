package jobs

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
)

func TestMarketHours(t *testing.T) {
	a := assert.New(t)

	before := time.Date(2016, 1, 29, 12, 0, 0, 0, time.UTC)
	during := time.Date(2016, 1, 29, 16, 0, 0, 0, time.UTC)
	after := time.Date(2016, 1, 29, 22, 0, 0, 0, time.UTC)

	marketStart := marketStartUtc(before)
	marketStartTomorrow := marketStart.AddDate(0, 0, 1)

	s := MarketHours{}

	shouldBeMarketStart := s.GetNextRunTime(&before)
	a.InTimeDelta(marketStart, shouldBeMarketStart, 1*time.Second)

	shouldBeDuring := s.GetNextRunTime(&during)
	a.InTimeDelta(during, shouldBeDuring, 1*time.Hour)

	shouldBeMarketStartTomorrow := s.GetNextRunTime(&after)
	a.InTimeDelta(marketStartTomorrow, shouldBeMarketStartTomorrow, 1*time.Second)
}
