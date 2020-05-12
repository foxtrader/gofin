# GoFin

[![Minimum Go version](https://img.shields.io/badge/go-1.14.0+-9cf.svg)](#go-version-requirements)
[![Go Report Card](https://goreportcard.com/badge/github.com/foxtrader/gofin)](https://goreportcard.com/report/github.com/foxtrader/gofin)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/foxtrader/gofin/pulls)
[![LICENSE](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)

Financial package for exchanges, calculation, technical analysis, financial data, common types & functions
金融包, 支持交易所, 盈亏, 技术分析, 数据, 常用类型和函数

|   Module    | Description |
| ----------- | ----------- |
|  `ex`       | Unified API of major global `crypto/stock/forex/futures` exchanges. 全球主流`加密货币/股票/外汇/期货`交易所的统一API |
|  `ta`       | Technical Analysis include 50+ indicators. 包括50+指标的技术分析 |
|  `fincalc`  | Profit and loss calc like sharpe ratio and more. 夏普比率等盈亏计算 |
|  `findata`  | Download data from major financial platforms. 从主流财经平台下载数据 |
|  `fintypes` | Common types & functions. 通用类型和函数 |

[QQ中文社区](https://gitter.im/foxtrader/gofin)

## Usage

```
go get -u github.com/foxtrader/gofin

```

## Features

### Ex

**Ex** is unified api for major global exchanges, including cryptocurrencies, stocks, forex, indices and futures.

**Ex**是全球主流交易所的统一API，包括加密货币、股票、外汇、期货。

| Exchange | Spot | Margin | Futures | Streaming-API |
|----------|------|------|------|------|
| Binance  |  OK  |  OK  |  OK  | TODO |
| Huobi    | TODO | TODO | TODO | TODO |
| Kraken   | TODO | TODO | TODO | TODO |
| Bitstamp | TODO | TODO | TODO | TODO |
| IB(InteractiveBrokers) | TODO | TODO | TODO | TODO |
| CTP      | TODO | TODO | TODO | TODO |

### TA (Technical Analysis) 技术分析

Kline patterns and 50+ indicators

### FinCalc 金融计算

PNL: Sharpe Ratio 夏普比率, Annualized Return 年化回报, Max Draw down 最大回撤

### FinData 金融数据

Support Yahoo Finance, CoinMarketCap, CryptoCompare, ECB

### FinTypes 金融基础类型和定义

Financial industry common types & functions.

### Contributors

### Dependencies

[github.com/markcheno/go-talib](https://github.com/markcheno/go-talib)

