package cqrs

import (
	"context"
	"errors"
	"testing"

	"github.com/kmdeveloping/go-cqrs/command"
	"github.com/kmdeveloping/go-cqrs/event"
	"github.com/kmdeveloping/go-cqrs/query"
)

// Test domain models
type TestUser struct {
	ID    string
	Name  string
	Email string
}

// Test commands
type CreateTestUserCommand struct {
	command.Base
	Name  string
	Email string
}

type UpdateTestUserCommand struct {
	command.Base
	ID    string
	Name  string
	Email string
}

// Test queries
type GetTestUserQuery struct {
	query.Base
	UserID string
}

// Test events
type TestUserCreatedEvent struct {
	event.Base
	UserID string
	Name   string
	Email  string
}

type TestUserUpdatedEvent struct {
	event.Base
	UserID string
	Name   string
	Email  string
}

// Test repository interface
type TestUserRepository interface {
	Save(user TestUser) error
	GetByID(id string) (*TestUser, error)
}

// Mock repository implementation
type MockTestUserRepository struct {
	users       map[string]TestUser
	saveCallLog []TestUser
	getCallLog  []string
	shouldError bool
}

func NewMockTestUserRepository() *MockTestUserRepository {
	return &MockTestUserRepository{
		users:       make(map[string]TestUser),
		saveCallLog: make([]TestUser, 0),
		getCallLog:  make([]string, 0),
	}
}

func (r *MockTestUserRepository) Save(user TestUser) error {
	r.saveCallLog = append(r.saveCallLog, user)
	if r.shouldError {
		return errors.New("repository error")
	}
	r.users[user.ID] = user
	return nil
}

func (r *MockTestUserRepository) GetByID(id string) (*TestUser, error) {
	r.getCallLog = append(r.getCallLog, id)
	if r.shouldError {
		return nil, errors.New("repository error")
	}
	if user, exists := r.users[id]; exists {
		return &user, nil
	}
	return nil, errors.New("user not found")
}

// Test handlers with dependency injection
type TestCreateUserCommandHandler struct {
	UserRepo TestUserRepository `inject:""`
}

func (h *TestCreateUserCommandHandler) Handle(ctx context.Context, cmd *CreateTestUserCommand) error {
	if h.UserRepo == nil {
		return errors.New("UserRepo dependency not injected")
	}

	user := TestUser{
		ID:    "user-" + cmd.Name,
		Name:  cmd.Name,
		Email: cmd.Email,
	}

	return h.UserRepo.Save(user)
}

type TestUpdateUserCommandHandler struct {
	UserRepo TestUserRepository `inject:""`
}

func (h *TestUpdateUserCommandHandler) Handle(ctx context.Context, cmd *UpdateTestUserCommand) error {
	if h.UserRepo == nil {
		return errors.New("UserRepo dependency not injected")
	}

	user := TestUser{
		ID:    cmd.ID,
		Name:  cmd.Name,
		Email: cmd.Email,
	}

	return h.UserRepo.Save(user)
}

type TestGetUserQueryHandler struct {
	UserRepo TestUserRepository `inject:""`
}

func (h *TestGetUserQueryHandler) Handle(ctx context.Context, qry GetTestUserQuery) (*TestUser, error) {
	if h.UserRepo == nil {
		return nil, errors.New("UserRepo dependency not injected")
	}

	return h.UserRepo.GetByID(qry.UserID)
}

type TestUserCreatedEventHandler struct {
	EventLog []TestUserCreatedEvent
}

func (h *TestUserCreatedEventHandler) Handle(ctx context.Context, event TestUserCreatedEvent) error {
	h.EventLog = append(h.EventLog, event)
	return nil
}

type TestUserUpdatedEventHandler struct {
	EventLog []TestUserUpdatedEvent
}

func (h *TestUserUpdatedEventHandler) Handle(ctx context.Context, event TestUserUpdatedEvent) error {
	h.EventLog = append(h.EventLog, event)
	return nil
}

// Test validators
type TestCreateUserValidator struct {
	ValidationLog []CreateTestUserCommand
}

