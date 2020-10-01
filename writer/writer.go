package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	dict "tinkoff-invest-dumper/dictionary"
	"tinkoff-invest-dumper/eventer"
)

type Logger interface {
	Fatalln(v ...interface{})
}

type Dictionary interface {
	GetFIGIByTicker(t dict.Ticker) (dict.Figi, error)
	GetTickerByFIGI(figi dict.Figi) (dict.Ticker, error)
}

type Writer struct {
	dictionary Dictionary
	logger     Logger
}

func NewWriter(lg Logger, dc Dictionary) *Writer {
	return &Writer{
		dictionary: dc,
		logger:     lg,
	}
}

func (s *Writer) OrderbookWriter(ch chan eventer.OrderbookEvent, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Fatalln(err)
		}
	}()

	for event := range ch {
		row := map[string]interface{}{
			"ticker": event.Ticker,
			"figi":   event.Figi,

			"t":  event.Event.Time,
			"lt": event.LocalTime.Format(time.RFC3339Nano),
			"b":  event.Event.OrderBook.Bids,
			"a":  event.Event.OrderBook.Asks,
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

func (s *Writer) CandleWriter(ch chan eventer.CandleEvent, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Fatalln(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Fatalln(err)
		}
	}()

	for event := range ch {
		row := map[string]interface{}{
			"ticker": event.Ticker,
			"figi":   event.Figi,

			"t":  event.Event.Time,
			"lt": event.LocalTime.Format(time.RFC3339Nano),
			"o":  event.Event.Candle.OpenPrice,
			"c":  event.Event.Candle.ClosePrice,
			"h":  event.Event.Candle.HighPrice,
			"l":  event.Event.Candle.LowPrice,
			"v":  event.Event.Candle.Volume,
			"ts": event.Event.Candle.TS,
			"i":  event.Event.Candle.Interval,
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
