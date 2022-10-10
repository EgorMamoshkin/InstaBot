package main

import (
	"InstaBot/clients/tgclient"
	event_consumer "InstaBot/consumer/event-consumer"
	"InstaBot/events/telegram"
	"InstaBot/storage/postgres"
	"context"
	"flag"
	"log"
)

const (
	tgBotHost = "api.telegram.org"
	// storagePath = "storage"
	batchSize = 50
)

func main() {
	dsn := "postgres://postgres:Link895olN@localhost:5432/instagramBotDB?sslmode=disable"
	s, err := postgres.New(dsn)
	if err != nil {
		log.Fatal("can't connect to storage", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Init(ctx); err != nil {
		log.Fatal("can't init new table:", err)
	}

	eventsProcessor := telegram.New(
		tgclient.New(tgBotHost, mustToken()),
		s,
	)
	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(ctx); err != nil {
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
