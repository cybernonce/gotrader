package ws

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wsg011/gotrader/trader/constant"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
)

type WsClient struct {
	url      string
	Conn     *websocket.Conn
	wch      chan []byte
	imp      WsImp
	exchange constant.ExchangeType
	subMap   map[string][]string
	priv     bool

	recvPingTime time.Time
	recvPongTime time.Time
	pingInterval time.Duration
	pongTimeout  time.Duration

	mutex  sync.Mutex
	quit   chan struct{}
	closed bool
	epoch  int64
}

// NewWsClient 新建Ws客户端
func NewWsClient(url string, imp WsImp, exchangeType constant.ExchangeType, pingInterval, pongTimeout time.Duration) *WsClient {
	return &WsClient{
		url:          url,
		imp:          imp,
		wch:          make(chan []byte, 1024),
		pingInterval: pingInterval,
		pongTimeout:  pongTimeout,
		exchange:     exchangeType,
	}
}

func (ws *WsClient) SetPingInterval(t time.Duration) {
	ws.pingInterval = t
}

func (ws *WsClient) SetpPongTimeout(t time.Duration) {
	ws.pongTimeout = t
}

func (ws *WsClient) SetRecvPingTime(t time.Time) {
	ws.recvPingTime = t
}

func (ws *WsClient) SetRecvPongTime(t time.Time) {
	ws.recvPongTime = t
}

// 用于监控ws是否断开
func (ws *WsClient) WatchClosed() {
	<-ws.quit
}

func (ws *WsClient) Dial(typ ConnectType) error {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second * 10,
	}
	conn, _, err := dialer.Dial(ws.url, nil)
	if err != nil {
		return fmt.Errorf("ws.Dial:%v", err)
	}

	now := time.Now()
	ws.Conn = conn
	ws.closed = false
	ws.recvPingTime = now
	ws.recvPongTime = now
	ws.quit = make(chan struct{})
	if ws.subMap == nil {
		ws.subMap = make(map[string][]string)
	}

	ws.Conn.SetPingHandler(func(message string) error {
		ws.recvPingTime = time.Now()
		return conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(time.Second*10))
	})

	ws.Conn.SetPongHandler(func(message string) error {
		ws.recvPongTime = time.Now()
		return nil
	})

	atomic.AddInt64(&ws.epoch, 1)
	go ws.readLoop()
	go ws.pingLoop()
	go ws.writeLoop()
	ws.imp.OnConnected(ws, typ)
	return nil
}

func (ws *WsClient) reconnect() {
	for {
		err := ws.Dial(Reconnect)
		if err != nil {
			log.WithError(err).Errorln("Reconnect failed, retrying...")
			time.Sleep(5 * time.Second) // 重连前等待
			continue
		}
		log.Infof("Reconnect success.")

		// 断开重连，订阅
		for symbol, topics := range ws.subMap {
			for _, topic := range topics {
				streams := ws.imp.Subscribe(symbol, topic)
				log.Infof("resubscribe %s", streams)
				ws.Write(streams)
				time.Sleep(100 * time.Millisecond)
			}
		}
		break
	}
}

func (ws *WsClient) Close() {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if ws.closed {
		log.Warn("already closed")
		return
	}
	ws.closed = true
	ws.Conn.Close()
	close(ws.quit)
}

func (ws *WsClient) pingLoop() {
	epoch := ws.epoch
	for !ws.closed || epoch != atomic.LoadInt64(&ws.epoch) {
		if time.Since(ws.recvPongTime) > ws.pongTimeout {
			log.Info("Recv PONG timeout")
			return
		}

		select {
		case <-time.After(ws.pingInterval):
			ws.imp.Ping(ws)
		case <-ws.quit:
			return
		}

		// now := time.Now()
		// bs := []byte(fmt.Sprintf("%d", utils.Millisec(now)))
		// if err := ws.conn.WriteControl(websocket.PingMessage, bs, now.Add(time.Second*10)); err != nil {
		// 	log.WithError(err).Errorln("control ping failed")
		// 	return
		// }
	}
}

func (ws *WsClient) readLoop() {
	conn := ws.Conn
	log.Println("Start WS read loop")
	epoch := ws.epoch
	var needReconnect bool
	defer func() {
		ws.Close()
		if needReconnect {
			ws.reconnect() // 断开时重连
		}
	}()
	for !ws.closed || epoch != atomic.LoadInt64(&ws.epoch) {
		_, body, err := conn.ReadMessage()
		if err != nil {
			log.WithError(err).Errorf("websocket conn read timeout")
			needReconnect = true
			return
		}
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.WithField("err", err).Error("handle panic")
				}
			}()
			ws.imp.Handle(ws, body)
		}()
	}
}

func (ws *WsClient) writeLoop() {
	log.Println("Start WS write loop")
	epoch := ws.epoch

	for !ws.closed || epoch != atomic.LoadInt64(&ws.epoch) {
		select {
		case <-ws.quit:
			return
		case bs := <-ws.wch:
			if err := ws.Conn.WriteMessage(websocket.TextMessage, bs); err != nil {
				log.WithError(err).Errorln("write failed")
				return
			}
		}
	}
}

func (ws *WsClient) Subscribe(symbol string, topic string) {
	ovs, ok := ws.subMap[symbol]
	if !ok {
		ovs = append(ovs, topic)
		ws.subMap[symbol] = ovs
	}
	streams := ws.imp.Subscribe(symbol, topic)
	ws.Write(streams)
}

func (ws *WsClient) Write(req interface{}) error {
	bs, err := sonic.Marshal(req)
	if err != nil {
		return err
	}
	ws.wch <- bs
	return nil
}

func (ws *WsClient) WriteBytes(bs []byte) {
	ws.wch <- bs
}
