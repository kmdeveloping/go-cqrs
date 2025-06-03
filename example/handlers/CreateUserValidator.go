package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/validator"
)

// CreateUserValidator demonstrates dependency injection with auto-registration
type CreateUserValidator struct {
	// Dependencies auto-injected using the 'inject' tag
	UserRepo UserRepository `inject:""`
	Logger   Logger         `inject:""`
}

// Ensure validator implements the interface
var _ validator.IValidatorHandler[commands.CreateUserCommand] = (*CreateUserValidator)(nil)

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *commands.CreateUserCommand) error {
	v.Logger.Infof("Validating CreateUserCommand for: %s (%s)", cmd.Name, cmd.Email)

	// Basic validation
	if strings.TrimSpace(cmd.Name) == "" {
		v.Logger.Error("Name is required")
		return fmt.Errorf("name is required")
	}

	if strings.TrimSpace(cmd.Email) == "" {
		v.Logger.Error("Email is required")
		return fmt.Errorf("email is required")
	}

	// Simple email format validation
	if !strings.Contains(cmd.Email, "@") || !strings.Contains(cmd.Email, ".") {
		v.Logger.Error("Invalid email format")
		return fmt.Errorf("invalid email format")
	}

	// Check if user already exists using injected repository
	existingUser, err := v.UserRepo.GetByEmail(cmd.Email)
	if err == nil && existingUser != nil {
		v.Logger.Error("User with email already exists")
		return fmt.Errorf("user with email %s already exists", cmd.Email)
	}

	v.Logger.Info("CreateUserCommand validation passed")
	return nil
}
