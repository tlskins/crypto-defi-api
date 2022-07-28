package types

import "time"

type TokenTracker struct {
	Id                    string                            `bson:"_id" json:"id"`
	TokenInfo             *TokenInfo                        `bson:"tokenInfo" json:"tokenInfo"`
	DiscordId             string                            `bson:"discordId" json:"discordId"`
	InputAmount           int                               `json:"inputAmount" json:"inputAmount"`
	LastSnapshot          map[string]*TokenSnapshot         `bson:"lastSnapshot" json:"lastSnapshot"`
	LastSnapAlertSettings map[string]*LastSnapAlertSettings `bson:"lastSnapAlertSettings" json:"lastSnapAlertSettings"`
	AbsoluteAlertSettings map[string]*AbsoluteAlertSettings `bson:"absoluteAlertSettings" json:"absoluteAlertSettings"`
}

func (t TokenTracker) TargetTokens() (out map[string]*TokenInfo) {
	out = make(map[string]*TokenInfo)
	for _, settings := range t.LastSnapAlertSettings {
		out[settings.TargetToken.Symbol] = settings.TargetToken
	}
	for _, settings := range t.AbsoluteAlertSettings {
		out[settings.TargetToken.Symbol] = settings.TargetToken
	}
	return
}

type TokenSnapshot struct {
	TokenInfo *TokenInfo `bson:"tokenInfo" json:"tokenInfo"`
	Price     float64    `bson:"price" json:"price"`
	At        time.Time  `bson:"at" json:"at"`
}

type LastSnapAlertSettings struct {
	TargetToken             *TokenInfo `bson:"targetToken" json:"targetToken"`
	Decimals                int        `bson:"decimals" json:"decimals"`
	FixedPriceChange        float64    `bson:"fixedPriceChange,omitempty" json:"fixedPriceChange,omitempty"`
	InvertedFixedPriceAlert bool       `bson:"invertedFixedPriceAlert,omitempty" json:"invertedFixedPriceAlert,omitempty"`
	PctPriceChange          float64    `bson:"percentPriceChange,omitempty" json:"percentPriceChange,omitempty"`
}

type AbsoluteAlertSettings struct {
	TargetToken *TokenInfo `bson:"targetToken" json:"targetToken"`
	Decimals    int        `bson:"decimals" json:"decimals"`
	PriceAbove  float64    `bson:"priceAbove,omitempty" json:"priceAbove,omitempty"`
	PriceBelow  float64    `bson:"priceBelow,omitempty" json:"priceBelow,omitempty"`
}

// type Token struct {
// 	Name       string  `json:"name"`
// 	Mint       string  `json:"mint"`
// 	USD        float64 `json:"usd"`
// 	DefaultAmt int     `json:"defaultAmt"`
// }

// var AllTokens = map[string]*Token{
// 	"SOL": {
// 		Name:       "SOL",
// 		Mint:       "So11111111111111111111111111111111111111112",
// 		DefaultAmt: 10,
// 	},
// 	"DUST": {
// 		Name:       "DUST",
// 		Mint:       "DUSTawucrTsGU8hcqRdHDCbuYhCPADMLM2VcCb8VnFnQ",
// 		DefaultAmt: 500,
// 	},
// 	"USDC": {
// 		Name:       "USDC",
// 		Mint:       "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
// 		DefaultAmt: 250,
// 	},
// }