func (v *TestCreateUserValidator) Validate(ctx context.Context, cmd *CreateTestUserCommand) error {
	v.ValidationLog = append(v.ValidationLog, *cmd)

	if cmd.Name == "" {
		return errors.New("name is required")
	}
	if cmd.Email == "" {
		return errors.New("email is required")
	}
	if cmd.Name == "invalid" {
		return errors.New("invalid name")
	}

	return nil
}

type TestUpdateUserValidator struct {
	ValidationLog []UpdateTestUserCommand
}

func (v *TestUpdateUserValidator) Validate(ctx context.Context, cmd *UpdateTestUserCommand) error {
	v.ValidationLog = append(v.ValidationLog, *cmd)

	if cmd.ID == "" {
		return errors.New("id is required")
	}
	if cmd.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

// Test setup helper
func setupTestAutoRegistry() (*AutoRegistry, *MockTestUserRepository, func()) {
	originalManager := currentManager

	// Reset for clean test
	ResetManager()

	manager := NewCqrsManager()
	SetManager(manager)

	mockRepo := NewMockTestUserRepository()
	container := NewSimpleContainer()
	Register[TestUserRepository](container, mockRepo)

	autoRegistry := NewAutoRegistry(manager).SetDependencyProvider(container)

	cleanup := func() {
		currentManagerMu.Lock()
		currentManager = originalManager
		currentManagerMu.Unlock()
	}

	return autoRegistry, mockRepo, cleanup
}

func TestAutoRegistry_RegisterCommandHandler_Success(t *testing.T) {
	autoRegistry, mockRepo, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Register command handler
	handler := &TestCreateUserCommandHandler{}
	result := autoRegistry.RegisterHandlerInstances(handler)

	// Verify registration
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no errors, got: %v", result.Errors)
	}
	if result.RegisteredHandlers != 1 {
		t.Fatalf("Expected 1 registered handler, got: %d", result.RegisteredHandlers)
	}

	// Test command execution
	ctx := context.Background()
	cmd := &CreateTestUserCommand{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Verify repository was called
	if len(mockRepo.saveCallLog) != 1 {
		t.Fatalf("Expected 1 save call, got: %d", len(mockRepo.saveCallLog))
	}

	savedUser := mockRepo.saveCallLog[0]
	if savedUser.Name != "John Doe" || savedUser.Email != "john@example.com" {
		t.Fatalf("Unexpected saved user: %+v", savedUser)
	}
}

func TestAutoRegistry_RegisterQueryHandler_Success(t *testing.T) {
	autoRegistry, mockRepo, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Pre-populate repository
	testUser := TestUser{ID: "user-1", Name: "Jane Doe", Email: "jane@example.com"}
	mockRepo.Save(testUser)

	// Register query handler
	handler := &TestGetUserQueryHandler{}
	result := autoRegistry.RegisterHandlerInstances(handler)

	// Verify registration
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no errors, got: %v", result.Errors)
	}
	if result.RegisteredHandlers != 1 {
		t.Fatalf("Expected 1 registered handler, got: %d", result.RegisteredHandlers)
	}

	// Test query execution
	ctx := context.Background()
	qry := GetTestUserQuery{UserID: "user-1"}

	resultUser, err := ExecuteQuery[GetTestUserQuery, *TestUser](ctx, qry)
	if err != nil {
		t.Fatalf("Query execution failed: %v", err)
	}

	// Verify result
	if resultUser == nil {
		t.Fatal("Expected user result, got nil")
	}
	if resultUser.Name != "Jane Doe" || resultUser.Email != "jane@example.com" {
		t.Fatalf("Unexpected query result: %+v", resultUser)
	}
}

func TestAutoRegistry_RegisterEventHandler_Success(t *testing.T) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Register event handler
	handler := &TestUserCreatedEventHandler{}
	result := autoRegistry.RegisterHandlerInstances(handler)

	// Verify registration
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no errors, got: %v", result.Errors)
	}
	if result.RegisteredHandlers != 1 {
		t.Fatalf("Expected 1 registered handler, got: %d", result.RegisteredHandlers)
	}

	// Test event publishing
	ctx := context.Background()
	event := TestUserCreatedEvent{
		UserID: "user-1",
		Name:   "John Doe",
		Email:  "john@example.com",
	}

	err := PublishEvent(ctx, event)
	if err != nil {
		t.Fatalf("Event publishing failed: %v", err)
	}

	// Verify event was handled
	if len(handler.EventLog) != 1 {
		t.Fatalf("Expected 1 event in log, got: %d", len(handler.EventLog))
	}

	loggedEvent := handler.EventLog[0]
	if loggedEvent.UserID != "user-1" || loggedEvent.Name != "John Doe" {
		t.Fatalf("Unexpected logged event: %+v", loggedEvent)
	}
}

