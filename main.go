package main

import (
	"fmt"
	"github.com/phayes/hookserve/hookserve"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var app *cli.App
var listenValue int
var secretValue string
var rootValue string
var gitBaseValue string
var gogsBaseValue string
var gogsUrlValue string
var gogsTokenValue string
var gogsOwnerValue string
var runFlags = []cli.Flag{
	cli.IntFlag{
		EnvVar:      "UPDATE_TULEAP_LISTEN",
		Value:       8888,
		Destination: &listenValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_SECRET",
		Destination: &secretValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_ROOT",
		Value:       "/var/lib/tuleap/gitolite/repositories",
		Destination: &rootValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_GIT_BASE",
		Value:       "git@github.com:udistrital",
		Destination: &gitBaseValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_GOGS_BASE",
		Value:       "ssh://git@gogs.udistritaloas.edu.co:10022",
		Destination: &gogsBaseValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_GOGS_URL",
		Value:       "https://gogs.udistritaloas.edu.co/",
		Destination: &gogsUrlValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_GOGS_TOKEN",
		Destination: &gogsTokenValue,
	},
	cli.StringFlag{
		EnvVar:      "UPDATE_TULEAP_GOGS_OWNER",
		Value:       "gogsadmin",
		Destination: &gogsOwnerValue,
	},
}

func init() {
	app = cli.NewApp()
	app.Usage = "Updates tuleap & GOGS"
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Action: runAction,
			Flags:  runFlags,
		},
	}
}

func runGitCommand(git_dir string, args []string) (output string, err error) {
	log.Printf("git args: %v\n", args)
	command := exec.Command("git", args...)
	command.Env = []string{"GIT_DIR=" + git_dir}
	output, err = combineOutput(command)
	return
}

func getMatch(eventRepo string, match, project, repo *string) (found int) {
	var matches []string
	var err error
	if matches, err = filepath.Glob(fmt.Sprintf("%s/*/%s.git", rootValue, eventRepo)); err != nil {
		log.Print(err)
		return 0
	}
	len_matches := len(matches)
	log.Printf("len matches for %v: %v\n", eventRepo, len_matches)
	if len_matches != 1 {
		return len_matches
	}
	*match = matches[0]
	log.Printf("match: %v\n", *match)
	split_match := strings.Split(*match, "/")
	*repo, split_match = split_match[len(split_match)-1], split_match[:len(split_match)-1]
	*repo = strings.TrimSuffix(*repo, ".git")
	log.Printf("repo: %v\n", *repo)
	*project, split_match = split_match[len(split_match)-1], split_match[:len(split_match)-1]
	log.Printf("project: %v\n", *project)
	return 1
}

func combineOutput(command *exec.Cmd) (output string, err error) {
	var combined_output []byte
	if combined_output, err = command.CombinedOutput(); err != nil {
		log.Print(err)
	}
	output = string(combined_output[:])
	output = strings.TrimSpace(output)
	if output != "" {
		log.Print(output)
	}
	return
}

func ensureGogsOrgRepo(org_name, repo_name string) (output string, err error) {
	command := exec.Command("/usr/local/lib/tuleap-gogs-hook/ensure_org_repo")
	command.Env = []string{
		"GOGS_URL=" + gogsUrlValue,
		"ORG_OWNER=" + gogsOwnerValue,
		"GOGS_TOKEN=" + gogsTokenValue,
		"ORG_NAME=" + org_name,
		"REPO_NAME=" + repo_name,
	}
	output, err = combineOutput(command)
	return
}

func processPushEvent(event hookserve.Event) (err error) {
	var match, project, repo string
	numMatches := getMatch(event.Repo, &match, &project, &repo)
	if numMatches != 1 {
		// el repo no se encuentra en tuleap o se encuentra mas de una vez
		// subir a drone con el nombre de organización que tiene en github
		// bajar en tuleap en una ubicación temporal
		match, err = ioutil.TempDir("", "update_tuleap_"+event.Repo)
		if err != nil {
			log.Print(err)
			return
		}
		defer os.RemoveAll(match)
		project = event.Owner
		repo = event.Repo
		runGitCommand(match, []string{"init"})
	}
	runGitCommand(match, []string{
		"fetch",
		fmt.Sprintf("%s/%s%s", gitBaseValue, event.Repo, ".git"),
		fmt.Sprintf("%s:%s", event.Branch, event.Branch),
	})
	gogsRepoRef := fmt.Sprintf("%s/%s/%s%s", gogsBaseValue, project, event.Repo, ".git")
	ensureGogsOrgRepo(project, event.Repo)
	git_output, _ := runGitCommand(match, []string{
		"ls-remote",
		gogsRepoRef,
		event.Branch,
	})
	if git_output == "" {
		// si el branch no existe en gogs crearlo
		runGitCommand(match, []string{
			"push",
			gogsRepoRef,
			fmt.Sprintf("%s:%s", event.Branch, event.Branch),
		})
	}
	if event.Branch != "master" {
		// si el branch es master siempre hacer push a develop
		runGitCommand(match, []string{
			"push",
			gogsRepoRef,
			fmt.Sprintf("%s:%s", "master", "develop"),
		})
	}
	return
}

func processEvents(events chan hookserve.Event) {
	select {
	case event := <-events:
		if event.Type == "push" {
			processPushEvent(event)
		}
	}
}

func runAction(ctx *cli.Context) (err error) {
	server := hookserve.NewServer()
	server.Port = listenValue
	server.Secret = secretValue
	server.GoListenAndServe()
	for {
		processEvents(server.Events)
	}
	return
}

func main() {
	app.Run(os.Args)
}
