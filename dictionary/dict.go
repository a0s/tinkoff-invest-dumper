package dictionary

import (
	"context"
	"errors"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"time"
)

type Ticker string // eg. MSFT
type Figi string   // eg. BBG000BPH459

type Dictionary struct {
	figiInstrument   map[Figi]sdk.Instrument
	tickerInstrument map[Ticker]sdk.Instrument
}

func ErrorTickerNotFound(t Ticker) error {
	return errors.New(fmt.Sprintf("Ticker not found: %v", t))
}

func NewDictionary(client *sdk.SandboxRestClient, tickers []Ticker) (*Dictionary, error) {
	fs := map[Figi]sdk.Instrument{}
	ts := map[Ticker]sdk.Instrument{}

TICKERS:
	for _, ticker := range tickers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		foundInstruments, err := client.InstrumentByTicker(ctx, string(ticker))
		if err != nil {
			return nil, err
		}
		if len(foundInstruments) == 0 {
			return nil, ErrorTickerNotFound(ticker)
		}
		for _, instrument := range foundInstruments {
			if instrument.Ticker == string(ticker) {
				ts[ticker] = instrument
				fs[Figi(instrument.FIGI)] = instrument
				continue TICKERS
			}
		}

		cancel()
	}

	return &Dictionary{figiInstrument: fs, tickerInstrument: ts}, nil
}

func (d *Dictionary) GetFIGIByTicker(t Ticker) Figi {
	return Figi(d.tickerInstrument[t].FIGI)
}

func (d *Dictionary) GetTickerByFIGI(f Figi) Ticker {
	return Ticker(d.figiInstrument[f].Ticker)
}

func MergeTickers(s1, s2 []Ticker) []Ticker {
	tickers := append(s1, s2...)

	table := map[Ticker]bool{}
	for _, ticker := range tickers {
		table[ticker] = true
	}

	keys := make([]Ticker, len(table))
	i := 0
	for k := range table {
		keys[i] = k
		i++
	}

	return keys
}