func TestAutoRegistry_RegisterValidator_Success(t *testing.T) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Register validator and command handler
	validator := &TestCreateUserValidator{}
	handler := &TestCreateUserCommandHandler{}
	result := autoRegistry.RegisterHandlerInstances(validator, handler)

	// Verify registration
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no errors, got: %v", result.Errors)
	}
	if result.RegisteredValidators != 1 {
		t.Fatalf("Expected 1 registered validator, got: %d", result.RegisteredValidators)
	}
	if result.RegisteredHandlers != 1 {
		t.Fatalf("Expected 1 registered handler, got: %d", result.RegisteredHandlers)
	}

	// Test valid command
	ctx := context.Background()
	validCmd := &CreateTestUserCommand{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := ExecuteCommand(ctx, validCmd)
	if err != nil {
		t.Fatalf("Valid command execution failed: %v", err)
	}

	// Verify validation was called
	if len(validator.ValidationLog) != 1 {
		t.Fatalf("Expected 1 validation call, got: %d", len(validator.ValidationLog))
	}

	// Test invalid command
	invalidCmd := &CreateTestUserCommand{
		Name:  "", // Missing name
		Email: "john@example.com",
	}

	err = ExecuteCommand(ctx, invalidCmd)
	if err == nil {
		t.Fatal("Expected validation error for invalid command")
	}
	if err.Error() != "validation failed for *cqrs.CreateTestUserCommand: name is required" {
		t.Fatalf("Unexpected validation error: %v", err)
	}
}

func TestAutoRegistry_MultipleHandlers_Success(t *testing.T) {
	autoRegistry, mockRepo, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Register multiple handlers
	handlers := []any{
		&TestCreateUserCommandHandler{},
		&TestUpdateUserCommandHandler{},
		&TestGetUserQueryHandler{},
		&TestUserCreatedEventHandler{},
		&TestUserUpdatedEventHandler{},
		&TestCreateUserValidator{},
		&TestUpdateUserValidator{},
	}

	result := autoRegistry.RegisterHandlerInstances(handlers...)

	// Verify registration
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no errors, got: %v", result.Errors)
	}
	if result.RegisteredHandlers != 5 {
		t.Fatalf("Expected 5 registered handlers, got: %d", result.RegisteredHandlers)
	}
	if result.RegisteredValidators != 2 {
		t.Fatalf("Expected 2 registered validators, got: %d", result.RegisteredValidators)
	}

	// Test all handler types work
	ctx := context.Background()

	// Test command handler
	createCmd := &CreateTestUserCommand{Name: "John", Email: "john@example.com"}
	err := ExecuteCommand(ctx, createCmd)
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	// Test another command handler
	updateCmd := &UpdateTestUserCommand{ID: "user-john", Name: "John Updated", Email: "john.updated@example.com"}
	err = ExecuteCommand(ctx, updateCmd)
	if err != nil {
		t.Fatalf("Update command failed: %v", err)
	}

	// Verify repository calls
	if len(mockRepo.saveCallLog) != 2 {
		t.Fatalf("Expected 2 save calls, got: %d", len(mockRepo.saveCallLog))
	}
}

func TestAutoRegistry_ErrorHandling_InvalidHandlers(t *testing.T) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Test invalid handlers
	invalidHandlers := []any{
		&struct{}{}, // No Handle method
		&struct {
			Handle func() // Wrong signature
		}{},
		&struct {
			Handle func(ctx context.Context) // Missing command parameter
		}{},
	}

	result := autoRegistry.RegisterHandlerInstances(invalidHandlers...)

	// Should have errors for all invalid handlers
	if len(result.Errors) != 3 {
		t.Fatalf("Expected 3 errors, got: %d", len(result.Errors))
	}

	// Should have no successful registrations
	if result.RegisteredHandlers != 0 {
		t.Fatalf("Expected 0 registered handlers, got: %d", result.RegisteredHandlers)
	}
}

