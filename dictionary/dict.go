package dictionary

import (
	"context"
	"errors"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"sort"
	"strings"
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

func (d *Dictionary) GetFIGIByTicker(t Ticker) (Figi, error) {
	ins, ok := d.tickerInstrument[t]
	if !ok {
		return "", errors.New(fmt.Sprint("ticker not found:", t))
	}

	return Figi(ins.FIGI), nil
}

func (d *Dictionary) GetTickerByFIGI(f Figi) (Ticker, error) {
	ins, ok := d.figiInstrument[f]
	if !ok {
		return "", errors.New(fmt.Sprint("figi not found:", f))
	}

	return Ticker(ins.Ticker), nil
}

type sortTicker []Ticker

func (a sortTicker) Len() int           { return len(a) }
func (a sortTicker) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortTicker) Less(i, j int) bool { return strings.Compare(string(a[i]), string(a[j])) == -1 }

func MergeTickers(s1, s2 []Ticker) []Ticker {
	tickers := append(s1, s2...)

	table := map[Ticker]bool{}
	for _, ticker := range tickers {
		table[ticker] = true
	}

	var keys []Ticker
	for k, _ := range table {
		keys = append(keys, k)
	}

	sort.Sort(sortTicker(keys))
	return keys
}
