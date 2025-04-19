package binanceportfolio

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cybernonce/gotrader/pkg/httpx"
	"github.com/cybernonce/gotrader/pkg/utils"
	"github.com/cybernonce/gotrader/trader/constant"
)

var httpClient = httpx.NewClient()

type RestClient struct {
	apiKey       string
	secretKey    string
	passPhrase   string
	exchangeType constant.ExchangeType
	stopChan     chan struct{}
}

type BaseOkRsp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func NewRestClient(apiKey, secretKey, passPhrase string, exchangeType constant.ExchangeType) *RestClient {
	client := &RestClient{
		apiKey:       apiKey,
		secretKey:    secretKey,
		passPhrase:   passPhrase,
		exchangeType: exchangeType,
		stopChan:     make(chan struct{}),
	}
	return client
}
func (client *RestClient) HttpRequest(method string, uri string, param map[string]interface{}) ([]byte, *http.Response, error) {
	if param == nil {
		param = make(map[string]interface{}, 1)
	}
	header := map[string]string{
		"X-MBX-APIKEY": client.apiKey,
	}
	param["timestamp"] = time.Now().UnixMilli() - 1000
	toSignStr := utils.UrlEncodeParams(param)
	signature := utils.GenHexDigest(utils.HmacSha256(toSignStr, client.secretKey))
	url := fmt.Sprintf("%s%s?%s&signature=%s", RestUrl, uri, toSignStr, signature)
	args := &httpx.Request{
		Url:    url,
		Head:   header,
		Method: method,
	}
	body, res, err := httpClient.Request(args)
	if err != nil {
		return nil, res, err
	}
	return *body, res, err
}

func (client *RestClient) HttpGet(url string) ([]byte, *http.Response, error) {
	body, res, err := httpClient.Get(url)
	if err != nil {
		return nil, res, err
	}
	return *body, res, nil
}
