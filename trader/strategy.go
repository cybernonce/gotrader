package trader

import "github.com/cybernonce/gotrader/trader/types"

type Strategy interface {
	GetName() string
	GetSymbol() string
	GetHedgeSymbol() string
	OnBookTicker(bookticker *types.BookTicker)
	OnOrderBook(orderbook *types.OrderBook)
	OnTrade(trade *types.Trade)
	OnOrder(order []*types.Order)
	Run()
	Start()
	Close()
}
