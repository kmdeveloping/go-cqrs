//go:generate go run ../tools/gen-handler-registry/main.go

package main

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/queries"
)

func init() {
	BootstrapCqrs()
}

func main() {
	err := cqrs.ExecuteCommand(commands.DoSomethingCommand{Something: "Helloooooo"})
	if err != nil {
		log.Fatal(err)
		return
	}

	result, er := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](queries.GetNameQuery{ID: 987})
	if er != nil {
		log.Fatal(er)
		return
	}

	log.Println(result.UserName)
}
