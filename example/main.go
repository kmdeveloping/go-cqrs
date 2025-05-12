//go:generate go run ../tools/gen-handler-registry/main.go

package main

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/queries"
)

func init() {
	m := cqrs.NewCqrsManager()
	m.AddMetricsDecorator()
	m.AddLoggingDecorator()

	registerHandlers()
}

func main() {
	doSomethingCommand := commands.DoSomethingCommand{
		Something: "Helloooooo",
	}
	err := cqrs.ExecuteCommand(doSomethingCommand)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Now the command result should be set by the handler since we're using a pointer interface
	//fmt.Println(doSomethingCommand.GetResult())

	result, er := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](queries.GetNameQuery{ID: 987})
	if er != nil {
		log.Fatal(er)
		return
	}

	log.Println(result.UserName)
}
