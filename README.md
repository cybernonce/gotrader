## gotrader

golang Trader Bot


### MarketData service

INFO[0060] ===== BookTicker 延迟统计 (样本数: 17400) =====      package=exchange
INFO[0060] BookTicker JSON解析延迟: P50=6微秒, P90=12微秒, P99=20微秒, P999=79微秒  package=exchange
INFO[0060] BookTicker 处理延迟: P50=0微秒, P90=1微秒, P99=1微秒, P999=3微秒  package=exchange

INFO[0061] ===== Trade 延迟统计 (样本数: 3700) =====            package=exchange
INFO[0061] Trade JSON解析延迟: P50=4微秒, P90=8微秒, P99=14微秒, P999=28微秒  package=exchange
INFO[0061] Trade 处理延迟: P50=0微秒, P90=1微秒, P99=2微秒, P999=8微秒  package=exchange

INFO[0060] BookTicker 行情延迟: P50=-1325微秒, P90=-1003微秒, P99=-220微秒, P999=4837微秒  package=exchange
INFO[0061] Trade 行情延迟: P50=-836微秒, P90=3027微秒, P99=7611微秒, P999=10992微秒  package=exchange

优化后，多ws client

INFO[0060] ===== BookTicker 延迟统计 (样本数: 28203) =====      package=exchange
INFO[0060] BookTicker JSON解析延迟: P50=6微秒, P90=12微秒, P99=20微秒, P999=93微秒  package=exchange
INFO[0060] BookTicker 处理延迟: P50=0微秒, P90=1微秒, P99=1微秒, P999=5微秒  package=exchange
INFO[0060] BTC_USDT_SWAP    jsonDelay 4 processDelay 0 feedDelay -830 us onTradeHandle  package=exchange

INFO[0060] ===== Trade 延迟统计 (样本数: 12400) =====           package=exchange
INFO[0060] Trade JSON解析延迟: P50=4微秒, P90=7微秒, P99=13微秒, P999=69微秒  package=exchange
INFO[0060] Trade 处理延迟: P50=0微秒, P90=1微秒, P99=2微秒, P999=6微秒  package=exchange

INFO[0060] BookTicker 行情延迟: P50=-1260微秒, P90=-828微秒, P99=-209微秒, P999=3587微秒  package=exchange
INFO[0060] Trade 行情延迟: P50=-805微秒, P90=593微秒, P99=5425微秒, P999=9931微秒  package=exchange