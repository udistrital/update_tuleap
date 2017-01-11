package main

import (
	"fmt"
	"github.com/phayes/hookserve/hookserve"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

var app *cli.App
var listenValue int
var secretValue string
var rootValue string

func init() {
	app = cli.NewApp()
	app.Usage = "Updates tuleap"
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Action: runaction,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:        "listen, l",
					EnvVar:      "UPDATE_TULEAP_LISTEN",
					Value:       8888,
					Destination: &listenValue,
				},
				cli.StringFlag{
					Name:        "secret, s",
					EnvVar:      "UPDATE_TULEAP_SECRET",
					Destination: &secretValue,
				},
				cli.StringFlag{
					Name:        "root, r",
					EnvVar:      "UPDATE_TULEAP_ROOT",
					Value:       "/var/lib/tuleap/gitolite/repositories",
					Destination: &rootValue,
				},
			},
		},
	}
}

func runaction(ctx *cli.Context) (err error) {
	server := hookserve.NewServer()
	server.Port = listenValue
	server.Secret = secretValue
	server.GoListenAndServe()

	// Everytime the server receives a webhook event, print the results
	for {
		select {
		case event := <-server.Events:
			if event.Type == "push" {
				var matches []string
				var err error
				if matches, err = filepath.Glob(rootValue + "/*/" + event.Repo + ".git"); err != nil {
					fmt.Println(err.Error())
					continue
				}
				for match := range matches {
					fmt.Println(match)
				}
				// buscar en los directorios de tuleap event.Repo
				// si no lo encontramos:
				// fail
				// si encontramos mas de uno
				// fail
				// cd "the repo dir" && git fetch "the repo url" +refs/heads/*:refs/heads/* --prune
				fmt.Println("%v", event)
			}
		}
	}

	return
}

func main() {
	app.Run(os.Args)
}
