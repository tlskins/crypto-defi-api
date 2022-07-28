package tokentracker

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"

	"github.com/my_projects/sol-arb-api/api"
	pr "github.com/my_projects/sol-arb-api/presenters"
	s "github.com/my_projects/sol-arb-api/store"
	t "github.com/my_projects/sol-arb-api/types"
)

func CheckTokens(
	store *s.Store,
	client *http.Client,
	discClient *discordgo.Session,
	sendAlerts bool,
	saveTrackers bool,
) (err error) {
	// get trackers
	log.Println("Getting token trackers...")
	var trackers []*t.TokenTracker
	if trackers, err = store.GetTokenTrackers(); err != nil {
		return
	}
	log.Printf("Found %v trackers\n", len(trackers))

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
	log.Println("Getting quotes...")
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
					if sErr := api.HttpGetRequest(client, "GET", "https://quote-api.jup.ag/v1/quote", nil, params, out); sErr != nil {
						spew.Dump(sErr)
					} else {
						(*outQuote)[inputAmount] = out.BestQuote().Price(outTkn)
						log.Printf("Quoted: %v %s @ %f %s\n", inputAmount, inTkn.Symbol, (*outQuote)[inputAmount], outTkn.Symbol)
					}
				}(inTkn, outTkn, inputAmount, &outQuote)
			}
		}
	}
	wg.Wait()

	log.Println("Checking against alerts...")
	// check quotes against alerts
	for _, tracker := range trackers {
		log.Printf("Checking tracker %s for %s\n", tracker.Id, tracker.TokenInfo.Symbol)
		if tracker.LastSnapshot == nil {
			tracker.LastSnapshot = make(map[string]*t.TokenSnapshot)
		}
		activeAlerts := map[string]string{}

		for _, tgtTkn := range tracker.TargetTokens() {
			log.Printf("Checking target token %s\n", tgtTkn.Symbol)
			lastPrice := 0.0
			lastSnap := tracker.LastSnapshot[tgtTkn.Symbol]
			if lastSnap != nil {
				lastPrice = lastSnap.Price
			}
			bestPrice := quotes[tracker.TokenInfo][tgtTkn][tracker.InputAmount]

			// check quote against last snap settings
			lastSettings := tracker.LastSnapAlertSettings[tgtTkn.Symbol]
			if lastSettings != nil {
				log.Println("Checking against last settings")
				currBestPrice := bestPrice
				lastBestPrice := lastPrice
				if lastSettings.InvertedFixedPriceAlert {
					currBestPrice = 1 / bestPrice
					lastBestPrice = 1 / lastPrice
				}
				if lastSnap == nil ||
					lastSettings.FixedPriceChange > 0 && currBestPrice >= lastBestPrice+lastSettings.FixedPriceChange ||
					lastSettings.FixedPriceChange > 0 && currBestPrice <= lastBestPrice-lastSettings.FixedPriceChange ||
					lastSettings.PctPriceChange > 0 && currBestPrice >= lastBestPrice*(1.0+lastSettings.PctPriceChange) ||
					lastSettings.PctPriceChange > 0 && currBestPrice <= lastBestPrice*(1.0-lastSettings.PctPriceChange) {
					activeAlerts[tgtTkn.Symbol] = pr.AlertMsg(tracker, tgtTkn, bestPrice, lastSettings.Decimals)
					tracker.LastSnapshot[tgtTkn.Symbol] = &t.TokenSnapshot{
						TokenInfo: tgtTkn,
						Price:     bestPrice,
						At:        time.Now(),
					}
					log.Printf("Alerting %s\n", activeAlerts[tgtTkn.Symbol])
					continue
				}
			}

			// check quote against absolute settings
			absSettings := tracker.AbsoluteAlertSettings[tgtTkn.Symbol]
			if absSettings != nil {
				log.Println("Checking against abs settings")
				if bestPrice >= absSettings.PriceAbove ||
					bestPrice <= absSettings.PriceBelow {

					activeAlerts[tgtTkn.Symbol] = pr.AlertMsg(tracker, tgtTkn, bestPrice, absSettings.Decimals)
					tracker.LastSnapshot[tgtTkn.Symbol] = &t.TokenSnapshot{
						TokenInfo: tgtTkn,
						Price:     bestPrice,
						At:        time.Now(),
					}
					log.Printf("Alerting %s\n", activeAlerts[tgtTkn.Symbol])
					continue
				}
			}
		}

		if sendAlerts {
			// send alert dms
			log.Println("Send alert dms...")
			spew.Dump(activeAlerts)
			for _, alertMsg := range activeAlerts {
				var dmChannel *discordgo.Channel
				if dmChannel, err = discClient.UserChannelCreate(tracker.DiscordId); err != nil {
					return
				}
				var msg *discordgo.Message
				if msg, err = discClient.ChannelMessageSend(dmChannel.ID, alertMsg); err != nil {
					return
				}
				spew.Dump(msg)
			}
		}

		if saveTrackers {
			// update snapshot
			if _, err = store.UpsertTokenTracker(tracker); err != nil {
				return
			}
		}
	}
	return
}
