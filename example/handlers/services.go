package handlers

import (
	"fmt"
	"log"
)

// Service interfaces for dependency injection
type Logger interface {
	Info(message string)
	Error(message string)
	Infof(format string, args ...interface{})
}

type NotificationService interface {
	SendWelcomeEmail(email, name string) error
	SendNotification(userID string, message string) error
}

type UserRepository interface {
	Save(user User) error
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	List() ([]*User, error)
}

// Domain model for dependency injection example
type User struct {
	ID    int
	Name  string
	Email string
}

// Mock implementations

// ConsoleLogger - simple console logger implementation
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(message string) {
	log.Printf("[INFO] %s", message)
}

func (l *ConsoleLogger) Error(message string) {
	log.Printf("[ERROR] %s", message)
}

func (l *ConsoleLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// MockNotificationService - mock notification service
type MockNotificationService struct{}

func (s *MockNotificationService) SendWelcomeEmail(email, name string) error {
	log.Printf("📧 Sending welcome email to %s (%s)", name, email)
	return nil
}

func (s *MockNotificationService) SendNotification(userID, message string) error {
	log.Printf("🔔 Sending notification to user %s: %s", userID, message)
	return nil
}

// InMemoryUserRepository - simple in-memory user repository
type InMemoryUserRepository struct {
	users  map[int]*User
	nextID int
}

func (r *InMemoryUserRepository) Save(user User) error {
	if r.users == nil {
		r.users = make(map[int]*User)
		r.nextID = 1
	}

	if user.ID == 0 {
		user.ID = r.nextID
		r.nextID++
	}

	r.users[user.ID] = &user
	log.Printf("💾 Saved user: %+v", user)
	return nil
}

func (r *InMemoryUserRepository) GetByID(id int) (*User, error) {
	if r.users == nil {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}

	return user, nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*User, error) {
	if r.users == nil {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user with email %s not found", email)
}

func (r *InMemoryUserRepository) List() ([]*User, error) {
	if r.users == nil {
		return []*User{}, nil
	}

	var users []*User
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}
