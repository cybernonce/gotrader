package okxv5

import (
	"fmt"
	"sync"

	"github.com/cybernonce/gotrader/exchange/base"
	"github.com/cybernonce/gotrader/pkg/ws"
	"github.com/cybernonce/gotrader/trader/constant"
	"github.com/cybernonce/gotrader/trader/types"
)

type OkxV5Exchange struct {
	exchangeType constant.ExchangeType

	restClient   *RestClient
	pubWsClients []*ws.WsClient // 修改为WebSocket客户端数组
	pubWsMutex   sync.RWMutex   // 添加互斥锁以保护并发访问
	pubWsIndex   int            // 当前使用的WebSocket索引
	priWsClient  *ws.WsClient // 私有WebSocket客户端

	// callbacks
	onBooktickerCallback func(*types.BookTicker)
	onOrderCallback      func([]*types.Order)
	onTradeCallback      func([]*types.Trade)
}

// 最大WebSocket连接数
const maxWsConnections = 20

func NewOkxV5Swap(params *types.ExchangeParameters) *OkxV5Exchange {
	apiKey := params.AccessKey
	secretKey := params.SecretKey
	passPhrase := params.Passphrase

	// new client
	client := NewRestClient(apiKey, secretKey, passPhrase, constant.OkxV5Swap)
	exchange := &OkxV5Exchange{
		exchangeType: constant.OkxV5Swap,
		restClient:   client,
		pubWsClients: make([]*ws.WsClient, 0, maxWsConnections), // 初始化WebSocket客户端数组
	}
	
	// 创建第一个公共WebSocket连接
	exchange.ensurePubWsClient()
	
	// priWsClient
	if len(apiKey) > 0 {
		priWsClient := NewOkPriWsClient(apiKey, secretKey, passPhrase, exchange.OnPriWsHandle)
		if err := priWsClient.Dial(ws.Connect); err != nil {
			log.Errorf("priWsClient.Dial err %s", err)
		} else {
			exchange.priWsClient = priWsClient
			log.Infof("priWsClient.Dial success")
		}
	}
	return exchange
}

func NewOkxV5Spot(params *types.ExchangeParameters) *OkxV5Exchange {
	apiKey := params.AccessKey
	secretKey := params.SecretKey
	passPhrase := params.Passphrase

	// new client
	client := NewRestClient(apiKey, secretKey, passPhrase, constant.OkxV5Spot)
	exchange := &OkxV5Exchange{
		exchangeType: constant.OkxV5Spot,
		restClient:   client,
		pubWsClients: make([]*ws.WsClient, 0, maxWsConnections), // 初始化WebSocket客户端数组
	}
	
	// 创建第一个公共WebSocket连接
	exchange.ensurePubWsClient()

	if len(apiKey) > 0 {
		priWsClient := NewOkPriWsClient(apiKey, secretKey, passPhrase, exchange.OnPriWsHandle)
		if err := priWsClient.Dial(ws.Connect); err != nil {
			log.Errorf("priWsClient.Dial err %s", err)

		} else {
			exchange.priWsClient = priWsClient
			log.Infof("priWsClient.Dial success")
		}
	}
	return exchange
}

// 确保至少有一个公共WebSocket连接可用
func (okx *OkxV5Exchange) ensurePubWsClient() *ws.WsClient {
	okx.pubWsMutex.Lock()
	defer okx.pubWsMutex.Unlock()
	
	if len(okx.pubWsClients) == 0 {
		pubWsClient := NewOkPubWsClient(okx.OnPubWsHandle)
		if err := pubWsClient.Dial(ws.Connect); err != nil {
			log.Errorf("pubWsClient.Dial err %s", err)
			return nil
		} else {
			okx.pubWsClients = append(okx.pubWsClients, pubWsClient)
			log.Infof("pubWsClient.Dial success, index: 0")
			return pubWsClient
		}
	}
	
	return okx.pubWsClients[0]
}

// 获取下一个可用的WebSocket客户端，实现负载均衡
func (okx *OkxV5Exchange) getNextPubWsClient() *ws.WsClient {
	okx.pubWsMutex.Lock()
	defer okx.pubWsMutex.Unlock()
	
	// 如果还可以创建更多连接
	if len(okx.pubWsClients) < maxWsConnections {
		pubWsClient := NewOkPubWsClient(okx.OnPubWsHandle)
		if err := pubWsClient.Dial(ws.Connect); err != nil {
			log.Errorf("pubWsClient.Dial err %s", err)
			// 如果创建新连接失败，使用现有连接
			okx.pubWsIndex = (okx.pubWsIndex + 1) % len(okx.pubWsClients)
			return okx.pubWsClients[okx.pubWsIndex]
		} else {
			index := len(okx.pubWsClients)
			okx.pubWsClients = append(okx.pubWsClients, pubWsClient)
			okx.pubWsIndex = index
			log.Infof("pubWsClient.Dial success, index: %d", index)
			return pubWsClient
		}
	}
	
	// 已达到最大连接数，使用轮询方式选择下一个连接
	okx.pubWsIndex = (okx.pubWsIndex + 1) % len(okx.pubWsClients)
	return okx.pubWsClients[okx.pubWsIndex]
}

// 获取当前连接数
func (okx *OkxV5Exchange) GetPubWsClientCount() int {
	okx.pubWsMutex.RLock()
	defer okx.pubWsMutex.RUnlock()
	return len(okx.pubWsClients)
}

