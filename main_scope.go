package main

import (
	"context"
	"encoding/json"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ticker string
type figi string

type wrappedEvent struct {
	time   time.Time
	ticker ticker
	event  sdk.OrderBookEvent
}
type eventChannel chan *wrappedEvent

type mainScope struct {
	tickers          []ticker
	figiInstrument   map[figi]sdk.Instrument
	tickerInstrument map[ticker]sdk.Instrument
	channels         map[figi]eventChannel
}

func NewMainScope(tickers []ticker) *mainScope {
	return &mainScope{
		tickers:          tickers,
		figiInstrument:   map[figi]sdk.Instrument{},
		tickerInstrument: map[ticker]sdk.Instrument{},
		channels:         map[figi]eventChannel{},
	}
}

func (s *mainScope) initInstruments(restClient *sdk.SandboxRestClient, logger *log.Logger) {
TICKERS:
	for _, ticker := range s.tickers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Hmm, Possible resource leak, 'defer' is called in a 'for' loop
		defer cancel()

		foundInstruments, err := restClient.InstrumentByTicker(ctx, string(ticker))
		if err != nil {
			logger.Fatalln(err)
		}
		for _, instrument := range foundInstruments {
			if instrument.Ticker == string(ticker) {
				s.tickerInstrument[ticker] = instrument
				s.figiInstrument[figi(instrument.FIGI)] = instrument
				continue TICKERS
			}
		}
	}
}

func (s *mainScope) initChannels() {
	for _, instrument := range s.tickerInstrument {
		s.channels[figi(instrument.FIGI)] = make(eventChannel)
	}
}

func (s *mainScope) eventReceiver(streamingClient *sdk.StreamingClient, logger *log.Logger) {
	err := streamingClient.RunReadLoop(func(event interface{}) error {
		e := event.(sdk.OrderBookEvent)
		figi := figi(e.OrderBook.FIGI)
		ticker := ticker(s.figiInstrument[figi].Ticker)

		ce := wrappedEvent{
			time:   time.Now(),
			ticker: ticker,
			event:  e,
		}

		s.channels[figi] <- &ce
		return nil
	})
	if err != nil {
		logger.Fatalln(err)
	}
}

func (s *mainScope) subscribeOrderbooks(streamingClient *sdk.StreamingClient, logger *log.Logger) {
	for _, instrument := range s.tickerInstrument {
		err := streamingClient.SubscribeOrderbook(instrument.FIGI, *depth, requestID())
		if err != nil {
			logger.Fatalln(err)
		}
		logger.Println("Subscribed", instrument.Ticker, instrument.FIGI)
	}
}

func (s *mainScope) unsubscribeOrderbooks(streamingClient *sdk.StreamingClient, logger *log.Logger) {
	for _, instrument := range s.tickerInstrument {
		err := streamingClient.UnsubscribeOrderbook(instrument.FIGI, *depth, requestID())
		if err != nil {
			logger.Fatalln(err)
		}
		logger.Println("Unsubscribed", instrument.Ticker, instrument.FIGI)
	}
}

func (s *mainScope) initDiskWriters(logger *log.Logger) {
	for _, t := range s.tickers {
		ticker := t
		go func() {
			instrument := s.tickerInstrument[ticker]
			ch := s.channels[figi(instrument.FIGI)]

			filename, err := filepath.Abs(filepath.Join(".", string(ticker)))
			if err != nil {
				logger.Fatalln(err)
			}

			f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatalln(err)
			}

			defer func() {
				if err := f.Close(); err != nil {
					logger.Fatalln(err)
				}
			}()

			for event := range ch {
				row := map[string]interface{}{
					"time":   event.time.Format(time.RFC3339Nano),
					"ticker": event.ticker,
					"event":  event.event,
				}

				jsonBytes, err := json.Marshal(row)
				if err != nil {
					logger.Fatalln(err)
				}

				_, err = f.WriteString(fmt.Sprintf("%v\n", string(jsonBytes)))
				if err != nil {
					logger.Fatalln(err)
				}
			}
		}()
	}
}

func flagTickers(flag string) (tikers []ticker) {
	flagArr := strings.Split(flag, ",")
	for _, str := range flagArr {
		tikers = append(tikers, ticker(str))
	}
	return tikers
}
