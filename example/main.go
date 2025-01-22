package main

import (
	"fmt"
	"github.com/kmdeveloping/go-cqrs/example/contracts/commands"
	"github.com/kmdeveloping/go-cqrs/example/contracts/queries"
	"github.com/kmdeveloping/go-cqrs/example/handlers/commandHandlers"
	"github.com/kmdeveloping/go-cqrs/example/handlers/queryHandlers"
)

func main() {

	cmd := &commands.DoSomethingCommand{CustomerNumber: "4563490848"}

	err := commandHandlers.NewDoSomethingCommand().Execute(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}

	query := &queries.GetSomethingQuery{
		CustomerNumber: "098502985838",
	}

	result, err := queryHandlers.NewGetSomethingQuery().Execute(query)
	if err != nil {
		fmt.Println(err.Error())
	}
	if result != nil {
		for _, r := range *result {
			fmt.Println(r)
		}
	}
}
