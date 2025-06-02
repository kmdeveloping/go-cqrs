package handlers

import (
	"context"
	"fmt"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/cqrs"
	"github.com/kmdeveloping/go-cqrs/example/commands"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

// CreateUserCommandHandler demonstrates dependency injection with auto-registration
type CreateUserCommandHandler struct {
	// Dependencies auto-injected using the 'inject' tag
	UserRepo            UserRepository      `inject:""`
	NotificationService NotificationService `inject:""`
	Logger              Logger              `inject:""`
}

// Ensure handler implements the interface
var _ command.ICommandHandler[commands.CreateUserCommand] = (*CreateUserCommandHandler)(nil)

func (h *CreateUserCommandHandler) Handle(ctx context.Context, cmd *commands.CreateUserCommand) error {
	h.Logger.Infof("Processing CreateUserCommand for: %s (%s)", cmd.Name, cmd.Email)

	// Check if user already exists
	existingUser, err := h.UserRepo.GetByEmail(cmd.Email)
	if err == nil && existingUser != nil {
		h.Logger.Error("User with email already exists")
		return fmt.Errorf("user with email %s already exists", cmd.Email)
	}

	// Create new user
	user := User{
		Name:  cmd.Name,
		Email: cmd.Email,
	}

	// Save user using injected repository
	if err := h.UserRepo.Save(user); err != nil {
		h.Logger.Error("Failed to save user")
		return fmt.Errorf("failed to save user: %w", err)
	}

	h.Logger.Info("User created successfully")

	// Send welcome email using injected notification service
	if err := h.NotificationService.SendWelcomeEmail(user.Email, user.Name); err != nil {
		h.Logger.Error("Failed to send welcome email")
		// Don't fail the command for notification errors, just log
	}

	// Publish event using clean API
	return cqrs.PublishEvent(ctx, events.UserCreatedEvent{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
	})
}
