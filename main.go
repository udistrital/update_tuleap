package main

import (
	"fmt"
	"github.com/phayes/hookserve/hookserve"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var app *cli.App
var listenValue int
var secretValue string
var rootValue string
var gitbaseValue string
var gitsuffixValue string
var dronebaseValue string

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
				cli.StringFlag{
					Name:        "dronebase",
					EnvVar:      "UPDATE_TULEAP_DRONE_BASE",
					Value:       "ssh://git@ci.udistritaloas.edu.co:10022",
					Destination: &dronebaseValue,
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
				fmt.Printf("len matches for %v: %v\n", event.Repo, len(matches))
				if len(matches) != 1 {
					continue
				}
				match := matches[0]
				fmt.Printf("match: %v\n", match)
				cmnd := "git"
				args := []string{
					"fetch",
					fmt.Sprintf("%s/%s%s", gitbaseValue, event.Repo, gitsuffixValue),
					fmt.Sprintf("%s:%s", event.Branch, event.Branch),
				}
				fmt.Printf("cmnd & args: %v %v\n", cmnd, args)
				command := exec.Command(cmnd, args...)
				command.Env = []string{"GIT_DIR=" + match}
				if combined_output, err = command.CombinedOutput(); err != nil {
					fmt.Println(err.Error())
				}
				if len(combined_output) != 0 {
					fmt.Print(string(combined_output[:]))
				}
				if event.Branch != "master" {
					continue
				}
				split_match := strings.Split(match, "/")
				split_match_size := len(split_match)
				if split_match_size < 2 {
					continue
				}
				tuleap_project_name := split_match[split_match_size-2]
				args = []string{
					"push",
					fmt.Sprintf("%s/%s/%s%s", dronebaseValue, tuleap_project_name, event.Repo, gitsuffixValue),
					"master:develop",
				}
				fmt.Printf("cmnd & args: %v %v\n", "echo", args) // replace "echo" with cmnd after testing
				command = exec.Command("echo", args...)          // replace "echo" with cmnd after testing
				command.Env = []string{"GIT_DIR=" + match}
				if combined_output, err = command.CombinedOutput(); err != nil {
					fmt.Println(err.Error())
				}
				if len(combined_output) != 0 {
					fmt.Print(string(combined_output[:]))
				}
			}
		}
	}

	return
}

func main() {
	app.Run(os.Args)
}
