package main

import (
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"os"
	"tinkoff-invest-dumper/config"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	sandboxRestClient := sdk.NewSandboxRestClient(config.Conf.Token)
	streamingClient, err := sdk.NewStreamingClient(logger, config.Conf.Token)
	if err != nil {
		logger.Fatalln("create streaming client:", err)
	}
	defer func() {
		err := streamingClient.Close()
		if err != nil {
			logger.Fatalln("close streaming client:", err)
		}
	}()

	scope, err := NewMainScope(sandboxRestClient, parseTickersList(config.Conf.Orderbook), parseTickersList(config.Conf.Candle), logger)
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
