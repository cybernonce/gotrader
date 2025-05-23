package types

import "github.com/cybernonce/gotrader/trader/constant"

type Trade struct {
	Symbol     string
	MarketType constant.ExchangeType
	TradeID    string
	Side       constant.OrderSide
	Price      float64
	Size       float64
	Count      int64
	ExchangeTs int64
	LocalTs    int64
	EventTs    int64
}