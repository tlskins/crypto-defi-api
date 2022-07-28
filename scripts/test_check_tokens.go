package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
	"github.com/my_projects/sol-arb-api/api"
	h "github.com/my_projects/sol-arb-api/entrypoints"
	c "github.com/my_projects/sol-arb-api/services/token_tracker"
	s "github.com/my_projects/sol-arb-api/store"
)

func main() {
	var err error
	configPath := flag.String("config", "../config.prod.yml", "path for config")
	flag.Parse()
	if err := godotenv.Load(*configPath); err != nil {
		log.Fatalln("Error loading config:", err)
	}

	var store *s.Store
	store, err = h.InitStore()
	defer store.Close()
	if err != nil {
		log.Fatal(err)
	}

	discClient := h.InitDiscordClient()
	client := api.NewHttpClient()
	err = c.CheckTokens(store, client, discClient, false, false)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Finished...")
}
