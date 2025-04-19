package main

import (
	"time"

	"github.com/cybernonce/gotrader/exchange/okxv5"
	"github.com/cybernonce/gotrader/pkg/utils"

	"github.com/cybernonce/gotrader/trader/types"

	"github.com/sirupsen/logrus"
)

func main() {
	timeFormat := "20060102 15:04:05.999"
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: timeFormat})
	var log = logrus.WithField("package", "exchange")

	// rspHandle := func(data interface{}) {
	// 	// t.Logf("testRspHandle data: %s", data)
	// }
	// wsClient := okxv5swap.NewOkPubWsClient(rspHandle)
	// if err := wsClient.Dial(1); err != nil {
	// 	fmt.Printf("Dial err %s", err)
	// 	return
	// }

	// param := map[string]interface{}{
	// 	"op": "subscribe",
	// 	"args": []map[string]string{
	// 		{
	// 			"channel": "bbo-tbt",
	// 			"instId":  "BTC-USDT",
	// 		},
	// 	},
	// }
	// wsClient.Subscribe(param)

	epoch := 0
	onBookTickerHandle := func(bookticker *types.BookTicker) {
		epoch += 1
		// log.Infof("onBookTickerHandle %v", bookticker)
		if epoch%100 == 0 {
			amount := bookticker.AskPrice * bookticker.AskQty
			amount += 1

			processDelay := utils.Microsec(time.Now()) - bookticker.Ts
			feedDelay := bookticker.Ts - bookticker.ExchangeTs
			log.Infof("%-16s processDelay %v feedDelay %v us", bookticker.Symbol, processDelay, feedDelay)
		}

	}
	params := &types.ExchangeParameters{
		AccessKey:  "5cf85d68-213c-4d42-8265-7ace3cf55694",
		SecretKey:  "F05B2DDF1F299C8060C810C0EB1DBC30",
		Passphrase: "I/6Ad2qolM05Lh",
	}
	exchange := okxv5.NewOkxV5Swap(params)
	// symbols := []string{"ETH_USDT", "ETH_USDT_SWAP"}
	symbols := []string{"ETH_USDT", "ETH_USDT_SWAP", "BTC_USDT", "BTC_USDT_SWAP", "SOL_USDT", "SOL_USDT_SWAP",
				"DOGE_USDT", "DOGE_USDT_SWAP", "EOS_USDT", "EOS_USDT_SWAP", "ETC_USDT", "ETC_USDT_SWAP", "PEPE_USDT", "PEPE_USDT_PERP"}
	err := exchange.SubscribeBookTicker(symbols, onBookTickerHandle)
	if err != nil {
		log.Errorf("SubscribeBookticker err %s", err)
		return
	}

	// onOrdersHandle := func(orders []*types.Order) {
	// 	for _, order := range orders {
	// 		log.Infof("order %+v", order)
	// 	}
	// }
	// time.Sleep(time.Second)
	// err = exchange.SubscribeOrders([]string{"ETH_USDT_SWAP"}, onOrdersHandle)
	// if err != nil {
	// 	log.Errorf("SubscribeOrders err %s", err)
	// 	return
	// }

	select {}
}
