package config

import (
	"flag"
	"github.com/octago/sflags/gen/gflag"
	"path/filepath"
	"strings"
	"time"
	dict "tinkoff-invest-dumper/dictionary"
)

var VersionString = "development"

type Logger interface {
	Fatalln(v ...interface{})
}

type Config struct {
	Token               string    `flag:"token" desc:"your sandbox's token"`
	Path                string    `flag:"path" desc:"path to storage dir"`
	TimeSuffixEnabled   bool      `flag:"time-suffix-enabled" desc:"add the time suffix to every filename on (re)start"`
	TimeSuffixFormat    string    `flag:"time-suffix-format" desc:"go format of the time suffix (see https://golang.org/src/time/format.go)"`
	TimeSuffixStartedAt time.Time `flag:"-"`
	Orderbook           string    `flag:"orderbook" desc:"list of tickers to subscribe for orderbooks"`
	OrderbookDepth      int       `flag:"orderbook-depth" desc:"depth of orderbook: from 1 to 20"`
	Candle              string    `flag:"candle" desc:"list of tickers to subscribe for candles"`
	CandleInterval      string    `flag:"candle-interval" desc:"interval of candles: 1min,2min,3min,5min,10min,15min,30min,hour,2hour,4hour,day,week,month"`
	Version             bool      `flag:"version" desc:"show version info"`

	logger Logger
}

func NewConfig(l Logger) *Config {
	config := &Config{
		Path:                ".",
		TimeSuffixEnabled:   false,
		TimeSuffixFormat:    "2006010215",
		TimeSuffixStartedAt: time.Now().UTC(),
		OrderbookDepth:      20,
		CandleInterval:      "1min",
		Version:             false,

		logger: l,
	}
	err := gflag.ParseToDef(config)
	if err != nil {
		l.Fatalln("new config:", err)
	}
	flag.Parse()

	return config
}

func (c *Config) BuildOrderbookPath(ticker dict.Ticker) string {
	var arr []string

	arr = append(arr, string(ticker))
	if c.TimeSuffixEnabled {
		startedAt := c.TimeSuffixStartedAt.Format(c.TimeSuffixFormat)
		arr = append(arr, startedAt)
	}
	arr = append(arr, "obk")

	path, err := filepath.Abs(filepath.Join(c.Path, strings.Join(arr, "-")))
	if err != nil {
		c.logger.Fatalln(err)
	}
	return path
}

func (c *Config) BuildCandlePath(ticker dict.Ticker) string {
	var arr []string

	arr = append(arr, string(ticker))
	if c.TimeSuffixEnabled {
		startedAt := c.TimeSuffixStartedAt.Format(c.TimeSuffixFormat)
		arr = append(arr, startedAt)
	}
	arr = append(arr, "cdl")

	path, err := filepath.Abs(filepath.Join(c.Path, strings.Join(arr, "-")))
	if err != nil {
		c.logger.Fatalln(err)
	}
	return path
}

func (c *Config) GetOrderbookTickers() []dict.Ticker {
	return stringToTickerList(c.Orderbook)
}

func (c *Config) GetCandleTickers() []dict.Ticker {
	return stringToTickerList(c.Candle)
}

func stringToTickerList(flag string) []dict.Ticker {
	var tickers []dict.Ticker
	flags := strings.Split(flag, ",")
	for _, f := range flags {
		if f != "" {
			tickers = append(tickers, dict.Ticker(f))
		}
	}
	return tickers
}
