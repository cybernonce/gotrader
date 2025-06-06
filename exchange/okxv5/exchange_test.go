package okxv5

import (
	"testing"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/wsg011/gotrader/exchange/base"
	"github.com/wsg011/gotrader/pkg/utils"
	"github.com/wsg011/gotrader/trader/constant"
	"github.com/wsg011/gotrader/trader/types"
)

var (
	apiKey      = "b6585a77-34ea-4244-93dc-7e028195c116"
	secretKey   = "C5CA4D0476EBB1E146F8E5A9C102902C"
	passphrase  = "m2rA3JyH*J"
	symbol      = "APE_USDT_SWAP"
	hedgeSymbol = "APE_USDT"
	askPrice    = 0.0
	bidPrice    = 0.0
)

var exchange *OkxV5Exchange

func TestMain(t *testing.T) {
	params := &types.ExchangeParameters{
		AccessKey:  apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
	}
	exchange = NewOkxV5Swap(params)
	name := exchange.GetName()
	t.Logf("init Exchang name %s", name)
}

func TestFetchTickers(t *testing.T) {
	resp, err := exchange.FetchTickers()
	if err != nil {
		t.Fatalf("HttpRequest failed: %v", err)
	}

	for _, ticker := range resp {
		if ticker.Symbol == symbol {
			bidPrice = ticker.BidPrice
			askPrice = ticker.AskPrice
			t.Logf("%s ticker [%f | %f]", symbol, ticker.AskPrice, ticker.BidPrice)
		}
	}
	// t.Logf("Tickers %v", resp[0])
}

func TestFetchSymbols(t *testing.T) {
	resp, err := exchange.FetchSymbols()
	if err != nil {
		t.Fatalf("FetchSymbols err: %s", err)
	}

	for _, symbolinfo := range resp {
		if symbolinfo.Symbol == symbol {
			t.Logf("symbol info %s PxPrec %v QtyPrec %v FaceVal %v Multi %v",
				symbolinfo.Symbol, symbolinfo.PxPrec, symbolinfo.QtyPrec, symbolinfo.FaceVal, symbolinfo.Multiplier)
		}
		// t.Logf("symbol info %s PxPrec %v FaceVal %v Multiplier %v", symbolinfo.Symbol, symbolinfo.PxPrec, symbolinfo.FaceVal, symbolinfo.Multiplier)
	}
}

func TestFetchFundingRate(t *testing.T) {
	resp, err := exchange.FetchFundingRateHistory(symbol, 10)
	if err != nil {
		t.Fatalf("FetchFundingRateHistory err: %s", err)
	}

	fundingRates := make([]float64, 0)
	for i := 1; i < len(resp); i++ {
		fr := resp[i].FundingRate
		fundingRates = append(fundingRates, fr)
	}
	mean, _ := stats.Mean(fundingRates)
	t.Logf("funding rate history mean %v", mean)
}

func TestFetchAssetBalance(t *testing.T) {
	params := &types.ExchangeParameters{
		AccessKey:  apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
	}
	exchange = NewOkxV5Swap(params)

	assetBalance, err := exchange.FetchAssetBalance()
	if err != nil {
		t.Errorf("FetchAssetBalance error %s", err)
	}
	t.Logf("asset balance %+v", assetBalance)

	bethAsset, exists := assetBalance.Assets["BETH"]
	if exists {
		transfer := base.TransferParam{
			FromType:     "funding",
			ToType:       "swap",
			Assert:       bethAsset.Coin,
			Amount:       bethAsset.Total,
			TransferType: "0",
		}
		transferId, err := exchange.PrivateTransfer(transfer)
		if err != nil {
			t.Errorf("PrivateTransfer err %s", err)
		}
		t.Logf("PrivateTransfer success, trans id %s", transferId)
	}

}

// func TestFetchPosition(t *testing.T) {
// 	resp, err := exchange.FetchPositons()
// 	if err != nil {
// 		t.Fatalf("FetchPositons err: %s", err)
// 	}

// 	for _, position := range resp {
// 		if position.Symbol == symbol {
// 			t.Logf("symbol position %v", position.Position)
// 		}
// 		t.Logf("symbol %s position %v", position.Symbol, position.Position)
// 	}
// }

// func TestCreateBatchOrders(t *testing.T) {
// 	now := time.Now()
// 	orderNum := 3
// 	// 5 合约 order
// 	orders := make([]*types.Order, 0)
// 	for i := 0; i < orderNum; i++ {
// 		price := askPrice + 0.01*float64(i+1)
// 		orderinfo := &types.Order{
// 			Symbol:       symbol,
// 			Side:         constant.OrderSell,
// 			Type:         constant.Limit,
// 			ClientID:     utils.RandomString(32),
// 			Price:        utils.FormatFloat(price, 3),
// 			OrigQty:      "1",
// 			ExchangeType: exchange.exchangeType,
// 			CreateAt:     utils.Millisec(now),
// 			Status:       constant.OrderSubmit,
// 		}
// 		orders = append(orders, orderinfo)
// 		// t.Logf("create order %v", orderinfo)
// 	}
// 	// 5 现货 order
// 	for i := 0; i < orderNum; i++ {
// 		price := bidPrice - 0.01*float64(i+1)
// 		orderinfo := &types.Order{
// 			Symbol:       hedgeSymbol,
// 			Side:         constant.OrderBuy,
// 			Type:         constant.Limit,
// 			ClientID:     utils.RandomString(32),
// 			Price:        utils.FormatFloat(price, 3),
// 			OrigQty:      "1",
// 			ExchangeType: exchange.exchangeType,
// 			CreateAt:     utils.Millisec(now),
// 			Status:       constant.OrderSubmit,
// 		}
// 		orders = append(orders, orderinfo)
// 	}

