package commands

import "github.com/kmdeveloping/go-cqrs/command"

// CreateUserCommand demonstrates a command for the dependency injection example
type CreateUserCommand struct {
	command.Base
	Name  string
	Email string
}

var _ command.ICommand = (*CreateUserCommand)(nil)
