package okxv5

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/wsg011/gotrader/pkg/utils"
)

type OkInterestLimit struct {
	Currency      string `json:"ccy"`  
	
	// 母账户维度借币限额, 如果已配置可用额度，该字段代表当前交易账户的借币限额       
	Quota       string `json:"loanQuota"`   
	// 母子账户已借额度, 如果已配置可用额度，该字段代表当前交易账户的已借额度
	Used      string `json:"usedLmt"`
	// 杠杆下的币对限额
	LeverageQuota      string `json:"leveragedQuota"`
}

type Debt struct {
	Records []OkInterestLimit `json:"records"`
}

type OkInterestLimitResponse struct {
	Code string            `json:"code"`
	Data []Debt            `json:"data"`
}

func (client *RestClient) FetchInterestLimit(ccy string) (map[string]OkInterestLimit, error) {
	// https://www.okx.com/docs-v5/zh/#trading-account-rest-api-get-borrow-interest-and-limit
	queryDict := map[string]interface{}{}
	if ccy != "" {
		queryDict["ccy"] = ccy
	}
	queryDict["side"] = 2

	payload := utils.UrlEncodeParams(queryDict)
	url := RestUrl + "/api/v5/account/interest-limits"
	if len(payload) > 0 {
		url = url + "?" + payload
	}
	fmt.Println(fmt.Sprintf("url:%s", url))

	body, _, err := client.HttpRequest("GET", "/api/v5/account/interest-limits", nil)
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