// 	t.Logf("create batch orders %v", orders)
// 	result, err := exchange.CreateBatchOrders(orders)
// 	if err != nil {
// 		t.Errorf("CreateBatchOrders error: %s", err)
// 	}

// 	for _, orderRes := range result {
// 		t.Logf("orderRes %v", orderRes)
// 	}

// 	time.Sleep(5 * time.Second)
// 	cancelResult, err := exchange.CancelBatchOrders(orders)
// 	if err != nil {
// 		t.Errorf("CancelBatchOrders err %s", err)
// 	}
// 	for _, order := range cancelResult {
// 		t.Logf("cance orderRes %v", order)
// 	}

// }

func TestCreateBatchOrdersWs(t *testing.T) {
	var now time.Time

	// 订阅orders频道
	onOrder := func(orders []*types.Order) {
		costTime := time.Since(now)
		t.Logf("本地下单 To 回报 %v", costTime)
		for _, order := range orders {
			// t.Logf("orderinfo %+v", order)
			createAtTime := time.Unix(0, order.CreateAt*int64(time.Millisecond))
			t.Logf("交易撮合 To 回报: %v\n", time.Since(createAtTime))
		}
	}

	err := exchange.SubscribeOrders([]string{hedgeSymbol, symbol}, onOrder)
	if err != nil {
		t.Fatalf("SubscribeOrders error %s", err)
	}
	time.Sleep(1 * time.Second)

	for i := 0; i < 3; i++ {

		// 下单
		orderNum := 1
		// 合约 order
		orders := make([]*types.Order, 0)
		for i := 0; i < orderNum; i++ {
			price := askPrice + 0.01*float64(i+1)
			orderinfo := &types.Order{
				Symbol:   symbol,
				Side:     constant.OrderSell,
				Type:     constant.Limit,
				ClientID: utils.RandomString(32),
				Price:    utils.FormatFloat(price, 3),
				OrigQty:  "10",
				Exchange: exchange.exchangeType,
				CreateAt: utils.Millisec(now),
				Status:   constant.OrderSubmit,
			}
			orders = append(orders, orderinfo)
			// t.Logf("create order %v", orderinfo)
		}
		// 现货 order
		for i := 0; i < orderNum; i++ {
			price := bidPrice - 0.01*float64(i+1)
			orderinfo := &types.Order{
				Symbol:   hedgeSymbol,
				Side:     constant.OrderBuy,
				Type:     constant.Limit,
				ClientID: utils.RandomString(32),
				Price:    utils.FormatFloat(price, 3),
				OrigQty:  "1",
				Exchange: exchange.exchangeType,
				CreateAt: utils.Millisec(now),
				Status:   constant.OrderSubmit,
			}
			orders = append(orders, orderinfo)
		}

		// t.Logf("create batch ws orders %v", orders)
		// now = time.Now()
		// err = exchange.CreateBatchOrdersWs(orders)
		// if err != nil {
		// 	t.Errorf("CreateBatchOrders error: %s", err)
		// }

		// time.Sleep(5 * time.Second)
		// cancelResult, err := exchange.CancelBatchOrders(orders)
		// if err != nil {
		// 	t.Errorf("CancelBatchOrders err %s", err)
		// }
		// for _, order := range cancelResult {
		// 	t.Logf("cance orderRes %v", order)
		// }
		// time.Sleep(5 * time.Second)
	}

}

// func TestExchange(t *testing.T) {
// 	params := &types.ExchangeParameters{
// 		AccessKey:  "",
// 		SecretKey:  "",
// 		Passphrase: "",
// 	}
// 	exchange := NewOkxV5Swap(params)
// 	name := exchange.GetName()
// 	t.Logf("init exchange %s", name)

// 	param := map[string]interface{}{
// 		"op": "subscribe",
// 		"args": []map[string]string{
// 			{
// 				"channel": "bbo-tbt",
// 				"instId":  "BTC-USDT",
// 			},
// 		},
// 	}

// 	exchange.Subscribe(param)
// }

// func TestSwapWs(t *testing.T) {
// 	testRspHandle := func(data interface{}) {
// 		// t.Logf("testRspHandle data: %s", data)
// 	}
// 	okxWsClient := NewOkPubWsClient(testRspHandle)
// 	if err := okxWsClient.Dial(ws.Connect); err != nil {
// 		t.Fatalf("Dial error: %s", err)
// 	}
// 	t.Log("Dial success")
// 	param := map[string]interface{}{
// 		"op": "subscribe",
// 		"args": []map[string]string{
// 			{
// 				"channel": "bbo-tbt",
// 				"instId":  "BTC-USDT",
// 			},
// 		},
// 	}
// 	okxWsClient.Write(param)

// 	time.Sleep(3 * time.Second)
// 	okxWsClient.Close()
// }

func TestFetchInterestLimit(t *testing.T) {
	params := &types.ExchangeParameters{
		AccessKey:  apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
	}
	exchange = NewOkxV5Swap(params)
	name := exchange.GetName()
	t.Logf("init Exchang name %s", name)

	// Test with specific currency
	result, err := exchange.FetchInterestLimit("USDT")
	if err != nil {
		t.Fatalf("FetchInterestLimit for USDT failed: %v", err)
	}
	t.Logf("Sample interest limit data: %+v", result)
}
