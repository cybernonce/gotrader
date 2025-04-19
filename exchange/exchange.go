package exchange

import (
	"fmt"

	"github.com/cybernonce/gotrader/exchange/binanceportfolio"
	"github.com/cybernonce/gotrader/exchange/binancespot"
	"github.com/cybernonce/gotrader/exchange/binanceufutures"
	"github.com/cybernonce/gotrader/exchange/okxv5"
	"github.com/cybernonce/gotrader/trader"
	"github.com/cybernonce/gotrader/trader/constant"
	"github.com/cybernonce/gotrader/trader/types"
)

func NewExchange(exchangeType constant.ExchangeType, params *types.ExchangeParameters) trader.Exchange {
	switch exchangeType {
	case constant.OkxV5Swap:
		return okxv5.NewOkxV5Swap(params)
	case constant.OkxV5Spot:
		return okxv5.NewOkxV5Spot(params)
	case constant.BinanceSpot:
		return binancespot.NewBinanceSpot(params)
	case constant.BinanceUFutures:
		return binanceufutures.NewBinanceUFutures(params)
	case constant.BinancePortfolio:
		return binanceportfolio.NewBinancePortfoli(params)
	default:
		panic(fmt.Sprintf("new exchange error [%v]", exchangeType))
	}
}
