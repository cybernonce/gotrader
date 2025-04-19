package okxv5

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cybernonce/gotrader/pkg/utils"
)

type PositionTier struct {
	Tier           string `json:"tier"`  
	MaxSize        string `json:"maxSz"` 
	MaxLever        string `json:"maxLever"` 
}

type PositionTierResponse struct {
	Code string         `json:"code"`
	Data []PositionTier `json:"data"`
}

type OkInterestLimit struct {
	Currency      string `json:"ccy"`  
	// 母账户维度借币限额, 如果已配置可用额度，该字段代表当前交易账户的借币限额       
	Quota       string `json:"loanQuota"`   
	// 母子账户已借额度, 如果已配置可用额度，该字段代表当前交易账户的已借额度
	Used      string `json:"usedLmt"`
	// 不同杠杆下的币对限额
	LeverageQuota      []PositionTier
}

type Debt struct {
	Records []OkInterestLimit `json:"records"`
}

type OkInterestLimitResponse struct {
	Code string            `json:"code"`
	Data []Debt            `json:"data"`
}

func (client *RestClient) FetchInterestLimit(ccy string) (map[string]OkInterestLimit, error) {
	interest0, err := client.fetchInterestLimit(ccy)
	if err != nil{
		return nil, err
	}

	interest1, err := client.fetchPositionTier(ccy)
	if err != nil{
		return nil, err
	}

	interestLimit := interest0[ccy]
	interestLimit.LeverageQuota = interest1[ccy].LeverageQuota
	interest0[ccy] = interestLimit
	return interest0, nil
}

func (client *RestClient) fetchInterestLimit(ccy string) (map[string]OkInterestLimit, error) {
	// https://www.okx.com/docs-v5/zh/#trading-account-rest-api-get-borrow-interest-and-limit
	queryDict := map[string]interface{}{}
	if ccy != "" {
		queryDict["ccy"] = ccy
	}
	queryDict["side"] = 2

	payload := utils.UrlEncodeParams(queryDict)
	url := "/api/v5/account/interest-limits"
	if len(payload) > 0 {
		url = url + "?" + payload
	}

	body, _, err := client.HttpRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ok get /api/v5/account/interest-limits err: %v", err)
	}

	var result OkInterestLimitResponse
	if err := sonic.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal interest limit response err: %v \n %v", err, body)
	}

	interest := map[string]OkInterestLimit{}
	for _, debt := range result.Data {
		for _, record := range debt.Records{
			if record.Currency == ccy {
				interest[record.Currency] = record
			}
		}
	}
	return interest, nil
}

func (client *RestClient) fetchPositionTier(ccy string) (map[string]OkInterestLimit, error) {
	// 页面：https://www.okx.com/zh-hans/trade-market/position/margin
	// 接口：https://www.okx.com/docs-v5/zh/#public-data-rest-api-get-position-tiers
	queryDict := map[string]interface{}{
		"instType": "MARGIN",
		"tdMode":   "cross",
		"ccy":   ccy,
	}

	payload := utils.UrlEncodeParams(queryDict)
	url := "/api/v5/public/position-tiers"
	if len(payload) > 0 {
		url = url + "?" + payload
	}

	body, _, err := client.HttpRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ok get /api/v5/public/position-tiers err: %v", err)
	}

	var result PositionTierResponse
	if err := sonic.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal position tier response err: %v \n %v", err, string(body))
	}

	interest := map[string]OkInterestLimit{}
	interest[ccy] = OkInterestLimit{
		Currency: ccy,
		LeverageQuota: result.Data,
	}

	return interest, nil
}