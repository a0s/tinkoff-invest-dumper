package main

import (
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"os"
	"tinkoff-invest-dumper/config"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	sandboxRestClient := sdk.NewSandboxRestClient(*config.Token)
	streamingClient, err := sdk.NewStreamingClient(logger, *config.Token)
	if err != nil {
		logger.Fatalln(err)
	}
	defer streamingClient.Close()

	scope, err := NewMainScope(sandboxRestClient, parseTickersList(*config.Orderbook), parseTickersList(*config.Candle), logger)
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
