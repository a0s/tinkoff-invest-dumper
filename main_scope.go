package main

import (
	"context"
	"encoding/json"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ticker string // eg. MSFT
type figi string   // eg. BBG000BPH459

type wrappedEvent struct {
	time   time.Time
	ticker ticker
	event  interface{}
}
type eventChannel chan *wrappedEvent

type mainScope struct {
	orderbookTickers []ticker
	candleTickers    []ticker

	orderbookFigiChannels map[figi]eventChannel
	candlesFigiChannels   map[figi]eventChannel

	figiInstrument   map[figi]sdk.Instrument
	tickerInstrument map[ticker]sdk.Instrument

	logger *log.Logger
}

func NewMainScope(orderbookTickers []ticker, candleTickers []ticker, logger *log.Logger) *mainScope {
	return &mainScope{
		orderbookTickers: orderbookTickers,
		candleTickers:    candleTickers,

		orderbookFigiChannels: map[figi]eventChannel{},
		candlesFigiChannels:   map[figi]eventChannel{},

		figiInstrument:   map[figi]sdk.Instrument{},
		tickerInstrument: map[ticker]sdk.Instrument{},

		logger: logger,
	}
}

func (s *mainScope) initInstruments(restClient *sdk.SandboxRestClient) {
TICKERS:
	for _, ticker := range s.allTickers() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		foundInstruments, err := restClient.InstrumentByTicker(ctx, string(ticker))
		if err != nil {
			s.logger.Fatalln(err)
		}
		for _, instrument := range foundInstruments {
			if instrument.Ticker == string(ticker) {
				s.tickerInstrument[ticker] = instrument
				s.figiInstrument[figi(instrument.FIGI)] = instrument
				continue TICKERS
			}
		}

		cancel()
	}
}

func (s *mainScope) initChannels() {
	for _, ticker := range s.allTickers() {
		instrument := s.tickerInstrument[ticker]
		figi := figi(instrument.FIGI)

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
		var t ticker
		var f figi

		switch realEvent := event.(type) {
		case sdk.OrderBookEvent:
			f = figi(realEvent.OrderBook.FIGI)
		case sdk.CandleEvent:
			f = figi(realEvent.Candle.FIGI)
		default:
			s.logger.Fatalln("unsupport event type", event)
		}
		t = ticker(s.figiInstrument[f].Ticker)

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

func (s *mainScope) subscribeOrderbooks(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.orderbookTickers {
		instrument := s.tickerInstrument[ticker]
		err := streamingClient.SubscribeOrderbook(instrument.FIGI, *orderbookDepth, requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Subscribed to orderbook", instrument.Ticker, instrument.FIGI)
	}
}

func (s *mainScope) unsubscribeOrderbooks(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.orderbookTickers {
		instrument := s.tickerInstrument[ticker]
		err := streamingClient.UnsubscribeOrderbook(instrument.FIGI, *orderbookDepth, requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Unsubscribed from orderbook", instrument.Ticker, instrument.FIGI)
	}
}

func (s *mainScope) subscribeCandles(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.candleTickers {
		instrument := s.tickerInstrument[ticker]
		err := streamingClient.SubscribeCandle(instrument.FIGI, sdk.CandleInterval(*candleInterval), requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Subscribed to candles", instrument.Ticker, instrument.FIGI)
	}
}

func (s *mainScope) unsubscribeCandles(streamingClient *sdk.StreamingClient) {
	for _, ticker := range s.candleTickers {
		instrument := s.tickerInstrument[ticker]
		err := streamingClient.UnsubscribeCandle(instrument.FIGI, sdk.CandleInterval(*candleInterval), requestID())
		if err != nil {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Unsubscribed from candles", instrument.Ticker, instrument.FIGI)
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
			"ticker":     wrappedEvent.ticker,
			"figi":       figi(event.OrderBook.FIGI),
			"time":       event.Time,
			"local_time": wrappedEvent.time.Format(time.RFC3339Nano),
			"bids":       event.OrderBook.Bids,
			"asks":       event.OrderBook.Asks,
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
			"ticker":     wrappedEvent.ticker,
			"figi":       figi(event.Candle.FIGI),
			"time":       event.Time,
			"local_time": wrappedEvent.time.Format(time.RFC3339Nano),

			"o":     event.Candle.OpenPrice,
			"c":     event.Candle.ClosePrice,
			"h":     event.Candle.HighPrice,
			"l":     event.Candle.LowPrice,
			"v":     event.Candle.Volume,
			"time2": event.Candle.TS,
			"i":     event.Candle.Interval,
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

func (s *mainScope) initDiskWriters() {
	for _, ticker := range s.allTickers() {

		instrument := s.tickerInstrument[ticker]
		figi := figi(instrument.FIGI)

		if _, ok := findTicker(s.orderbookTickers, ticker); ok {
			filePath, err := filepath.Abs(filepath.Join(*path, string(ticker), "_orderbook"))
			if err != nil {
				s.logger.Fatalln(err)
			}
			ch := s.orderbookFigiChannels[figi]
			go s.orderbookWriter(ch, filePath)
		}

		if _, ok := findTicker(s.candleTickers, ticker); ok {
			filePath, err := filepath.Abs(filepath.Join(*path, string(ticker), "_candles"))
			if err != nil {
				s.logger.Fatalln(err)
			}
			ch := s.candlesFigiChannels[figi]
			go s.candleWriter(ch, filePath)
		}
	}
}

func (s *mainScope) allTickers() []ticker {
	tickers := append(s.orderbookTickers, s.candleTickers...)

	table := map[ticker]bool{}
	for _, ticker := range tickers {
		table[ticker] = true
	}

	keys := make([]ticker, len(table))
	i := 0
	for k := range table {
		keys[i] = k
		i++
	}

	return keys
}

func listToTickers(flag string) (tickers []ticker) {
	flagArr := strings.Split(flag, ",")
	for _, str := range flagArr {
		tickers = append(tickers, ticker(str))
	}
	return tickers
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func requestID() string {
	b := make([]rune, 12)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func findTicker(slice []ticker, val ticker) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
