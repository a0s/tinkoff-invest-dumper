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

	orderbook      = flag.String("orderbook", "", "list of tickers to subscribe for orderbooks")
	orderbookDepth = flag.Int("orderbook-depth", 1, "depth of orderbook")

	candle         = flag.String("candle", "", "list of tickers to subscribe for candles")
	candleInterval = flag.String("candleInterval", "1min", "interval of candles: 1min,2min,3min,5min,10min,15min,30min,hour,2hour,4hour,day,week,month")

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

	scope := NewMainScope(listToTickers(*orderbook), listToTickers(*candle), logger)
	scope.initInstruments(sandboxRestClient)
	scope.initChannels()
	scope.initDiskWriters()

	go scope.eventReceiver(streamingClient)

	scope.subscribeOrderbooks(streamingClient)
	scope.subscribeCandles(streamingClient)
	defer scope.unsubscribeOrderbooks(streamingClient)
	defer scope.unsubscribeCandles(streamingClient)

	select {} // sleep(0), epta
}
