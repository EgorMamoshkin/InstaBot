package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/EgorMamoshkin/InstaBot/auth/authserver"
	"github.com/EgorMamoshkin/InstaBot/auth/handler"
	"github.com/EgorMamoshkin/InstaBot/clients/tgclient"
	event_consumer "github.com/EgorMamoshkin/InstaBot/consumer/event-consumer"
	"github.com/EgorMamoshkin/InstaBot/events/telegram"
	"github.com/EgorMamoshkin/InstaBot/storage/postgres"
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

	tg := tgclient.New(tgBotHost, token)

	respHandler := handler.New(telegram.InstagramAPI, tg, s)
	server := authserver.New(respHandler)

	go func() {
		err = server.StartLS()
		if err != nil {
			log.Printf("server crashed: %s", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Init(ctx); err != nil {
		log.Fatal("can't init new table:", err)
	}

	eventsProcessor := telegram.New(
		tg,
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
