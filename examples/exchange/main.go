package main

import (
	"sort"
	"time"

	"github.com/cybernonce/gotrader/exchange/okxv5"
	"github.com/cybernonce/gotrader/pkg/utils"

	"github.com/cybernonce/gotrader/trader/types"

	"github.com/sirupsen/logrus"
)

var (
	timeFormat = "20060102 15:04:05.999"
	log        = logrus.WithField("package", "exchange")
)

// 延迟统计器
type LatencyStats struct {
	// 保存延迟数据的切片
	processingDelays []int64
	feedDelays       []int64
	jsonDelays       []int64 // JSON解析延迟
	// 上次输出统计信息的时间
	lastOutputTime time.Time
	// 输出间隔
	outputInterval time.Duration
	// 数据来源标识
	source string
}

func NewLatencyStats(source string) *LatencyStats {
	return &LatencyStats{
		processingDelays: make([]int64, 0, 10000),
		feedDelays:       make([]int64, 0, 10000),
		jsonDelays:       make([]int64, 0, 10000), // JSON解析延迟
		lastOutputTime:   time.Now(),
		outputInterval:   time.Minute, // 每分钟输出一次统计信息
		source:           source,
	}
}

// 添加延迟数据
func (ls *LatencyStats) AddLatency(jsonDelay, processDelay, feedDelay int64) {
	ls.jsonDelays = append(ls.jsonDelays, jsonDelay)
	ls.processingDelays = append(ls.processingDelays, processDelay)
	ls.feedDelays = append(ls.feedDelays, feedDelay)
}

// 计算百分位数
func percentile(data []int64, p float64) int64 {
	if len(data) == 0 {
		return 0
	}
	
	// 创建数据副本并排序
	dataCopy := make([]int64, len(data))
	copy(dataCopy, data)
	sort.Slice(dataCopy, func(i, j int) bool {
		return dataCopy[i] < dataCopy[j]
	})
	
	// 计算百分位数对应的索引
	idx := int(float64(len(dataCopy)-1) * p)
	return dataCopy[idx]
}

// 输出统计信息
func (ls *LatencyStats) OutputStats() {
	now := time.Now()
	if now.Sub(ls.lastOutputTime) < ls.outputInterval {
		return
	}
	
	if len(ls.processingDelays) == 0 || len(ls.feedDelays) == 0 {
		return
	}
	
	// 计算处理延迟的P50、P90、P99、P999
	processP50 := percentile(ls.processingDelays, 0.5)
	processP90 := percentile(ls.processingDelays, 0.9)
	processP99 := percentile(ls.processingDelays, 0.99)
	processP999 := percentile(ls.processingDelays, 0.999)
	
	// 计算行情延迟的P50、P90、P99、P999
	feedP50 := percentile(ls.feedDelays, 0.5)
	feedP90 := percentile(ls.feedDelays, 0.9)
	feedP99 := percentile(ls.feedDelays, 0.99)
	feedP999 := percentile(ls.feedDelays, 0.999)
	
	// 计算JSON解析延迟的P50、P90、P99、P999
	jsonP50 := percentile(ls.jsonDelays, 0.5)
	jsonP90 := percentile(ls.jsonDelays, 0.9)
	jsonP99 := percentile(ls.jsonDelays, 0.99)
	jsonP999 := percentile(ls.jsonDelays, 0.999)
	
	// 输出统计信息，标识数据来源
	log.Infof("===== %s延迟统计 (样本数: %d) =====", ls.source, len(ls.processingDelays))
	log.Infof("%sJSON解析延迟: P50=%d微秒, P90=%d微秒, P99=%d微秒, P999=%d微秒", ls.source, jsonP50, jsonP90, jsonP99, jsonP999)
	log.Infof("%s处理延迟: P50=%d微秒, P90=%d微秒, P99=%d微秒, P999=%d微秒", ls.source, processP50, processP90, processP99, processP999)
	log.Infof("%s行情延迟: P50=%d微秒, P90=%d微秒, P99=%d微秒, P999=%d微秒", ls.source, feedP50, feedP90, feedP99, feedP999)
	
	// 重置统计数据
	ls.processingDelays = make([]int64, 0, 10000)
	ls.feedDelays = make([]int64, 0, 10000)
	ls.jsonDelays = make([]int64, 0, 10000)
	ls.lastOutputTime = now
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: timeFormat})

	// 创建BookTicker延迟统计器
	bookTickerStats := NewLatencyStats("BookTicker ")
	
	// 创建Trade延迟统计器
	tradeStats := NewLatencyStats("Trade ")

	epoch := 0
	onBookTickerHandle := func(bookticker *types.BookTicker) {
		epoch += 1
		// log.Infof("onBookTickerHandle %v", bookticker)
		
		jsonDelay := bookticker.EventTs - bookticker.LocalTs
		processDelay := utils.Microsec(time.Now()) - bookticker.EventTs
		feedDelay := bookticker.LocalTs - bookticker.ExchangeTs
		
		// 添加延迟数据到统计器
		bookTickerStats.AddLatency(jsonDelay, processDelay, feedDelay)
		
		if epoch%100 == 0 {
			amount := bookticker.AskPrice * bookticker.AskQty
			amount += 1

			log.Infof("%-16s jsonDelay %v processDelay %v feedDelay %v us onBookTickerHandle", 
				bookticker.Symbol, jsonDelay, processDelay, feedDelay)
			bookTickerStats.OutputStats()
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
	
	tradeEpoch := 0
	onTradeHandle := func(trades []*types.Trade) {
		for _, trade := range trades {
			tradeEpoch++
			
			jsonDelay := trade.EventTs - trade.LocalTs
			processDelay := utils.Microsec(time.Now()) - trade.EventTs
			feedDelay := trade.LocalTs - trade.ExchangeTs
			
			// 添加延迟数据到统计器
			tradeStats.AddLatency(jsonDelay, processDelay, feedDelay)
				
			if tradeEpoch%100 == 0 {
				log.Infof("%-16s jsonDelay %v processDelay %v feedDelay %v us onTradeHandle", 
					trade.Symbol, jsonDelay, processDelay, feedDelay)

				tradeStats.OutputStats()		
			}
		}
	}
	err = exchange.SubscribeTrades(symbols, onTradeHandle)
	if err != nil {
		log.Errorf("SubscribeTrades err %s", err)
		return
	}

	select {}
}
