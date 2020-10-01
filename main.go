package main

import (
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"os"
	conf "tinkoff-invest-dumper/config"
	dict "tinkoff-invest-dumper/dictionary"
	"tinkoff-invest-dumper/eventer"
	"tinkoff-invest-dumper/writer"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	config := conf.NewConfig(logger)
	if config.Version {
		fmt.Printf("%s\n", conf.VersionString)
		os.Exit(0)
	}

	sandboxRestClient := sdk.NewSandboxRestClient(config.Token)

	streamingClient, err := sdk.NewStreamingClient(logger, config.Token)
	if err != nil {
		logger.Fatalln("streaming client:", err)
	}
	defer func() {
		err := streamingClient.Close()
		if err != nil {
			logger.Fatalln("streaming client:", err)
		}
	}()

	dictionary, err := dict.NewDictionary(sandboxRestClient, dict.MergeTickers(config.GetOrderbookTickers(), config.GetCandleTickers()))
	if err != nil {
		logger.Fatalln("dictionary:", err)
	}

	receiver := eventer.NewEventReceiver(logger, streamingClient, dictionary)
	writer := writer.NewWriter(logger, dictionary)

	for _, ticker := range config.GetOrderbookTickers() {
		channel := receiver.SubscribeToOrderbook(ticker, config.OrderbookDepth)
		figi, err := dictionary.GetFIGIByTicker(ticker)
		if err != nil {
			logger.Fatalln("subscribe ticker:", err)
		}
		logger.Println("Subscribed to orderbook", ticker, figi)
		path := config.BuildOrderbookPath(ticker)
		go writer.OrderbookWriter(channel, path)
	}

	for _, ticker := range config.GetCandleTickers() {
		channel := receiver.SubscribeToCandle(ticker, config.CandleInterval)
		figi, err := dictionary.GetFIGIByTicker(ticker)
		if err != nil {
			logger.Fatalln("subscribe candle:", err)
		}
		logger.Println("Subscribed to candles", ticker, figi)
		path := config.BuildCandlePath(ticker)
		go writer.CandleWriter(channel, path)
	}

	go receiver.Start()

	select {} // sleep(0), epta
}
