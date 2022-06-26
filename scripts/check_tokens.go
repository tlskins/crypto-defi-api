package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
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
	AlertMsg := func(inTkn, tgtTkn *t.TokenInfo, bestPrice float64, decimals int) string {
		return fmt.Sprintf(
			"%s @ %s (1 %s = %s %s)",
			inTkn.Symbol,
			api.RoundToStr(bestPrice, decimals),
			tgtTkn.Symbol,
			fmt.Sprintf("%.4f", 1/bestPrice),
			inTkn.Symbol,
		)
	}

	// work

	trackers, err := store.GetTokenTrackers()
	if err != nil {
		log.Fatal(err)
	}
	for _, tracker := range trackers {
		if tracker.TokenInfo == nil {
			log.Fatalf("Token for tracker %s not found", tracker.Id)
		}

		if tracker.LastSnapshot == nil {
			tracker.LastSnapshot = make(map[string]*t.TokenSnapshot)
		}
		activeAlerts := map[string]string{}

		// get current quotes
		for _, tgtTkn := range tracker.TargetTokens() {
			log.Printf("Quoting %s to %s\n", tracker.TokenInfo.Symbol, tgtTkn.Symbol)
			params := map[string]string{
				"inputMint":  tracker.TokenInfo.Address,
				"outputMint": tgtTkn.Address,
				"amount":     fmt.Sprintf("%.0f", float64(tracker.InputAmount)*math.Pow(10, float64(tracker.TokenInfo.Decimals))),
				"slippage":   "0.2",
			}
			out := &t.JupResp{}
			if err = api.HttpGetRequest(client, "GET", "https://quote-api.jup.ag/v1/quote", nil, params, out); err != nil {
				log.Fatal(err)
			}

			bestQuote := out.BestQuote()
			bestPrice := bestQuote.Price(tgtTkn)
			lastSnap := tracker.LastSnapshot[tgtTkn.Symbol]

			// check quote against last snap settings
			lastSettings := tracker.LastSnapAlertSettings[tgtTkn.Symbol]
			if lastSettings != nil {
				if lastSnap == nil ||
					lastSettings.FixedPriceChange > 0 && bestPrice >= lastSnap.Price+lastSettings.FixedPriceChange ||
					lastSettings.FixedPriceChange > 0 && bestPrice <= lastSnap.Price-lastSettings.FixedPriceChange ||
					lastSettings.PctPriceChange > 0 && bestPrice >= lastSnap.Price*(1.0+lastSettings.PctPriceChange) ||
					lastSettings.PctPriceChange > 0 && bestPrice <= lastSnap.Price*(1.0-lastSettings.PctPriceChange) {

					activeAlerts[tgtTkn.Symbol] = AlertMsg(tracker.TokenInfo, tgtTkn, bestPrice, lastSettings.Decimals)
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

					activeAlerts[tgtTkn.Symbol] = AlertMsg(tracker.TokenInfo, tgtTkn, bestPrice, absSettings.Decimals)
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
