package main

import (
	"flag"
	"log"
	"os"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

func main() {
	var err error

	configPath := flag.String("config", "../config.prod.yml", "path for config")
	flag.Parse()
	if err := godotenv.Load(*configPath); err != nil {
		log.Fatalln("Error loading config:", err)
	}

	userId := os.Getenv("ALERTER_DISCORD_ID")
	discordTkn := os.Getenv("DISCORD_BOT_TOKEN")

	spew.Dump(userId, discordTkn)

	discClient, err := dgo.New("Bot " + discordTkn)
	if err != nil {
		log.Fatal(err)
	}
	discClient.Identify.Intents = dgo.IntentsDirectMessages

	dmChannel, err := discClient.UserChannelCreate(userId)
	if err != nil {
		log.Fatal(err)
	}
	msg, err := discClient.ChannelMessageSend(dmChannel.ID, "test dm!")
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(msg)

	log.Println("Finished...")
}
