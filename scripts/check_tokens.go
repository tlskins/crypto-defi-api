package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"github.com/my_projects/sol-arb-api/api"
	s "github.com/my_projects/sol-arb-api/store"
	t "github.com/my_projects/sol-arb-api/types"
)

func main() {
	// init
	var err error
	configPath := flag.String("config", "../config.prod.yml", "path for config")
	flag.Parse()
	if err := godotenv.Load(*configPath); err != nil {
		log.Fatalln("Error loading config:", err)
	}
	mongoDBName := os.Getenv("DB_NAME")
	mongoHost := os.Getenv("DB_HOST")
	mongoUser := os.Getenv("DB_USER")
	mongoPwd := os.Getenv("DB_PWD")
	store, err := s.NewStore(mongoDBName, mongoHost, mongoUser, mongoPwd)
	if err != nil {
		log.Printf("Err creating store: %v\n", err)
	}
	client := api.NewHttpClient()

	// FOXY @ 0.0034 (1 USDC = 281 FOXY)
	AlertMsg := func(tracker *t.TokenTracker, tgtTkn *t.TokenInfo, bestPrice float64, decimals int) string {
		return fmt.Sprintf(
			"%v %s @ %s (1 %s = %s %s)",
			tracker.InputAmount,
			tracker.TokenInfo.Symbol,
			api.RoundToStr(bestPrice, decimals),
			tgtTkn.Symbol,
			fmt.Sprintf("%.4f", 1/bestPrice),
			tracker.TokenInfo.Symbol,
		)
	}

	// work

	trackers, err := store.GetTokenTrackers()
	if err != nil {
		log.Fatal(err)
	}

	// aggregate all quotes needed
	quotes := map[*t.TokenInfo]map[*t.TokenInfo]map[int]float64{}
	for _, tracker := range trackers {
		for _, tgtTkn := range tracker.TargetTokens() {
			if quotes[tracker.TokenInfo] == nil {
				quotes[tracker.TokenInfo] = make(map[*t.TokenInfo]map[int]float64)
			}
			if quotes[tracker.TokenInfo][tgtTkn] == nil {
				quotes[tracker.TokenInfo][tgtTkn] = make(map[int]float64)
			}
			quotes[tracker.TokenInfo][tgtTkn][tracker.InputAmount] = 0.0
		}
	}

	// multi thread get quotes
	var wg sync.WaitGroup
	for inTkn, outMap := range quotes {
		for outTkn, outQuote := range outMap {
			for inputAmount := range outQuote {
				wg.Add(1)
				go func(inTkn, outTkn *t.TokenInfo, inputAmount int, outQuote *map[int]float64) {
					defer wg.Done()
					log.Printf("Quoting %s to %s\n", inTkn.Symbol, outTkn.Symbol)
					params := map[string]string{
						"inputMint":  inTkn.Address,
						"outputMint": outTkn.Address,
						"amount":     fmt.Sprintf("%.0f", float64(inputAmount)*math.Pow(10, float64(inTkn.Decimals))),
						"slippage":   "0.2",
					}
					out := &t.JupResp{}
					if err = api.HttpGetRequest(client, "GET", "https://quote-api.jup.ag/v1/quote", nil, params, out); err != nil {
						log.Fatal(err)
					}

					(*outQuote)[inputAmount] = out.BestQuote().Price(outTkn)
				}(inTkn, outTkn, inputAmount, &outQuote)
			}
		}
	}
	wg.Wait()

	// check quotes against alerts
	for _, tracker := range trackers {
		if tracker.LastSnapshot == nil {
			tracker.LastSnapshot = make(map[string]*t.TokenSnapshot)
		}
		activeAlerts := map[string]string{}

		for _, tgtTkn := range tracker.TargetTokens() {
			lastSnap := tracker.LastSnapshot[tgtTkn.Symbol]
			bestPrice := quotes[tracker.TokenInfo][tgtTkn][tracker.InputAmount]

			// check quote against last snap settings
			lastSettings := tracker.LastSnapAlertSettings[tgtTkn.Symbol]
			if lastSettings != nil {
				if lastSnap == nil ||
					lastSettings.FixedPriceChange > 0 && bestPrice >= lastSnap.Price+lastSettings.FixedPriceChange ||
					lastSettings.FixedPriceChange > 0 && bestPrice <= lastSnap.Price-lastSettings.FixedPriceChange ||
					lastSettings.PctPriceChange > 0 && bestPrice >= lastSnap.Price*(1.0+lastSettings.PctPriceChange) ||
					lastSettings.PctPriceChange > 0 && bestPrice <= lastSnap.Price*(1.0-lastSettings.PctPriceChange) {

					activeAlerts[tgtTkn.Symbol] = AlertMsg(tracker, tgtTkn, bestPrice, lastSettings.Decimals)
					tracker.LastSnapshot[tgtTkn.Symbol] = &t.TokenSnapshot{
						TokenInfo: tgtTkn,
						Price:     bestPrice,
						At:        time.Now(),
					}
					continue
				}
			}

			// check quote against absolute settings
			absSettings := tracker.AbsoluteAlertSettings[tgtTkn.Symbol]
			if absSettings != nil {
				if bestPrice >= absSettings.PriceAbove ||
					bestPrice <= absSettings.PriceBelow {

					activeAlerts[tgtTkn.Symbol] = AlertMsg(tracker, tgtTkn, bestPrice, absSettings.Decimals)
					tracker.LastSnapshot[tgtTkn.Symbol] = &t.TokenSnapshot{
						TokenInfo: tgtTkn,
						Price:     bestPrice,
						At:        time.Now(),
					}
					continue
				}
			}
		}

		spew.Dump(activeAlerts)
		if _, err = store.UpsertTokenTracker(tracker); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Finished...")
}
