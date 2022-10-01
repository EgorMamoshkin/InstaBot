package InstaBot

import (
	"flag"
	"log"
)

const (
	tgBotHost = "api.telegram.org"
)

func main() {
	// tgClient := tg_client.New(tgBotHost, mustToken())

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
