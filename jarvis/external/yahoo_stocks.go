package external

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

// StockInfo represents information about a stock.
type StockInfo struct {
	Ticker string
	Format string
	Values []interface{}
}

const (
	// StockDefaultFormat is the default stock format string for Yahoo's stock api.
	StockDefaultFormat = "l1vc1p2"
)

// StockPrice returns stock price info from Yahoo for the given tickers.
func StockPrice(tickers []string, format string) ([]StockInfo, error) {
	if len(tickers) == 0 {
		return []StockInfo{}, nil
	}

	rawResults, meta, resErr := core.NewExternalRequest().AsGet().
		WithURL("http://download.finance.yahoo.com/d/quotes.csv").
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
