package main

import (
	"github.com/kmdeveloping/go-cqrs/core/cqrs"
	"github.com/kmdeveloping/go-cqrs/core/registry"
	"github.com/kmdeveloping/go-cqrs/example/contracts"
	"github.com/kmdeveloping/go-cqrs/example/handler"
)

var dispatcher cqrs.ICqrsManager

func init() {
	config := &cqrs.CqrsConfiguration{
		Handlers: registry.NewRegistry(),
	}

	registry.RegisterCommand[contracts.DoSomethingCommand](&handler.DoThatCommandHandler{}, &config.Handlers)

	dispatcher = cqrs.NewCqrsManager(config)
	dispatcher.UseLoggingDecorator()
	dispatcher.UseMetricsDecorator()
	dispatcher.UseErrorHandlerDecorator()
}

func main() {
	h := dispatcher.Execute(contracts.DoSomethingCommand{Something: "Hi"})
	if h != nil {
		return
	}
}
