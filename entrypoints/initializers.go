package entrypoints

import (
	"log"
	"os"

	dgo "github.com/bwmarrin/discordgo"

	s "github.com/my_projects/sol-arb-api/store"
)

func InitStore() (store *s.Store, err error) {
	log.Println("Init store...")
	mongoDBName := os.Getenv("DB_NAME")
	mongoHost := os.Getenv("DB_HOST")
	mongoUser := os.Getenv("DB_USER")
	mongoPwd := os.Getenv("DB_PWD")
	store, err = s.NewStore(mongoDBName, mongoHost, mongoUser, mongoPwd)
	return
}

func InitDiscordClient() (discClient *dgo.Session) {
	log.Println("Init discord...")
	discordTkn := os.Getenv("DISCORD_BOT_TOKEN")
	discClient, err := dgo.New("Bot " + discordTkn)
	if err != nil {
		log.Fatal(err)
	}
	discClient.Identify.Intents = dgo.IntentsDirectMessages
	return
}
