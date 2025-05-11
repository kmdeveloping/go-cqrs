//go:generate go run ../tools/gen-handler-registry/main.go

package main

import (
	"fmt"
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

	// this result is supposed to be set by the command handler
	// but it is not set since the ExecuteCommand method does not use a pointer
	// this is a bug in the library and the command handler should be able to set the result parameter
	fmt.Println(doSomethingCommand.Result)

	result, er := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](queries.GetNameQuery{ID: 987})
	if er != nil {
		log.Fatal(er)
		return
	}

	log.Println(result.UserName)
}
