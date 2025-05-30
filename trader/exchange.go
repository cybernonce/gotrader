package trader

import (
	"github.com/cybernonce/gotrader/exchange/base"
	"github.com/cybernonce/gotrader/trader/constant"
	"github.com/cybernonce/gotrader/trader/types"
)

type Exchange interface {
	GetName() (name string)
	GetType() (typ constant.ExchangeType)

	// rest Public
	FetchKline(symbol string, interval string, limit int64) ([]types.Kline, error)
	FetchFundingRate(symbol string) (*types.FundingRate, error)
	FetchFundingRateHistory(symbol string, limit int64) ([]*types.FundingRate, error)
	FetchSymbols() ([]*types.SymbolInfo, error)
	FetchTickers() ([]*types.Ticker, error)

	// rest Private
	FetchBalance() (*types.Assets, error)
	FetchAssetBalance() (*types.Assets, error)
	FetchPositons() ([]*types.Position, error)
	CreateBatchOrders([]*types.Order) ([]*types.OrderResult, error)
	CancelBatchOrders(orders []*types.Order) ([]*types.OrderResult, error)
	PrivateTransfer(transfer base.TransferParam) (string, error)

	// ws
	Subscribe(params map[string]interface{}) (err error)
	SubscribeBookTicker(symbols []string, callback func(*types.BookTicker)) (err error)
	SubscribeOrders(symbols []string, callback func(orders []*types.Order)) (err error)
}
