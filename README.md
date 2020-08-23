Orderbooks Dumper
=================

JSON dumper of [Tinkoff Invest OpenAPI](https://github.com/TinkoffCreditSystems/invest-openapi)'s orderbooks

Releases
--------

* binary with static linking for: [linux/amd64](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-amd64.tar.gz), [linux/arm64](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-arm64.tar.gz), [linux/armv6](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-armv6.tar.gz), [linux/armv7](https://github.com/a0s/tinkoff-invest-dumper/releases/latest/download/tinkoff-invest-dumper-armv7.tar.gz)

* docker image with amd64, arm64, armv6 and armv7 manifests: [a00s/tinkoff-invest-dumper](https://hub.docker.com/repository/docker/a00s/tinkoff-invest-dumper)

Run as binary
-------------

```shell script
tinkoff-invest-dumper --help                                                                                                                                             ruby-2.7.1
  -depth int
        depth of orderbook (default 1)
  -tickers string
        list of tickers
  -token string
        your sandbox's token
```

Run as Docker image
-------------------

```shell script
docker run --rm -it a00s/tinkoff-invest-dumper --help
Usage of /app/tinkoff-invest-dumper:
  -depth int
    	depth of orderbook (default 1)
  -tickers string
    	list of tickers
  -token string
    	your sandbox's token
```

Example
-------

`tinkoff-invest-dumper -token "$TINKOFF_SANDBOX" -tickers NVDA,MSFT,TSLA -depth 5`

```
2020/08/22 00:01:55 Subscribed MSFT BBG000BPH459
2020/08/22 00:01:55 Subscribed TSLA BBG000N9MNX3
2020/08/22 00:01:55 Subscribed NVDA BBG000BBJQV0
```

`tail -f NVDA`

```json
{"event":{"event":"orderbook","time":"2020-08-21T21:00:53.397580821Z","payload":{"figi":"BBG000BBJQV0","depth":5,"bids":[[507,6],[506.84,53],[506.8,100],[506.75,75],[506.74,10]],"asks":[[507.19,196],[507.2,52],[507.28,75],[507.29,301],[507.35,10]]}},"ticker":"NVDA","time":"2020-08-22T00:00:53.384211+03:00"}
{"event":{"event":"orderbook","time":"2020-08-21T21:01:55.105348328Z","payload":{"figi":"BBG000BBJQV0","depth":5,"bids":[[507,6],[506.84,53],[506.8,100],[506.75,75],[506.74,10]],"asks":[[507.19,196],[507.2,15],[507.28,75],[507.29,301],[507.35,10]]}},"ticker":"NVDA","time":"2020-08-22T00:01:55.089803+03:00"}
{"event":{"event":"orderbook","time":"2020-08-21T21:02:50.885678541Z","payload":{"figi":"BBG000BBJQV0","depth":5,"bids":[[507,6],[506.84,53],[506.8,100],[506.75,75],[506.74,10]],"asks":[[507.19,196],[507.2,15],[507.28,75],[507.29,301],[507.35,10]]}},"ticker":"NVDA","time":"2020-08-22T00:02:50.869458+03:00"}
```
