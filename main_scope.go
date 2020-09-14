package main

import (
	"encoding/json"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tinkoff-invest-dumper/config"
	dict "tinkoff-invest-dumper/dictionary"
)

type wrappedEvent struct {
	time   time.Time
	ticker dict.Ticker
	event  interface{}
}
type eventChannel chan *wrappedEvent

type mainScope struct {
	orderbookTickers []dict.Ticker
	candleTickers    []dict.Ticker

	orderbookFigiChannels map[dict.Figi]eventChannel
	candlesFigiChannels   map[dict.Figi]eventChannel

	dict   *dict.Dictionary
	logger *log.Logger
}

func NewMainScope(restClient *sdk.SandboxRestClient, orderbookTickers []dict.Ticker, candleTickers []dict.Ticker, logger *log.Logger) (*mainScope, error) {
	dictionary, err := dict.NewDictionary(restClient, dict.MergeTickers(orderbookTickers, candleTickers))
	if err != nil {
		return nil, err
	}

	scope := &mainScope{
		orderbookTickers: orderbookTickers,
		candleTickers:    candleTickers,

		orderbookFigiChannels: map[dict.Figi]eventChannel{},
		candlesFigiChannels:   map[dict.Figi]eventChannel{},

		dict:   dictionary,
		logger: logger,
	}

	return scope, nil
}

func (s *mainScope) initChannels() {
	for _, ticker := range dict.MergeTickers(s.orderbookTickers, s.candleTickers) {
		figi := s.dict.GetFIGIByTicker(ticker)

		if _, ok := findTicker(s.orderbookTickers, ticker); ok {
			s.orderbookFigiChannels[figi] = make(eventChannel)
		}

		if _, ok := findTicker(s.candleTickers, ticker); ok {
			s.candlesFigiChannels[figi] = make(eventChannel)
		}
	}
}

func (s *mainScope) eventReceiver(streamingClient *sdk.StreamingClient) {
	err := streamingClient.RunReadLoop(func(event interface{}) error {
		var f dict.Figi

		switch realEvent := event.(type) {
		case sdk.OrderBookEvent:
			f = dict.Figi(realEvent.OrderBook.FIGI)
		case sdk.CandleEvent:
			f = dict.Figi(realEvent.Candle.FIGI)
		default:
			s.logger.Fatalln("unsupported event type", event)
		}

		t := s.dict.GetTickerByFIGI(f)

		ce := wrappedEvent{
			time:   time.Now(),
			ticker: t,
			event:  event,
		}

		switch event.(type) {
		case sdk.OrderBookEvent:
			s.orderbookFigiChannels[f] <- &ce
		case sdk.CandleEvent:
			s.candlesFigiChannels[f] <- &ce
		}

		return nil
	})
	if err != nil {
		s.logger.Fatalln(err)
	}
}

