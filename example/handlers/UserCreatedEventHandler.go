package handlers

import (
	"context"
	"strconv"

	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/example/events"
)

// UserCreatedEventHandler demonstrates dependency injection with auto-registration
type UserCreatedEventHandler struct {
	// Dependencies auto-injected using the 'inject' tag
	NotificationService NotificationService `inject:""`
	Logger              Logger              `inject:""`
}

// Ensure handler implements the interface
var _ event.IEventHandler[events.UserCreatedEvent] = (*UserCreatedEventHandler)(nil)

func (h *UserCreatedEventHandler) Handle(ctx context.Context, e events.UserCreatedEvent) error {
	h.Logger.Infof("Processing UserCreatedEvent for user: %s (%s)", e.Name, e.Email)

	// Send additional notifications using injected service
	userIDStr := strconv.Itoa(e.UserID)
	message := "Welcome to our platform! Your account has been created successfully."

	if err := h.NotificationService.SendNotification(userIDStr, message); err != nil {
		h.Logger.Error("Failed to send notification")
		// Don't fail for notification errors, just log
	}

	h.Logger.Info("UserCreatedEvent processed successfully")
	return nil
}
