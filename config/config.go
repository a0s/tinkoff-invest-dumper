package config

import (
	"flag"
	"fmt"
	"github.com/octago/sflags/gen/gflag"
	"log"
	"os"
	"time"
)

var VersionString = "development"

var Conf Config

type Config struct {
	Token string `flag:"token" desc:"your sandbox's token"`
	Path  string `flag:"path" desc:"path to storage dir"`

	TimeSuffixEnabled   bool      `flag:"time-suffix-enabled" desc:"add the time suffix to every filename on (re)start"`
	TimeSuffixFormat    string    `flag:"time-suffix-format" desc:"go format of the time suffix (see https://golang.org/src/time/format.go)"`
	TimeSuffixStartedAt time.Time `flag:"-"`

	Orderbook      string `flag:"orderbook" desc:"list of tickers to subscribe for orderbooks"`
	OrderbookDepth int    `flag:"orderbook-depth" desc:"depth of orderbook: from 1 to 20"`

	Candle         string `flag:"candle" desc:"list of tickers to subscribe for candles"`
	CandleInterval string `flag:"candle-interval" desc:"interval of candles: 1min,2min,3min,5min,10min,15min,30min,hour,2hour,4hour,day,week,month"`

	Version bool `flag:"version" desc:"show version info"`
}

func NewConfig() *Config {
	config := &Config{
		Path:                ".",
		TimeSuffixEnabled:   false,
		TimeSuffixFormat:    "2006010215",
		TimeSuffixStartedAt: time.Now().UTC(),
		OrderbookDepth:      20,
		CandleInterval:      "1min",
		Version:             false,
	}
	err := gflag.ParseToDef(config)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	flag.Parse()

	return config
}

func init() {
	Conf = *NewConfig()
	if Conf.Version {
		fmt.Printf("%s\n", VersionString)
		os.Exit(0)
	}
}