func (okx *OkxV5Exchange) GetName() (name string) {
	return okx.exchangeType.Name()

}

func (okx *OkxV5Exchange) GetType() (typ constant.ExchangeType) {
	return okx.exchangeType
}

func (okx *OkxV5Exchange) FetchTickers() ([]*types.Ticker, error) {
	return okx.restClient.FetchTickers()
}

func (okx *OkxV5Exchange) FetchKline(symbol string, interval string, limit int64) ([]types.Kline, error) {
	return okx.restClient.FetchKline(symbol, interval, limit)
}

func (okx *OkxV5Exchange) FetchFundingRate(symbol string) (*types.FundingRate, error) {
	return okx.restClient.FetchFundingRate(symbol)
}

func (okx *OkxV5Exchange) FetchFundingRateHistory(symbol string, limit int64) ([]*types.FundingRate, error) {
	return okx.restClient.FetchFundingRateHistory(symbol, limit)
}

func (okx *OkxV5Exchange) FetchSymbols() ([]*types.SymbolInfo, error) {
	return okx.restClient.FetchSymbols()
}

func (okx *OkxV5Exchange) FetchBalance() (*types.Assets, error) {
	return okx.restClient.FetchBalance()
}

func (okx *OkxV5Exchange) FetchAssetBalance() (*types.Assets, error) {
	return okx.restClient.FetchAssetBalance()
}

func (okx *OkxV5Exchange) FetchPositons() ([]*types.Position, error) {
	return okx.restClient.FetchPositons()
}

func (okx *OkxV5Exchange) CreateBatchOrders(orders []*types.Order) ([]*types.OrderResult, error) {
	return okx.restClient.CreateBatchOrders(orders)
}

func (okx *OkxV5Exchange) CancelBatchOrders(orders []*types.Order) ([]*types.OrderResult, error) {
	return okx.restClient.CancelBatchOrders(orders)
}

func (okx *OkxV5Exchange) PrivateTransfer(transfer base.TransferParam) (string, error) {
	return okx.restClient.PrivateTransfer(transfer)
}

func (okx *OkxV5Exchange) Subscribe(params map[string]interface{}) error {
	// 使用负载均衡获取一个WebSocket客户端
	wsClient := okx.getNextPubWsClient()
	if wsClient == nil {
		return fmt.Errorf("no available pubWsClient")
	}
	
	// 发送订阅请求
	if err := wsClient.Write(params); err != nil {
		return fmt.Errorf("Subscribe err: %s", err)
	}
	return nil
}

func (okx *OkxV5Exchange) SubscribeBookTicker(symbols []string, callback func(*types.BookTicker)) error {
	for _, symbol := range symbols {
		// 使用负载均衡获取一个WebSocket客户端
		wsClient := okx.getNextPubWsClient()
		if wsClient == nil {
			return fmt.Errorf("no available pubWsClient")
		}
		
		// 发送订阅请求
		wsClient.Subscribe(symbol, "bbo-tbt")
	}

	okx.onBooktickerCallback = callback
	return nil
}

func (okx *OkxV5Exchange) SubscribeTrades(symbols []string, callback func([]*types.Trade)) error {
	for _, symbol := range symbols {
		// 使用负载均衡获取一个WebSocket客户端
		wsClient := okx.getNextPubWsClient()
		if wsClient == nil {
			return fmt.Errorf("no available pubWsClient")
		}
		
		// 发送订阅请求
		wsClient.Subscribe(symbol, "trades")
	}
	okx.onTradeCallback = callback
	return nil
}

// SubscribeOrder 订阅订单频道
func (okx *OkxV5Exchange) SubscribeOrders(symbols []string, callback func(orders []*types.Order)) error {
	/***
	{
		"op": "subscribe",
		"args": [{
			"channel": "orders",
			"instType": "FUTURES",
			"instFamily": "BTC-USD"
		}]
	}
	***/

	for _, symbol := range symbols {
		okx.priWsClient.Subscribe(symbol, "orders")
	}

	okx.onOrderCallback = callback
	return nil
}

func (okx *OkxV5Exchange) OnPubWsHandle(data interface{}) {
	switch v := data.(type) {
	case *types.BookTicker:
		// callback
		if okx.onBooktickerCallback != nil {
			okx.onBooktickerCallback(v)
		} else {
			log.Errorf("OnBookTicker Callback not set")
		}
	case *types.OrderBook:
		fmt.Println("OrderBook type", v)
	case []*types.Trade:
		if okx.onTradeCallback != nil {
			okx.onTradeCallback(v)
		} else {
			log.Errorf("onTrade Callback not set")
		}
	default:
		log.Errorf("Unknown type %s", v)
	}
}

func (okx *OkxV5Exchange) OnPriWsHandle(data interface{}) {
	switch v := data.(type) {
	case []*types.Order:
		if okx.onOrderCallback != nil {
			okx.onOrderCallback(v)
		} else {
			log.Errorf("onOrder Callback not set")
		}
	default:
		log.Errorf("Unknown type %s", v)
	}
}

// FetchInterestLimit gets the borrow interest and limit for margin or portfolio margin
func (okx *OkxV5Exchange) FetchInterestLimit(ccy string) (map[string]OkInterestLimit, error) {
	return okx.restClient.FetchInterestLimit(ccy)
}
