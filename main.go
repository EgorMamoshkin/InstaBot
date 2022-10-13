package main

import (
	"InstaBot/clients/tgclient"
	event_consumer "InstaBot/consumer/event-consumer"
	"InstaBot/events/telegram"
	"InstaBot/storage/postgres"
	"context"
	"flag"
	"fmt"
	"log"
)

const (
	tgBotHost = "api.telegram.org"
	// storagePath = "storage"
	batchSize = 50
)

func main() {
	token, pass := mustTokenPass()

	dsn := fmt.Sprintf("postgres://postgres:%s@localhost:5432/instagramBotDB?sslmode=disable", pass)
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
		tgclient.New(tgBotHost, token),
		s,
	)
	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	consumer.Start(ctx)

}

func mustTokenPass() (string, string) {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to your bot",
	)
	pass := flag.String(
		"psql-pass",
		"",
		"SQL DB user password",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token not received on app launch")
	}

	return *token, *pass
}
