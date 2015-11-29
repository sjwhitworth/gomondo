package main

import (
	"os"

	log "github.com/cihub/seelog"
	"github.com/sjwhitworth/gomondo"
)

func main() {
	defer log.Flush()

	clientId := os.Getenv("MONDO_CLIENT_ID")
	clientSecret := os.Getenv("MONDO_CLIENT_SECRET")
	userName := os.Getenv("MONDO_USERNAME")
	password := os.Getenv("MONDO_PASSWORD")

	// Authenticate with Mondo, and return an authenticated MondoClient.
	client, err := mondo.Authenticate(clientId, clientSecret, userName, password)
	if err != nil {
		panic(err)
	}

	// Retrieve all of the accounts.
	acs, err := client.Accounts()
	if err != nil {
		panic(err)
	}

	// Grab our account ID.
	accountId := acs[0].ID

	if _, err := client.RegisterWebhook(accountId, "YOUR_URL_HERE"); err != nil {
		log.Errorf("Error registering webhook: %v", err)
	}
}
