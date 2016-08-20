package main

import "fmt"

func init() { RegisterCommand("balance", balance{}) }

type balance struct{}

var currSymbol = map[string]string{
	"GBP": "£",
	"USD": "$",
}

func (h balance) Help() string { return "returns your current balance" }
func (h balance) Autocomplete(line string, pos int) (string, []string, string) {
	return line, nil, ""
}

func (h balance) F(cmd string, quitCh chan bool) error {
	if client == nil {
		return fmt.Errorf("not logged in to mondo. login first!")
	}

	b, err := client.Balance(accountId)
	if err != nil {
		return err
	}

	sym, ok := currSymbol[b.Currency]
	if !ok {
		fmt.Printf("unrecognised currency '%v'\n", b.Currency)
	}

	fmt.Printf("your balance is %s%v - you've spent %s%v so far today!\n", sym, float64(b.Balance)/100, sym, float64(b.SpendToday)/100)
	return nil
}
