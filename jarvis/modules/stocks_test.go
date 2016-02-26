package modules

import (
	"fmt"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/external"
)

func TestHandleStocks(t *testing.T) {
	assert := assert.New(t)
	s := &Stocks{}
	mb := core.NewMockBot(slack.UUIDv4().ToShortString())

	core.MockResponseFromString("GET", fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?f=%s&s=%s", external.StockDefaultFormat, "goog"), 200, `705.75,1642166,+6.19,"+0.88%"`)

	mb.Configuration()["foo"] = "bar"
	handleErr := s.handleStockPrice(mb, core.MockMessage("stock:price goog"))
	assert.Nil(handleErr)
}
