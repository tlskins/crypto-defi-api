package types

import "math"

type JupResp struct {
	Data []*Quote `json:"data"`
}

func (r JupResp) BestQuote() (best *Quote) {
	if len(r.Data) == 0 {
		return
	}
	return r.Data[0]
}

type Quote struct {
	InAmt          int           `json:"inAmount"`
	OutAmt         int           `json:"outAmount"`
	Amount         int           `json:"amount"`
	OtherAmtThresh int           `json:"otherAmountThreshold"`
	OutAmtSlippage int           `json:"outAmountWithSlippage"`
	SwapMode       string        `json:"swapMode"`
	PriceImpactPct float64       `json:"priceImpactPct"`
	MarketInfos    []*MarketInfo `json:"marketInfos"`
}

func (q Quote) Price(outTkn *TokenInfo) float64 {
	return float64(q.OutAmt) / math.Pow(10, float64(outTkn.Decimals))
}

type MarketInfo struct {
	Id           string  `json:"id"`
	Label        string  `json:"label"`
	InputMint    string  `json:"inputMint"`
	Output       string  `json:"outputMint"`
	NotEnoughLiq bool    `json:"notEnoughLiquidity"`
	InAmt        int     `json:"inAmount"`
	OutAmt       int     `json:"outAmount"`
	PriceImpact  float64 `json:"priceImpactPct"`
	LpFee        *Fee    `json:"lpFee"`
	PlatformFee  *Fee    `json:"platformFee"`
}

type Fee struct {
	Amount float64 `json:"amount"`
	Mint   string  `json:"mint"`
	Pct    float64 `json:"pct"`
}

type TokenInfo struct {
	ChainId  int      `json:"chainId"`
	Address  string   `json:"address"`
	Symbol   string   `json:"symbol"`
	Name     string   `json:"name"`
	Decimals int      `json:"decimals"`
	LogoUri  string   `json:"logoURI"`
	Tags     []string `json:"tags"`
}
