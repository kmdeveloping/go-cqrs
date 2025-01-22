package main

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/example/contracts/commands"
	"github.com/kmdeveloping/go-cqrs/example/contracts/queries"
	"github.com/kmdeveloping/go-cqrs/example/handlers/commandHandlers"
	"github.com/kmdeveloping/go-cqrs/example/handlers/queryHandlers"
)

func main() {
	if err := commandHandlers.NewDoSomethingCommand().Execute(&commands.DoSomethingCommand{CustomerNumber: "4563490848"}); err != nil {
		fmt.Println(err.Error())
	}

	if result, _ := queryHandlers.NewGetSomethingQuery().Execute(&queries.GetSomethingQuery{CustomerNumber: "098502985838"}); result != nil {
		for _, r := range *result {
			fmt.Println(r)
		}
	}
}
