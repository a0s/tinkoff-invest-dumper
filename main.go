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
	token         = flag.String("token", "", "your sandbox's token")
	tickers       = flag.String("tickers", "", "list of tickers")
	depth         = flag.Int("depth", 1, "depth of orderbook")
	version       = flag.Bool("version", false, "show version info")
	VersionString string
)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("tinkoff-invest-dumper %s\n", VersionString)
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

	scope := NewMainScope(flagTickers(*tickers))
	scope.initInstruments(sandboxRestClient, logger)
	scope.initChannels()
	scope.initDiskWriters(logger)

	go scope.eventReceiver(streamingClient, logger)

	scope.subscribeOrderbooks(streamingClient, logger)
	defer scope.unsubscribeOrderbooks(streamingClient, logger)

	select {} // sleep(0), epta
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func requestID() string {
	b := make([]rune, 12)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
