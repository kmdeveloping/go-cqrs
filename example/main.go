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
	config := &cqrs.CqrsConfiguration{
		Registry: registry.NewRegistry().
			RegisterCommandHandlers(getCommandServices()).
			RegisterQueryHandlers(getQueryServices()),
	}

	dispatcher = cqrs.NewCqrsManager(config)
	dispatcher.UseLoggingDecorator()
	dispatcher.UseMetricsDecorator()
	dispatcher.UseErrorHandlerDecorator()
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

func getCommandServices() []registry.CommandServices {
	return []registry.CommandServices{
		{Command: contracts.DoSomethingCommand{}, Handler: &handler.DoThatCommandHandler{}},
	}
}

func getQueryServices() []registry.QueryServices {
	return []registry.QueryServices{
		{Query: contracts.GetNewUserQuery{}, Handler: &handler.GetNewUserQueryHandler{}},
	}
}
