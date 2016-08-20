package main

import (
	"fmt"
	"os"

	mondo "github.com/sjwhitworth/gomondo"
)

func init() { RegisterCommand("login", auth{}) }

type auth struct{}

func (h auth) Help() string { return "login to mondo" }
func (h auth) Autocomplete(line string, pos int) (string, []string, string) {
	return line, nil, ""
}

func (h auth) F(cmd string, quitCh chan bool) error {
	if client != nil {
		return nil
	}
	clientId := os.Getenv("MONDO_CLIENT_ID")
	clientSecret := os.Getenv("MONDO_CLIENT_SECRET")

	if clientId == "" {
		return fmt.Errorf("could not read $MONDO_CLIENT_ID from environment")
	}

	if clientSecret == "" {
		return fmt.Errorf("could not read $MONDO_CLIENT_SECRET from environment")
	}

	username, err := l.PasswordPrompt("please enter the email you use for mondo: ")
	if err != nil {
		return err
	}

	password, err := l.PasswordPrompt("please enter your mondo password: ")
	if err != nil {
		return err
	}

	fmt.Printf("thanks! logging in...\n")

	c, err := mondo.Authenticate(clientId, clientSecret, username, password)
	if err != nil {
		return err
	}

	a, err := c.Accounts()
	if err != nil {
		return err
	}

	accountId = a[0].ID
	client = c

	return nil
}
