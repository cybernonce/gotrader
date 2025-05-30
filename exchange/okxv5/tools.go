package okxv5

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cybernonce/gotrader/trader/constant"
)

var (
	RestUrl  = "https://www.okx.com"
	PubWsUrl = "wss://ws.okx.com:8443/ws/v5/public"
	PriWsUrl = "wss://ws.okx.com:8443/ws/v5/private"
	// RestUrl  = "https://coloapi3.okx.com"
	// PubWsUrl = "wss://colows-d.okx.com/ws/v5/public"
	// PriWsUrl = "wss://colows-d.okx.com/ws/v5/private"

	OkxOrderSide = map[string]string{
		constant.OrderBuy.Name():   "BUY",
		constant.OrderSell.Name():  "SELL",
		constant.Long.Name():       "LONG",
		constant.Short.Name():      "SHORT",
		constant.CloseLong.Name():  "CLOSE_LONG",
		constant.CloseShort.Name(): "CLOSE_SHORT",
		constant.All.Name():        "ALL",
	}

	Side2Okx = map[string]string{
		constant.OrderBuy.Name():  "buy",
		constant.OrderSell.Name(): "sell",
	}

	Okx2Side = map[string]string{
		"buy":  constant.OrderBuy.Name(),
		"sell": constant.OrderSell.Name(),
	}

	OkxOrderType = map[string]string{
		constant.Limit.Name():    "LIMIT",
		constant.Market.Name():   "MARKET",
		constant.GTC.Name():      "GTC",
		constant.IOC.Name():      "IOC",
		constant.FOK.Name():      "FOK",
		constant.PostOnly.Name(): "POST_ONLY",
	}

	Type2Okx = map[string]string{
		constant.Limit.Name():    "limit",
		constant.Market.Name():   "market",
		constant.GTC.Name():      "gtc",
		constant.FOK.Name():      "fok",
		constant.IOC.Name():      "ioc",
		constant.PostOnly.Name(): "post_only",
	}

	Okx2Type = map[string]string{
		"limit":     constant.Limit.Name(),
		"market":    constant.Market.Name(),
		"gtc":       constant.GTC.Name(),
		"fok":       constant.FOK.Name(),
		"ioc":       constant.IOC.Name(),
		"post_only": constant.PostOnly.Name(),
	}

	Okex2Status = map[string]constant.OrderStatus{
		"live":             constant.OrderOpen,
		"partially_filled": constant.OrderPartialFilled,
		"canceled":         constant.OrderCanceled,
		"filled":           constant.OrderFilled,
	}

	Okex2MarginMode = map[string]string{
		"isolated": "FIXED",
		"cross":    "CROSSED",
	}

	TypeMap = map[string]string{
		"spot":    "18",
		"funding": "6",
		"swap":    "18",
		"cswap":   "18",
	}

	TransferMap = map[string]string{
		"1": "1",
		"2": "2",
	}
)

const (
	FetchKlineUri              = "/api/v5/market/candles?%s"
	OrderBookRest              = "/api/v5/market/books?%s"
	SymbolsRest                = "/api/v5/public/instruments"
	FetchFundingRateUri        = "/api/v5/public/funding-rate"
	FetchFundingRateHistoryUri = "/api/v5/public/funding-rate-history"
	TickerRest                 = "/api/v5/market/ticker?%s"
	TickersRest                = "/api/v5/market/tickers?%s"
	TradeRest                  = "/api/v5/market/trades?%s"
	FetchBalanceUri            = "/api/v5/account/balance"
	FetchPositionsUri          = "/api/v5/account/positions"
	FetchAssetBalanceUri       = "/api/v5/asset/balances"
	CreateSingleOrderUri       = "/api/v5/trade/order"
	CreateBatchOrderUri        = "/api/v5/trade/batch-orders"
	CancelSingleOrderUri       = "/api/v5/trade/cancel-order"
	CancelBatchOrderUri        = "/api/v5/trade/cancel-batch-orders"
	FetchOpenOrderUri          = "/api/v5/trade/orders-pending"
	FetchOrderWithIdUri        = "/api/v5/trade/order"
	FetchOrderDefault          = "/api/v5/trade/orders-history-archive"
	FetchUserTradesUri         = "/api/v5/trade/fills-history"
	PrivateTransferUri         = "/api/v5/asset/transfer"
	PrivateCurrenciesUri       = "/api/v5/asset/currencies"
	PrivateWithDrawUri         = "/api/v5/asset/withdrawal"
	FetchDepositHistoryUri     = "/api/v5/asset/deposit-history"
	FetchWithDrawHistoryUri    = "/api/v5/asset/withdrawal-history"
	PrivateDepositAddrUri      = "/api/v5/asset/deposit-address"
	FetchTransferStateUri      = "/api/v5/asset/transfer-state"
	TransferProcessing         = 58124 //提币处理中返回此code
)

// IsoTime eg: 2018-03-16T18:02:48.284Z
func IsoTime() string {
	utcTime := time.Now().UTC()
	iso := utcTime.String()
	isoBytes := []byte(iso)
	iso = string(isoBytes[:10]) + "T" + string(isoBytes[11:23]) + "Z"
	return iso
}

func OkInstId2Symbol(instId string) string {
	tmp := strings.Split(instId, "-")
	if len(tmp) == 2 {
		return fmt.Sprintf("%s_%s", tmp[0], tmp[1])
	} else if len(tmp) == 3 {
		return fmt.Sprintf("%s_%s_SWAP", tmp[0], tmp[1])
	}
	panic("bad instId:" + instId)
}

func Symbol2OkInstId(symbol string) string {
	tmp := strings.Split(symbol, "_")
	if len(tmp) == 2 {
		return fmt.Sprintf("%s-%s", tmp[0], tmp[1])
	} else if len(tmp) == 3 {
		return fmt.Sprintf("%s-%s-SWAP", tmp[0], tmp[1])
	}
	panic("bad symbol:" + symbol)
}

func BaseQuote(symbol string) (string, string) {
	tmp := strings.Split(symbol, "_")
	return tmp[0], tmp[1]
}

func IsPerpSymbol(symbol string) bool {
	return strings.Contains(symbol, "_PERP")
}

// generateSignature 生成签名
func generateOkxSignature(timestamp string, secretKey string) string {
	method := "GET"
	requestPath := "/users/self/verify"
	data := timestamp + method + requestPath

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func parseStringToFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseStringToInt(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
