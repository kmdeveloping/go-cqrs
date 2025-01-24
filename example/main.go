package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/core/registry"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
)

var dispatcher cqrs.ICqrsManager

func init() {

	services := []registry.Service{
		{
			TContract: contracts.DoSomethingCommand{},
			THandler:  &handler.DoThatCommandHandler{},
		},
	}

	config := &cqrs.CqrsConfiguration{
		Registry: registry.NewRegistry().RegisterHandlers(services),
	}

	dispatcher = cqrs.NewCqrsManager(config).UseLoggingDecorator()
}

func main() {
	h := dispatcher.Execute(contracts.DoSomethingCommand{Something: "Hello"})
	if h != nil {
		return
	}
}