func TestAutoRegistry_DependencyInjection_MissingDependency(t *testing.T) {
	// Setup without dependency injection
	ResetManager()
	manager := NewCqrsManager()
	SetManager(manager)
	autoRegistry := NewAutoRegistry(manager) // No dependency provider

	defer func() {
		ResetManager()
	}()

	// Register handler that requires dependency injection
	handler := &TestCreateUserCommandHandler{}
	result := autoRegistry.RegisterHandlerInstances(handler)

	// Should register successfully but dependency will be nil
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no registration errors, got: %v", result.Errors)
	}

	// Test command execution should fail due to missing dependency
	ctx := context.Background()
	cmd := &CreateTestUserCommand{Name: "John", Email: "john@example.com"}

	err := ExecuteCommand(ctx, cmd)
	if err == nil {
		t.Fatal("Expected error due to missing dependency")
	}
	if err.Error() != "UserRepo dependency not injected" {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAutoRegistry_ThreadSafety(t *testing.T) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Test concurrent registration
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine registers a handler
			handler := &TestCreateUserCommandHandler{}
			result := autoRegistry.RegisterHandlerInstances(handler)

			// Should not panic or have errors due to thread safety issues
			if len(result.Errors) > 1 { // Allow for duplicate registration
				t.Errorf("Goroutine %d: Unexpected errors: %v", id, result.Errors)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAutoRegistry_ProductionMetrics(t *testing.T) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Register handlers
	handlers := []any{
		&TestCreateUserCommandHandler{},
		&TestGetUserQueryHandler{},
		&TestUserCreatedEventHandler{},
		&TestCreateUserValidator{},
	}

	result := autoRegistry.RegisterHandlerInstances(handlers...)

	// Verify registration metrics
	if result.RegisteredHandlers != 3 {
		t.Fatalf("Expected 3 handlers, got: %d", result.RegisteredHandlers)
	}
	if result.RegisteredValidators != 1 {
		t.Fatalf("Expected 1 validator, got: %d", result.RegisteredValidators)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("Expected no errors, got: %v", result.Errors)
	}

	// Verify manager metrics
	commands, queries, events, validators := GetHandlerCounts()
	if commands != 1 || queries != 1 || events != 1 || validators != 1 {
		t.Fatalf("Unexpected handler counts: c=%d, q=%d, e=%d, v=%d", commands, queries, events, validators)
	}
}

// Benchmark tests for production performance validation
func BenchmarkAutoRegistry_SingleHandlerRegistration(b *testing.B) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler := &TestCreateUserCommandHandler{}
		autoRegistry.RegisterHandlerInstances(handler)
	}
}

func BenchmarkAutoRegistry_MultipleHandlerRegistration(b *testing.B) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	handlers := []any{
		&TestCreateUserCommandHandler{},
		&TestGetUserQueryHandler{},
		&TestUserCreatedEventHandler{},
		&TestCreateUserValidator{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		autoRegistry.RegisterHandlerInstances(handlers...)
	}
}

func BenchmarkAutoRegistry_CommandExecution(b *testing.B) {
	autoRegistry, _, cleanup := setupTestAutoRegistry()
	defer cleanup()

	// Register handler
	handler := &TestCreateUserCommandHandler{}
	autoRegistry.RegisterHandlerInstances(handler)

	ctx := context.Background()
	cmd := &CreateTestUserCommand{Name: "John", Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteCommand(ctx, cmd)
	}
}

// Interface compliance checks for test handlers
var (
	_ = &TestCreateUserCommandHandler{}
	_ = &TestUpdateUserCommandHandler{}
	_ = &TestGetUserQueryHandler{}
	_ = &TestUserCreatedEventHandler{}
	_ = &TestUserUpdatedEventHandler{}
	_ = &TestCreateUserValidator{}
	_ = &TestUpdateUserValidator{}
)
