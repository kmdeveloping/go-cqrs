package events

import "github.com/kmdeveloping/go-cqrs/event"

// UserCreatedEvent demonstrates an event for the dependency injection example
type UserCreatedEvent struct {
	event.Base
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

var _ event.IEvent = (*UserCreatedEvent)(nil)
