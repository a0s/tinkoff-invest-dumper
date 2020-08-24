Streaming Data Saver
====================
[![GitHub release](https://img.shields.io/github/release/a0s/tinkoff-invest-dumper.svg)](https://github.com/a0s/tinkoff-invest-dumper/releases/latest)
![Binary release](https://github.com/a0s/tinkoff-invest-dumper/workflows/Binary%20release/badge.svg)
![Docker image](https://github.com/a0s/tinkoff-invest-dumper/workflows/Docker%20image/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


JSON dumper of [Tinkoff Invest OpenAPI](https://github.com/TinkoffCreditSystems/invest-openapi)'s orderbooks

Releases
--------

* static linked binary for: [linux/amd64](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-amd64.tar.gz), [linux/arm64](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-arm64.tar.gz), [linux/armv6](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-armv6.tar.gz), [linux/armv7](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-armv7.tar.gz)

* docker image with amd64, arm64, armv6 and armv7 manifests: [a00s/tinkoff-invest-dumper](https://hub.docker.com/repository/docker/a00s/tinkoff-invest-dumper)

Options
-------

```shell script
tinkoff-invest-dumper --help                                                                                                                                             ruby-2.7.1
  -candle string
        list of tickers to subscribe for candles
  -candleInterval string
        interval of candles: 1min,2min,3min,5min,10min,15min,30min,hour,2hour,4hour,day,week,month (default "1min")
  -orderbook string
        list of tickers to subscribe for orderbooks
  -orderbook-depth int
        depth of orderbook: from 1 to 20 (default 20)
  -path string
        path to storage dir (default ".")
  -token string
        your sandbox's token
  -version
        show version info
```

Run as Docker image
-------------------

```shell script
docker run \
  --rm -it \
  --env TOKEN=$TINKOFF_SANDBOX \
  --volume `pwd`/data:/data \
  a00s/tinkoff-invest-dumper \
    --token "$TOKEN" \
    --path /data \
    --candle NVDA,MSFT,TSLA \
    --orderbook NVDA,MSFT,TSLA
```

Example
-------

`tinkoff-invest-dumper --token "$TINKOFF_SANDBOX" --candle NVDA,MSFT,TSLA --orderbook NVDA,MSFT,TSLA --orderbook-depth 2`

```
2020/08/24 12:49:15 Subscribed to orderbook NVDA BBG000BBJQV0
2020/08/24 12:49:15 Subscribed to orderbook MSFT BBG000BPH459
2020/08/24 12:49:15 Subscribed to orderbook TSLA BBG000N9MNX3
2020/08/24 12:49:15 Subscribed to candles NVDA BBG000BBJQV0
2020/08/24 12:49:15 Subscribed to candles MSFT BBG000BPH459
2020/08/24 12:49:15 Subscribed to candles TSLA BBG000N9MNX3
```

`tail -f NVDA_orderbook`

```json
{"a":[[514.31,75],[514.71,6]],"b":[[514.3,6],[514.25,10]],"figi":"BBG000BBJQV0","lt":"2020-08-24T12:49:24.866749+03:00","t":"2020-08-24T09:49:24.850272182Z","ticker":"NVDA"}
{"a":[[514.31,75],[514.71,6]],"b":[[514.3,6],[514.25,10]],"figi":"BBG000BBJQV0","lt":"2020-08-24T12:49:25.225449+03:00","t":"2020-08-24T09:49:25.26326835Z","ticker":"NVDA"}
{"a":[[514.31,75],[514.71,6]],"b":[[514.3,6],[514.25,10]],"figi":"BBG000BBJQV0","lt":"2020-08-24T12:49:25.480208+03:00","t":"2020-08-24T09:49:25.50689026Z","ticker":"NVDA"}
```

`tail -f NVDA_candles`

```json
{"c":514.48,"figi":"BBG000BBJQV0","h":514.71,"i":"1min","l":514.48,"lt":"2020-08-24T12:49:15.203217+03:00","o":514.5,"t":"2020-08-24T09:49:15.241791397Z","ticker":"NVDA","ts":"2020-08-24T09:49:00Z","v":11}
{"c":514.48,"figi":"BBG000BBJQV0","h":514.71,"i":"1min","l":514.48,"lt":"2020-08-24T12:49:19.747036+03:00","o":514.5,"t":"2020-08-24T09:49:19.786563182Z","ticker":"NVDA","ts":"2020-08-24T09:49:00Z","v":13}
```
