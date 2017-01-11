package main

import (
	"fmt"
	"github.com/phayes/hookserve/hookserve"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
)

var app *cli.App
var listenValue int
var secretValue string
var rootValue string
var gitbaseValue string
var gitsuffixValue string

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
				cli.StringFlag{
					Name:        "gitbase",
					EnvVar:      "UPDATE_TULEAP_GIT_BASE",
					Value:       "git@github.com:udistrital",
					Destination: &gitbaseValue,
				},
				cli.StringFlag{
					Name:        "gitsuffix",
					EnvVar:      "UPDATE_TULEAP_GIT_SUFFIX",
					Value:       ".git",
					Destination: &gitsuffixValue,
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

	for {
		select {
		case event := <-server.Events:
			if event.Type == "push" {
				var matches []string
				var combined_output []byte
				var err error
				if matches, err = filepath.Glob(fmt.Sprintf("%s/*/%s.git", rootValue, event.Repo)); err != nil {
					fmt.Println(err.Error())
					continue
				}
				fmt.Printf("len matches for %s: %d\n", event.Repo, len(matches))
				if len(matches) != 1 {
					continue
				}
				match := matches[0]
				fmt.Printf("match: %s\n", match)
				command := exec.Command(fmt.Sprintf("echo git fetch %s/%s%s %s:%s", gitbaseValue, match, gitsuffixValue, event.Branch, event.Branch))
				if combined_output, err = command.CombinedOutput(); err != nil {
					fmt.Println(err.Error())
				} else if len(combined_output) != 0 {
					fmt.Println(combined_output[:])
				}
			}
		}
	}

	return
}

func main() {
	app.Run(os.Args)
}
