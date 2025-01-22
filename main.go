package main

import (
	"fmt"

	"github.com/kmdeveloping/go-cqrs/contracts/commands"
	"github.com/kmdeveloping/go-cqrs/contracts/queries"
)

func main() {

	cmd := &commands.GetSomethingCommand{
		CustomerNumber: "0983409283",
	}

	err := commands.NewGetSomethingCommand().Execute(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}

	query := &queries.GetSomethingQuery{
		CustomerNumber: "098502985838",
	}

	result, err := queries.NewGetSomethingQuery().Execute(query)
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, r := range *result {
		fmt.Println(r)
	}
}
