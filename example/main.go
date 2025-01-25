package main

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/core/registry"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
)

var dispatcher cqrs.ICqrsManager

func init() {
	commandHandlers := []registry.CommandServices{
		{
			Command: contracts.DoSomethingCommand{},
			Handler: &handler.DoThatCommandHandler{},
		},
	}

	queryHandlers := []registry.QueryServices{
		{
			Query:   contracts.GetNewUserQuery{},
			Handler: &handler.GetNewUserQueryHandler{},
		},
	}

	config := &cqrs.CqrsConfiguration{
		Registry: registry.NewRegistry().
			RegisterCommandHandlers(commandHandlers).
			RegisterQueryHandlers(queryHandlers),
	}

	dispatcher = cqrs.NewCqrsManager(config)
}

func main() {
	h := dispatcher.Execute(contracts.DoSomethingCommand{Something: "Hello"})
	if h != nil {
		return
	}

	r, e := dispatcher.Get(contracts.GetNewUserQuery{
		FirstName: "Kolten",
		LastName:  "Mollencopf",
		Project:   "P1",
	})
	if e != nil {
		return
	}

	fmt.Println(r)
}
