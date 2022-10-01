package InstaBot

import (
	"flag"
	"log"
)

func main() {

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
