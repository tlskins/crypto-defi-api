package main

import (
	"fmt"
	"log"
	"math"

	"github.com/davecgh/go-spew/spew"
	"github.com/my_projects/sol-arb-api/api"
	t "github.com/my_projects/sol-arb-api/types"
)

var inTkn = "SOL"
var outTkn = "USDC"
var amount = 1

func main() {
	var err error
	client := api.NewHttpClient()

	tokenList := []*t.TokenInfo{}
	if err = api.HttpGetRequest(client, "GET", "https://cache.jup.ag/tokens", nil, nil, &tokenList); err != nil {
		log.Fatal(err)
	}
	tokenMap := make(map[string]*t.TokenInfo)
	for _, tkn := range tokenList {
		tokenMap[tkn.Symbol] = tkn
	}

	inTkn := tokenMap[inTkn]
	if inTkn == nil {
		log.Fatalf("Token not found for %s", inTkn)
	}
	outTkn := tokenMap[outTkn]
	if outTkn == nil {
		log.Fatalf("Token not found for %s", outTkn)
	}

	params := map[string]string{
		"inputMint":  inTkn.Address,
		"outputMint": outTkn.Address,
		"amount":     fmt.Sprintf("%.0f", float64(amount)*math.Pow(10, float64(inTkn.Decimals))),
		"slippage":   "0.2",
	}
	out := &t.JupResp{}
	if err = api.HttpGetRequest(client, "GET", "https://quote-api.jup.ag/v1/quote", nil, params, out); err != nil {
		log.Fatal(err)
	}
	spew.Dump(out.BestQuote())
	log.Printf("in %v %s out %v %s", amount, inTkn.Symbol, float64(out.BestQuote().OutAmt)/math.Pow(10, float64(outTkn.Decimals)), outTkn.Symbol)

	log.Println("Finished...")
}
