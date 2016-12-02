package main

import (
	"fmt"
	"github.com/phayes/hookserve/hookserve"
	"os"
	"strconv"
)

// correr este programa con gitolite en el servidor de tuleap

func main() {

	server := hookserve.NewServer()
	server.Port, _ = strconv.Atoi(os.Getenv("UPDATE_TULEAP_LISTEN"))
	server.Secret = os.Getenv("UPDATE_TULEAP_SECRET")
	server.GoListenAndServe()

	// Everytime the server receives a webhook event, print the results
	for event := range server.Events {
		if event.Type == "push" {
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
