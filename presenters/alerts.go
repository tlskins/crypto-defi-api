package presenters

import (
	"fmt"

	"github.com/my_projects/sol-arb-api/api"
	t "github.com/my_projects/sol-arb-api/types"
)

func AlertMsg(tracker *t.TokenTracker, tgtTkn *t.TokenInfo, bestPrice float64, decimals int) string {
	lastPrice := 0.0
	lastSnap := tracker.LastSnapshot[tgtTkn.Symbol]
	if lastSnap != nil {
		lastPrice = lastSnap.Price
	}
	direction := "UP"
	if bestPrice < lastPrice {
		direction = "DOWN"
	}
	lastPrice = lastPrice / float64(tracker.InputAmount)
	bestPrice = bestPrice / float64(tracker.InputAmount)
	return fmt.Sprintf(
		"%v %s %s from %s to %s (1 %s = %s %s)",
		tracker.InputAmount,
		tracker.TokenInfo.Symbol,
		direction,
		api.RoundToStr(lastPrice, decimals),
		api.RoundToStr(bestPrice, decimals),
		tgtTkn.Symbol,
		fmt.Sprintf("%.4f", 1/bestPrice),
		tracker.TokenInfo.Symbol,
	)
}
