package main

import (
	"fmt"
	"os"

	log "github.com/cihub/seelog"
	"github.com/olekukonko/tablewriter"
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

	log.Infof("Authenticated with Mondo successfully!")

	// Retrieve all of the accounts.
	acs, err := client.Accounts()
	if err != nil {
		panic(err)
	}

	if len(acs) == 0 {
		log.Errorf("No accounts with Mondo found :( Sign up!")
		return
	}

	// Grab our account ID.
	accountId := acs[0].ID

	// Get all transactions. You can also get a specific transaction by ID.
	transactions, err := client.Transactions(accountId, "", "", 100)
	if err != nil {
		panic(err)
	}

	if len(transactions) == 0 {
		log.Warnf("No transactions found. Sorry!")
		return
	}

	// Render a lovely table of all of your transactions.
	table := transactionsToTable(transactions...)
	table.Render()

	// Create a feed item in your feed.
	// Don't run this, unless you want to spam yourself, as there is no way currently to delete a feed item.
	// err = client.CreateFeedItem(id, "Hi there!", "https://blog.golang.org/gopher/gopher.png", "",
	// t "", "", "This is a test item, from Go")
	// if err != nil {
	// 	panic(err)
	// }
}

func transactionsToTable(transactions ...mondo.Transaction) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Time", "Merchant Name", "Amount", "Category", "Balance"})
	for _, v := range transactions {
		if v.Category == "mondo" {
			v.Merchant.Name = "Mondo"
		}
		table.Append([]string{v.ID, v.Created, v.Merchant.Name, fmt.Sprintf("%v", v.Amount), v.Category, fmt.Sprintf("£%.2f", float64(v.AccountBalance)/100)})
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	return table
}
