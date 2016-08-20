package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"

	"log"

	"github.com/peterh/liner"
	mondo "github.com/sjwhitworth/gomondo"
)

type Command interface {
	Help() string
	F(string, chan bool) error
	Autocomplete(string, int) (string, []string, string)
}

var (
	l       *liner.State
	cmds    = make(map[string]Command)
	history = "/tmp/mondoctl.history"

	client    *mondo.Client
	accountId string

	debug = false
)

func init() {
	log.SetOutput(ioutil.Discard)

	l = liner.NewLiner()
	l.SetCtrlCAborts(true)

	// Set an autocompleter per function.
	l.SetWordCompleter(func(line string, pos int) (string, []string, string) {
		fields := strings.Fields(line)

		if len(fields) > 0 {
			for k, v := range cmds {
				if strings.Replace(strings.ToLower(fields[0]), " ", "", -1) == strings.ToLower(k) {
					return v.Autocomplete(line, pos)
				}
			}
		}

		// Hmm is this the right thing to do?
		c := []string{}
		for k := range cmds {
			if strings.HasPrefix(k, line) {
				c = append(c, k)
			}
		}
		return "", c, ""
	})
}

func RegisterCommand(key string, f Command) {
	if _, exists := cmds[key]; exists {
		panic("clashing commands")
	}

	cmds[key] = f
}

var invalidCmd = fmt.Errorf("not a valid command")

func executeCmd(cmd string, quitCh chan bool) error {
	t := time.Now()

	cmd = strings.TrimSpace(cmd)

	fields := strings.Fields(cmd)
	firstarg := fields[0]

	// remove extra white space between fields
	trimcmd := ""
	for _, field := range fields {
		if len(field) > 0 {
			trimcmd += field + " "
		}
	}

	// Get rid of trailing space
	trimcmd = strings.TrimSpace(trimcmd)

	if _, in := cmds[firstarg]; !in {
		return invalidCmd
	}

	command := cmds[firstarg]
	err := command.F(trimcmd, quitCh)

	if debug {
		fmt.Printf("\033[94mCommand took %v ms\033[0m\n", time.Since(t).Nanoseconds()/1000000)
	}

	return err
}

func main() {
	defer l.Close()

	quitCh := make(chan bool)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			<-c
			quitCh <- true
		}
	}()

	interactive(quitCh)
}

func interactive(quitCh chan bool) {
	f, err := os.Open(history)
	defer f.Close()
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("ERROR: Error opening history: %v", err)
	}

	if err != nil && os.IsNotExist(err) {
		os.Create(history)
	}

	if err == nil {
		if _, err := l.ReadHistory(f); err != nil {
			fmt.Printf("ERROR: Error reading history: %v", err)
		}
	}

	defer func() {
		if f, err := os.Create(history); err != nil {
			fmt.Printf("Error writing history file: %v", err)
		} else {
			l.WriteHistory(f)
			f.Close()
		}
	}()

	fmt.Println(`
███╗   ███╗ ██████╗ ███╗   ██╗██████╗  ██████╗
████╗ ████║██╔═══██╗████╗  ██║██╔══██╗██╔═══██╗
██╔████╔██║██║   ██║██╔██╗ ██║██║  ██║██║   ██║
██║╚██╔╝██║██║   ██║██║╚██╗██║██║  ██║██║   ██║
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║██████╔╝╚██████╔╝
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═════╝  ╚═════╝
`)

	if client == nil {
		if err := executeCmd("login", quitCh); err != nil {
			fmt.Printf("error logging in: %v\n", err)
			os.Exit(1)
		}
	}

	for {
		fmt.Printf("\033[1m\033[94m")
		cmd, err := l.Prompt("mondoctl: ")
		fmt.Printf("\033[0m")

		if err != nil && err.Error() == "EOF" {
			fmt.Println("bye!")
			return
		} else if err == liner.ErrPromptAborted {
			return
		} else if err != nil {
			fmt.Println("bye!")
			panic(err)
		}

		if cmd == "" {
			continue
		}

		l.AppendHistory(cmd)

		err = executeCmd(cmd, quitCh)
		if err != nil {
			fmt.Printf("\033[1m\033[91mERROR: %v\033[0m\n", err)
		}
	}
}

func loadhistory(fpath string) {
	if f, err := os.Open(fpath); err == nil {
		l.ReadHistory(f)
		f.Close()
	}
}
