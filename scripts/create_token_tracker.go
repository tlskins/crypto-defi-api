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

	// sol tracker
	if _, err := store.UpsertTokenTracker(&t.TokenTracker{
		TokenInfo:    tokenMap["SOL"],
		DiscordId:    "timchi#0831",
		InputAmount:  1,
		LastSnapshot: make(map[string]*t.TokenSnapshot),
		LastSnapAlertSettings: map[string]*t.LastSnapAlertSettings{
			"USDC": &t.LastSnapAlertSettings{
				TargetToken:      tokenMap["USDC"],
				Decimals:         2,
				FixedPriceChange: 1.0,
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	// dust tracker
	if _, err := store.UpsertTokenTracker(&t.TokenTracker{
		TokenInfo:    tokenMap["DUST"],
		DiscordId:    "timchi#0831",
		InputAmount:  1,
		LastSnapshot: make(map[string]*t.TokenSnapshot),
		LastSnapAlertSettings: map[string]*t.LastSnapAlertSettings{
			"USDC": &t.LastSnapAlertSettings{
				TargetToken:      tokenMap["USDC"],
				Decimals:         2,
				FixedPriceChange: 0.1,
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	// foxy tracker
	if _, err := store.UpsertTokenTracker(&t.TokenTracker{
		TokenInfo:    tokenMap["FOXY"],
		DiscordId:    "timchi#0831",
		InputAmount:  1,
		LastSnapshot: make(map[string]*t.TokenSnapshot),
		LastSnapAlertSettings: map[string]*t.LastSnapAlertSettings{
			"USDC": &t.LastSnapAlertSettings{
				TargetToken:      tokenMap["USDC"],
				Decimals:         4,
				FixedPriceChange: 0.04,
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	log.Println("Finished...")
}
