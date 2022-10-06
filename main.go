package main

import (
	"InstaBot/clients/tgclient"
	event_consumer "InstaBot/consumer/event-consumer"
	"InstaBot/events/telegram"
	"InstaBot/storage/files"
	"flag"
	"log"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 50
)

func main() {

	eventsProcessor := telegram.New(
		tgclient.New(tgBotHost, mustToken()),
		files.New(storagePath),
	)
	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)

	}

}

func mustToken() string {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to your bot",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token not received on app launch")
	}

	return *token
}
