package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/EgorMamoshkin/InstaBot/apiclient/instagramapi"
	"github.com/EgorMamoshkin/InstaBot/auth/authserver"
	"github.com/EgorMamoshkin/InstaBot/clients/tgclient"
	event_consumer "github.com/EgorMamoshkin/InstaBot/consumer/event-consumer"
	"github.com/EgorMamoshkin/InstaBot/events/telegram"
	"github.com/EgorMamoshkin/InstaBot/storage/postgres"
	"github.com/gorilla/mux"
	"log"
)

const (
	tgBotHost = "api.telegram.org"

	// storagePath = "storage"
	batchSize    = 50
	InstAuthHost = "api.instagram.com"
	InstApiHost  = "graph.instagram.com"
	RedirectURI  = "https://188.225.60.154:8080/auth"
)

func main() {
	token, pass, appID, appSecret := mustTokenPass()

	dsn := fmt.Sprintf("postgres://postgres:%s@localhost:5432/instagramBotDB?sslmode=disable", pass)

	st, err := postgres.New(dsn)
	if err != nil {
		log.Fatal("can't connect to storage", err)
	}

	server := authserver.New(mux.NewRouter())

	go func() {
		err = server.StartLS()
		if err != nil {
			log.Printf("server crashed: %s", err)
		}
	}()

	apiClient := instagramapi.New(InstApiHost, InstAuthHost, appID, appSecret, RedirectURI)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := st.Init(ctx); err != nil {
		log.Fatal("can't init new table:", err)
	}

	eventsProcessor := telegram.New(
		tgclient.New(tgBotHost, token),
		apiClient,
		st,
	)
	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	consumer.Start(ctx)

}

func mustTokenPass() (string, string, string, string) {
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

	appID := flag.String(
		"app-id",
		"",
		"app id registered in instagram for developers",
	)

	appSecret := flag.String(
		"app-secret",
		"",
		"app secret code, registered in instagram for developers",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token not received on app launch")
	}

	return *token, *pass, *appID, *appSecret
}
