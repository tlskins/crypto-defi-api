package main

import (
	"log"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/my_projects/sol-arb-api/api"
	t "github.com/my_projects/sol-arb-api/types"
	uuid "github.com/satori/go.uuid"
)

func main() {
	var err error

	nodes := map[string]*t.TradeNode{
		"SOL": &t.TradeNode{
			Token: t.AllTokens["SOL"],
			Edges: make(map[string]*t.TradeEdge),
		},
		"USDC": &t.TradeNode{
			Token: t.AllTokens["USDC"],
			Edges: make(map[string]*t.TradeEdge),
		},
		"DUST": &t.TradeNode{
			Token: t.AllTokens["DUST"],
			Edges: make(map[string]*t.TradeEdge),
		},
	}

	edges := map[string]*t.TradeEdge{}

	graph := &t.TradeGraph{
		StartNode: nodes["USDC"],
		EndNode:   nodes["USDC"],
	}

	client := api.NewHttpClient()
	for _, fromNode := range nodes {
		graph.Nodes = append(graph.Nodes, fromNode)
		for _, toNode := range nodes {
			if fromNode == toNode {
				continue
			}
			log.Printf("Quoting from %s to %s\n", fromNode.Token.Name, toNode.Token.Name)
			params := map[string]string{
				"inputMint":  fromNode.Token.Mint,
				"outputMint": toNode.Token.Mint,
				"amount":     strconv.Itoa(fromNode.Token.DefaultAmt),
				"slippage":   "0.2",
			}
			jupResp := &t.JupResp{}
			if err = api.HttpGetRequest(client, "GET", "https://quote-api.jup.ag/v1/quote", nil, params, jupResp); err != nil {
				log.Fatal(err)
			}
			bestQuote := jupResp.BestQuote()
			edge := &t.TradeEdge{
				Id:       uuid.NewV4().String(),
				FromNode: fromNode,
				ToNode:   toNode,
				Quote:    bestQuote,
				PricePer: bestQuote.Price(),
			}

			edges[edge.Id] = edge
			fromNode.Edges[edge.Id] = edge
			toNode.Edges[edge.Id] = edge
		}
	}

	spew.Dump(graph)

	log.Println("Finished...")
}
