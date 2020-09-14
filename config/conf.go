package config

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var (
	Token = flag.String("token", "", "your sandbox's token")
	Path  = flag.String("path", ".", "path to storage dir")

	TimeSuffixEnabled   = flag.Bool("time-suffix-enabled", false, "add the time suffix to every filename on (re)start")
	TimeSuffixFormat    = flag.String("time-suffix-format", "2006010215", "go format of the time suffix (see https://golang.org/src/time/format.go)")
	TimeSuffixStartedAt = time.Now().UTC()

	Orderbook      = flag.String("orderbook", "", "list of tickers to subscribe for orderbooks")
	OrderbookDepth = flag.Int("orderbook-depth", 20, "depth of orderbook: from 1 to 20")

	Candle         = flag.String("candle", "", "list of tickers to subscribe for candles")
	CandleInterval = flag.String("candle-interval", "1min", "interval of candles: 1min,2min,3min,5min,10min,15min,30min,hour,2hour,4hour,day,week,month")

	version       = flag.Bool("version", false, "show version info")
	VersionString string
)

func init() {
	flag.Parse()
	if *version {
		fmt.Printf("%s\n", VersionString)
		os.Exit(0)
	}

	rand.Seed(time.Now().UnixNano()) // for RequestID
}
