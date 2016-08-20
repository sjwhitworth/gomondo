package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	mondo "github.com/sjwhitworth/gomondo"
)

func init() {
	RegisterCommand("ls", &transactions{
		ts: make([]*mondo.Transaction, 0),
	})
}

type transactions struct {
	ts []*mondo.Transaction
}

func (h *transactions) Help() string {
	return "list transactions. this command takes a list of key value pairs for filtering.\n\tyou can filter on transaction fields, sorting order and number of results returned\n\te.g.:\n\n\tls sort=desc n=100\n\tls sort=asc n=10 category=eating_out merchant=pret"
}
func (h *transactions) Autocomplete(line string, pos int) (string, []string, string) {
	return line, nil, ""
}

type ByTimeDescending []*mondo.Transaction

func (s ByTimeDescending) Len() int      { return len(s) }
func (s ByTimeDescending) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByTimeDescending) Less(i, j int) bool {
	return s[i].Created.After(s[j].Created)
}

func (h *transactions) F(cmd string, quitCh chan bool) error {
	if client == nil {
		return fmt.Errorf("not logged in to mondo. login first!")
	}

	// don't already have in memory, need to load up for the first time
	if len(h.ts) == 0 {
		fmt.Println("loading all of your transactions. this may take a few seconds..")
		// since is used as a pagination token. we want to grab all of the transactions, so set this to an empty string.
		since := ""
		for {
			ts, err := client.Transactions(accountId, since, "", 100)
			if err != nil {
				return err
			}
			h.ts = append(h.ts, ts...)
			if len(ts) != 100 {
				break
			}
			since = ts[len(ts)-1].ID
		}
	}

	filtered := make([]*mondo.Transaction, len(h.ts))
	copy(filtered, h.ts)

	var err error
	n := 1000

	// we can filter by key value pairs here, so parse them out. no reflection because I am lazy
	fields := strings.Split(cmd, " ")
	if len(fields) > 1 {
		for _, v := range fields[1:] {
			s := strings.Split(v, "=")
			if len(s) != 2 {
				return fmt.Errorf("invalid key=value parameter")
			}

			key, value := s[0], s[1]

			switch key {
			case "sort":
				if value == "desc" {
					// default is ascending so don't need to do anything in the other case
					sort.Sort(ByTimeDescending(filtered))
				}

			case "category":
				filtered = filterTransactions(filtered, func(t *mondo.Transaction) bool {
					return strings.Contains(t.Category, value)
				})

			case "merchant":
				filtered = filterTransactions(filtered, func(t *mondo.Transaction) bool {
					return strings.Contains(strings.ToLower(t.Merchant.Name), strings.ToLower(value))
				})

			case "n":
				n, err = strconv.Atoi(value)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(filtered) == 0 {
		fmt.Println("no matching transactions found")
		return nil
	}

	n = min(n, len(filtered))

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Merchant Name", "Time", "Amount", "Category", "Balance"})
	for _, v := range filtered[:n] {
		if v.Category == "mondo" {
			v.Merchant.Name = "Mondo"
		}
		t.Append([]string{v.Merchant.Name, v.Created.Format(time.RFC822), fmt.Sprintf("£%.2f", float64(v.Amount)/100), v.Category, fmt.Sprintf("£%.2f", float64(v.AccountBalance)/100)})
	}
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.Render()
	return nil
}

func filterTransactions(cand []*mondo.Transaction, f func(t *mondo.Transaction) bool) []*mondo.Transaction {
	filt := make([]*mondo.Transaction, 0, len(cand))
	for _, t := range cand {
		if ok := f(t); ok {
			filt = append(filt, t)
		}
	}
	return filt
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
