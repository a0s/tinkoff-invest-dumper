package eventer

import (
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"math/rand"
	"reflect"
	"time"
	dict "tinkoff-invest-dumper/dictionary"
)

type Logger interface {
	Fatalln(v ...interface{})
}

type Dictionary interface {
	GetFIGIByTicker(t dict.Ticker) (dict.Figi, error)
	GetTickerByFIGI(figi dict.Figi) (dict.Ticker, error)
}

type OrderbookEvent struct {
	Figi      dict.Figi
	Ticker    dict.Ticker
	LocalTime time.Time
	Event     sdk.OrderBookEvent
}

type CandleEvent struct {
	Figi      dict.Figi
	Ticker    dict.Ticker
	LocalTime time.Time
	Event     sdk.CandleEvent
}

type EventReceiver struct {
	streamingClient *sdk.StreamingClient
	logger          Logger
	dictionary      Dictionary

	orderbooks map[dict.Ticker][]chan OrderbookEvent
	candles    map[dict.Ticker][]chan CandleEvent
}

func NewEventReceiver(lg Logger, sc *sdk.StreamingClient, dc Dictionary, ) *EventReceiver {
	return &EventReceiver{
		streamingClient: sc,
		logger:          lg,
		dictionary:      dc,
		orderbooks:      make(map[dict.Ticker][]chan OrderbookEvent),
		candles:         make(map[dict.Ticker][]chan CandleEvent),
	}
}

func (l *EventReceiver) SubscribeToOrderbook(ticker dict.Ticker, depth int) chan OrderbookEvent {
	_, ok := l.orderbooks[ticker]
	if !ok {
		figi, err := l.dictionary.GetFIGIByTicker(ticker)
		if err != nil {
			l.logger.Fatalln("new subscription to orderbook:", err)
		}

		err = l.streamingClient.SubscribeOrderbook(string(figi), depth, requestID())
		if err != nil {
			l.logger.Fatalln("new subscription to orderbook:", err)
		}
		l.orderbooks[ticker] = []chan OrderbookEvent{}
	}

	ch := make(chan OrderbookEvent)
	l.orderbooks[ticker] = append(l.orderbooks[ticker], ch)
	return ch
}

func (l *EventReceiver) SubscribeToCandle(ticker dict.Ticker, interval string) chan CandleEvent {
	_, ok := l.candles[ticker]
	if !ok {
		figi, err := l.dictionary.GetFIGIByTicker(ticker)
		if err != nil {
			l.logger.Fatalln("new candle subscription:", err)
		}

		err = l.streamingClient.SubscribeCandle(string(figi), sdk.CandleInterval(interval), requestID())
		if err != nil {
			l.logger.Fatalln("new candle subscription:", err)
		}
		l.candles[ticker] = []chan CandleEvent{}
	}

	ch := make(chan CandleEvent)
	l.candles[ticker] = append(l.candles[ticker], ch)
	return ch
}

func (l *EventReceiver) WrapOrderbookEvent(e sdk.OrderBookEvent) *OrderbookEvent {
	figi := dict.Figi(e.OrderBook.FIGI)

	ticker, err := l.dictionary.GetTickerByFIGI(figi)
	if err != nil {
		l.logger.Fatalln("create orderbook event:", err)
	}

	return &OrderbookEvent{
		Figi:      figi,
		Ticker:    ticker,
		LocalTime: time.Now(),
		Event:     e,
	}
}

func (l *EventReceiver) WrapCandleEvent(e sdk.CandleEvent) *CandleEvent {
	figi := dict.Figi(e.Candle.FIGI)

	ticker, err := l.dictionary.GetTickerByFIGI(figi)
	if err != nil {
		l.logger.Fatalln("create candle event:", err)
	}

	return &CandleEvent{
		Figi:      figi,
		Ticker:    ticker,
		LocalTime: time.Now(),
		Event:     e,
	}
}

func (l *EventReceiver) Start() {
	for {
		err := l.streamingClient.RunReadLoop(func(event interface{}) error {
			switch sdkEvent := event.(type) {
			case sdk.OrderBookEvent:
				ob := l.WrapOrderbookEvent(sdkEvent)
				channels, ok := l.orderbooks[ob.Ticker]
				if !ok {
					l.logger.Fatalln("event receiver unknown channel:", ob.Ticker)
				}
				for _, ch := range channels {
					ch <- *ob
				}

			case sdk.CandleEvent:
				cd := l.WrapCandleEvent(sdkEvent)
				channels, ok := l.candles[cd.Ticker]
				if !ok {
					l.logger.Fatalln("event receiver unknown channel:", cd.Ticker)
				}
				for _, ch := range channels {
					ch <- *cd
				}

			default:
				l.logger.Fatalln("event receiver unsupported event type:", reflect.TypeOf(event))
			}

			return nil
		})
		if err != nil {
			l.logger.Fatalln("event lister:", err)
		}
	}
}

func requestID() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 12)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano()) // for requestID
}
//
//func (s *mainScope) unsubscribeOrderbook(streamingClient *sdk.StreamingClient) {
//	for _, ticker := range s.orderbookTickers {
//		figi := s.dict.GetFIGIByTicker(ticker)
//		err := streamingClient.UnsubscribeOrderbook(string(figi), config.Conf.OrderbookDepth, requestID())
//		if err != nil {
//			s.logger.Fatalln(err)
//		}
//		s.logger.Println("Unsubscribed from orderbook", ticker, figi)
//	}
//}
//
//
//func (s *mainScope) unsubscribeCandles(streamingClient *sdk.StreamingClient) {
//	for _, ticker := range s.candleTickers {
//		figi := s.dict.GetFIGIByTicker(ticker)
//		err := streamingClient.UnsubscribeCandle(string(figi), sdk.CandleInterval(config.Conf.CandleInterval), requestID())
//		if err != nil {
//			s.logger.Fatalln(err)
//		}
//		s.logger.Println("Unsubscribed from candles", ticker, figi)
//	}
//}
