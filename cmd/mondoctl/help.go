package main

import "fmt"

func init() { RegisterCommand("help", helper{}) }

type helper struct{}

func (h helper) Help() string { return "lists the help" }
func (h helper) Autocomplete(line string, pos int) (string, []string, string) {
	return line, nil, ""
}
func (h helper) F(cmd string, quitCh chan bool) error {
	for k, v := range cmds {
		fmt.Printf("%v\t%v\n\n", k, v.Help())
	}
	fmt.Printf("\n")
	return nil
}
