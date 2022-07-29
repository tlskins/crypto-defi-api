package main

import (
	"flag"
	"log"
	"os"

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
	// userId := os.Getenv("ALERTER_DISCORD_ID")
	client := api.NewHttpClient()

	// work
	tokenList := []*t.TokenInfo{}
	if err = api.HttpGetRequest(client, "GET", "https://cache.jup.ag/tokens", nil, nil, &tokenList); err != nil {
		log.Fatal(err)
	}
	tokenMap := make(map[string]*t.TokenInfo)
	for _, tkn := range tokenList {
		tokenMap[tkn.Symbol] = tkn
	}

	// get existing tracker
	tracker, err := store.GetTokenTracker("6c1c638c-aea8-40bc-878d-e0199031614a")
	if err != nil {
		log.Fatal(err)
	}

	// add settings
	tracker.LastSnapAlertSettings["ETH"] = &t.LastSnapAlertSettings{
		TargetToken:      tokenMap["ETH"],
		Decimals:         2,
		FixedPriceChange: 50.0,
	}

	// usdc tracker
	if _, err := store.UpsertTokenTracker(tracker); err != nil {
		log.Fatal(err)
	}

	log.Println("Finished...")
}
