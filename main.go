package main

import (
	"flag"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	token = flag.String("token", "", "your sandbox's token")
	path  = flag.String("path", ".", "path to storage dir")

	timeSuffixEnabled   = flag.Bool("time-suffix-enabled", false, "add the time suffix to every filename on (re)start")
	timeSuffixFormat    = flag.String("time-suffix-format", "2006010215", "go format of the time suffix (see https://golang.org/src/time/format.go)")
	timeSuffixStartedAt = time.Now().UTC()

	orderbook      = flag.String("orderbook", "", "list of tickers to subscribe for orderbooks")
	orderbookDepth = flag.Int("orderbook-depth", 20, "depth of orderbook: from 1 to 20")

	candle         = flag.String("candle", "", "list of tickers to subscribe for candles")
	candleInterval = flag.String("candle-interval", "1min", "interval of candles: 1min,2min,3min,5min,10min,15min,30min,hour,2hour,4hour,day,week,month")

	version       = flag.Bool("version", false, "show version info")
	VersionString string
)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("%s\n", VersionString)
		os.Exit(0)
	}

	rand.Seed(time.Now().UnixNano()) // for RequestID
	logger := log.New(os.Stdout, "", log.LstdFlags)

	sandboxRestClient := sdk.NewSandboxRestClient(*token)
	streamingClient, err := sdk.NewStreamingClient(logger, *token)
	if err != nil {
		logger.Fatalln(err)
	}
	defer streamingClient.Close()

	scope, err := NewMainScope(sandboxRestClient, parseTickersList(*orderbook), parseTickersList(*candle), logger)
	if err != nil {
		logger.Fatalln(err)
	}
	scope.initChannels()
	scope.initDiskWriters()

	go scope.eventReceiver(streamingClient)

	scope.subscribeOrderbook(streamingClient)
	scope.subscribeCandles(streamingClient)
	defer scope.unsubscribeOrderbook(streamingClient)
	defer scope.unsubscribeCandles(streamingClient)

	select {} // sleep(0), epta
}
