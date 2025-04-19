package okxv5

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cybernonce/gotrader/trader/types"
)

type OkBalance struct {
	AdjEq       string             `json:"adjEq"` // 美金层面有效保证金
	BorrowFroz  string             `json:"borrowFroz"`
	Details     []*OkBalanceDetail `json:"details"`
	Imr         string             `json:"imr"` // 占用保证金
	IsoEq       string             `json:"isoEq"`
	MgnRatio    string             `json:"mgnRatio"`
	Mmr         string             `json:"mmr"`
	NotionalUsd string             `json:"notionalUsd"` // 仓位美金价值
	OrdFroz     string             `json:"ordFroz"`
	TotalEq     string             `json:"totalEq"`
	UTime       string             `json:"uTime"`
}

// 去掉了一些不太可能使用的字段
type OkBalanceDetail struct {
	AvailBal      string `json:"availBal"`
	AvailEq       string `json:"availEq"`
	CashBal       string `json:"cashBal"`
	Ccy           string `json:"ccy"`
	CrossLiab     string `json:"crossLiab"`
	Eq            string `json:"eq"` // 币种总权益
	EqUsd         string `json:"eqUsd"`
	FixedBal      string `json:"fixedBal"`
	FrozenBal     string `json:"frozenBal"`
	IsoEq         string `json:"isoEq"`
	IsoLiab       string `json:"isoLiab"`
	IsoUpl        string `json:"isoUpl"`
	Imr           string `json:"imr"`
	MgnRatio      string `json:"mgnRatio"`
	NotionalLever string `json:"notionalLever"`
	OrdFrozen     string `json:"ordFrozen"`
	UTime         string `json:"uTime"`
	Upl           string `json:"upl"`
	UplLiab       string `json:"uplLiab"`
}

func (b OkBalanceDetail) ToAssets() types.Asset {
	free, _ := strconv.ParseFloat(b.CashBal, 64)
	frozen, _ := strconv.ParseFloat(b.FrozenBal, 64)
	total, _ := strconv.ParseFloat(b.Eq, 64)
	eqUsd, _ := strconv.ParseFloat(b.EqUsd, 64)
	return types.Asset{
		Coin:   b.Ccy,
		Free:   free,
		Frozen: frozen,
		Total:  total,
		EqUsd:  eqUsd,
	}
}

type BalanceRsp struct {
	BaseOkRsp
	Data []*OkBalance `json:"data"`
}

func (t *BalanceRsp) valid() bool {
	return t.Code == "0" && len(t.Data) > 0
}

func (client *RestClient) FetchBalance() (*types.Assets, error) {
	url := FetchBalanceUri
	body, _, err := client.HttpRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("ok get /api/v5/account/balance err:%v", err)
		return nil, err
	}

	response := new(BalanceRsp)
	if err = json.Unmarshal(body, response); err != nil {
		log.Errorf("ok get /api/v5/account/balance parser err:%v", err)
		return nil, err
	}

	if !response.valid() {
		err := fmt.Errorf("ok get /api/v5/account/balance fail, code:%s, msg:%s", response.Code, response.Msg)
		return nil, err
	}

	if len(response.Data) == 0 {
		err := fmt.Errorf("ok get /api/v5/account/balance empty")
		return nil, err
	}

	result, err := balanceTransform(response)
	if err != nil {
		err := fmt.Errorf("ok get /api/v5/account/balance transform err:%s", err)
		return nil, err
	}
	return result, nil
}

func balanceTransform(response *BalanceRsp) (*types.Assets, error) {
	bal := response.Data[0]
	assets := make(map[string]types.Asset, len(bal.Details))
	for _, a := range bal.Details {
		assets[a.Ccy] = a.ToAssets()
	}
	totalEq, _ := strconv.ParseFloat(bal.TotalEq, 64)
	uniMMr, _ := strconv.ParseFloat(bal.MgnRatio, 64)
	// imr, _ := strconv.ParseFloat(bal.Imr, 64)
	if (uniMMr == 0) && (totalEq > 0) {
		uniMMr = 20
	}

	// account margin
	adjEq, _ := strconv.ParseFloat(bal.AdjEq, 64)
	notionalUsd, _ := strconv.ParseFloat(bal.NotionalUsd, 64)
	imr, _ := strconv.ParseFloat(bal.Imr, 64)
	accountMargin := notionalUsd / adjEq
	freeUsdEq := adjEq - imr
	borrowed, _ := strconv.ParseFloat(bal.BorrowFroz, 64)
	return &types.Assets{
		Assets:        assets,
		TotalUsdEq:    totalEq,
		FreeUsdEq:     freeUsdEq,
		UniMMR:        uniMMr,
		AccountMargin: accountMargin,
		Borrowed:      borrowed,
	}, nil
}