func (s *mainScope) subscribeOrderbook(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.orderbookTickers {
		figi := s.dict.GetFIGIByTicker(ticker)
		err := streamingClient.SubscribeOrderbook(string(figi), *config.OrderbookDepth, requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Subscribed to orderbook", ticker, figi)
	}
}

func (s *mainScope) unsubscribeOrderbook(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.orderbookTickers {
		figi := s.dict.GetFIGIByTicker(ticker)
		err := streamingClient.UnsubscribeOrderbook(string(figi), *config.OrderbookDepth, requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Unsubscribed from orderbook", ticker, figi)
	}
}

func (s *mainScope) subscribeCandles(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.candleTickers {
		figi := s.dict.GetFIGIByTicker(ticker)
		err := streamingClient.SubscribeCandle(string(figi), sdk.CandleInterval(*config.CandleInterval), requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Subscribed to candles", ticker, figi)
	}
}

func (s *mainScope) unsubscribeCandles(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.candleTickers {
		figi := s.dict.GetFIGIByTicker(ticker)
		err := streamingClient.UnsubscribeCandle(string(figi), sdk.CandleInterval(*config.CandleInterval), requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Unsubscribed from candles", ticker, figi)
	}
}

func (s *mainScope) orderbookWriter(ch eventChannel, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Fatalln(err)
		}
	}()

	for wrappedEvent := range ch {
		event := wrappedEvent.event.(sdk.OrderBookEvent)
		row := map[string]interface{}{
			"ticker": wrappedEvent.ticker,
			"figi":   dict.Figi(event.OrderBook.FIGI),

			"t":  event.Time,
			"lt": wrappedEvent.time.Format(time.RFC3339Nano),
			"b":  event.OrderBook.Bids,
			"a":  event.OrderBook.Asks,
		}

		jsonBytes, err := json.Marshal(row)
		if err != nil {
			s.logger.Fatalln(err)
		}

		_, err = file.WriteString(fmt.Sprintf("%v\n", string(jsonBytes)))
		if err != nil {
			s.logger.Fatalln(err)
		}
	}
}

func (s *mainScope) candleWriter(ch eventChannel, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Fatalln(err)
		}
	}()

	for wrappedEvent := range ch {
		event := wrappedEvent.event.(sdk.CandleEvent)
		row := map[string]interface{}{
			"ticker": wrappedEvent.ticker,
			"figi":   dict.Figi(event.Candle.FIGI),

			"t":  event.Time,
			"lt": wrappedEvent.time.Format(time.RFC3339Nano),
			"o":  event.Candle.OpenPrice,
			"c":  event.Candle.ClosePrice,
			"h":  event.Candle.HighPrice,
			"l":  event.Candle.LowPrice,
			"v":  event.Candle.Volume,
			"ts": event.Candle.TS,
			"i":  event.Candle.Interval,
		}

		jsonBytes, err := json.Marshal(row)
		if err != nil {
			s.logger.Fatalln(err)
		}

		_, err = file.WriteString(fmt.Sprintf("%v\n", string(jsonBytes)))
		if err != nil {
			s.logger.Fatalln(err)
		}
	}
}

func (s *mainScope) buildFileName(ticker dict.Ticker) (orderbookName, candleName string) {
	var orderbook []string
	var candle []string

	orderbook = append(orderbook, string(ticker))
	candle = append(candle, string(ticker))

	if *config.TimeSuffixEnabled {
		startedAt := config.TimeSuffixStartedAt.Format(*config.TimeSuffixFormat)
		orderbook = append(orderbook, startedAt)
		candle = append(candle, startedAt)
	}

	orderbook = append(orderbook, "obk")
	candle = append(candle, "cdl")

	var err error
	orderbookName, err = filepath.Abs(filepath.Join(*config.Path, strings.Join(orderbook, "-")))
	if err != nil {
		s.logger.Fatalln(err)
	}
	candleName, err = filepath.Abs(filepath.Join(*config.Path, strings.Join(candle, "-")))
	if err != nil {
		s.logger.Fatalln(err)
	}
	return
}

func (s *mainScope) initDiskWriters() {
	for _, ticker := range dict.MergeTickers(s.orderbookTickers, s.candleTickers) {
		figi := s.dict.GetFIGIByTicker(ticker)

		orderbookFilePath, candleFilePath := s.buildFileName(ticker)

		if _, ok := findTicker(s.orderbookTickers, ticker); ok {
			ch := s.orderbookFigiChannels[figi]
			go s.orderbookWriter(ch, orderbookFilePath)
		}

		if _, ok := findTicker(s.candleTickers, ticker); ok {
			ch := s.candlesFigiChannels[figi]
			go s.candleWriter(ch, candleFilePath)
		}
	}
}

func parseTickersList(flag string) []dict.Ticker {
	var tickers []dict.Ticker
	flags := strings.Split(flag, ",")
	for _, f := range flags {
		if f != "" {
			tickers = append(tickers, dict.Ticker(f))
		}
	}
	return tickers
}

func requestID() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 12)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func findTicker(slice []dict.Ticker, val dict.Ticker) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
