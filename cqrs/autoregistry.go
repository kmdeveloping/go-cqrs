package cqrs

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"

	"github.com/kmdeveloping/go-cqrs/decorators"
)

// DependencyProvider interface for injecting dependencies into handlers
type DependencyProvider interface {
	// Resolve returns an instance of the requested type
	// Returns nil if the type cannot be resolved
	Resolve(t reflect.Type) any
}

// AutoRegistry provides runtime handler discovery and registration
type AutoRegistry struct {
	manager            *Manager
	dependencyProvider DependencyProvider
	logger             *log.Logger
	registeredTypes    map[reflect.Type]bool // Track registered types to avoid duplicates
	mu                 sync.RWMutex
}

// RegistrationResult contains information about the registration process
type RegistrationResult struct {
	RegisteredHandlers   int
	RegisteredValidators int
	Errors               []error
	Details              []string
}

// Add error to registration result with user-friendly formatting
func (r *RegistrationResult) addError(err error, handlerType reflect.Type) {
	r.Errors = append(r.Errors, err)
	r.Details = append(r.Details, fmt.Sprintf("❌ Failed to register %s: %v",
		handlerType.Name(), err))
}

// Add success to registration result
func (r *RegistrationResult) addSuccess(handlerType reflect.Type, category string) {
	switch category {
	case "command handler", "query handler", "event handler":
		r.RegisteredHandlers++
	case "validator":
		r.RegisteredValidators++
	}
	r.Details = append(r.Details, fmt.Sprintf("✅ Registered %s %s",
		category, handlerType.Name()))
}

// NewAutoRegistry creates a new auto-registry instance
func NewAutoRegistry(manager *Manager) *AutoRegistry {
	return &AutoRegistry{
		manager:         manager,
		logger:          log.New(log.Writer(), "[CQRS-AutoRegistry] ", log.LstdFlags),
		registeredTypes: make(map[reflect.Type]bool),
	}
}

// SetDependencyProvider configures the dependency injection provider
func (ar *AutoRegistry) SetDependencyProvider(provider DependencyProvider) *AutoRegistry {
	ar.dependencyProvider = provider
	return ar
}

// SetLogger configures a custom logger for user-readable output
func (ar *AutoRegistry) SetLogger(logger *log.Logger) *AutoRegistry {
	ar.logger = logger
	return ar
}

// RegisterFromPackage discovers and registers all handlers from specified packages
// This is the main entry point for runtime auto-registration
func (ar *AutoRegistry) RegisterFromPackage(packages ...any) *RegistrationResult {
	result := &RegistrationResult{}

	ar.logger.Printf("🚀 Starting auto-registration for %d packages", len(packages))

	for _, pkg := range packages {
		ar.discoverAndRegisterFromPackage(pkg, result)
	}

	ar.logRegistrationSummary(result)
	return result
}

// RegisterHandlerInstances registers pre-created handler instances
// Useful when you have existing instances with injected dependencies
func (ar *AutoRegistry) RegisterHandlerInstances(handlers ...any) *RegistrationResult {
	result := &RegistrationResult{}

	ar.logger.Printf("🚀 Registering %d handler instances", len(handlers))

	for _, handler := range handlers {
		ar.registerHandlerInstance(handler, result)
	}

	ar.logRegistrationSummary(result)
	return result
}

// discoverAndRegisterFromPackage scans a package for handler types and registers them
func (ar *AutoRegistry) discoverAndRegisterFromPackage(pkg any, result *RegistrationResult) {
	// Get the package path from any type in the package
	pkgType := reflect.TypeOf(pkg)
	if pkgType.Kind() == reflect.Ptr {
		pkgType = pkgType.Elem()
	}

	pkgPath := pkgType.PkgPath()
	ar.logger.Printf("📦 Scanning package: %s", pkgPath)

	// This is a simplified approach - in a real implementation, you might use:
	// - go/types package for full AST analysis
	// - Build tags and reflection registry
	// - Package scanning utilities

	// For now, we demonstrate with the provided handler instance
	ar.registerHandlerInstance(pkg, result)
}

// registerHandlerInstance attempts to register a single handler instance
func (ar *AutoRegistry) registerHandlerInstance(handler any, result *RegistrationResult) {
	handlerType := reflect.TypeOf(handler)
	handlerValue := reflect.ValueOf(handler)

	// Skip if already registered
	ar.mu.Lock()
	if ar.registeredTypes[handlerType] {
		ar.mu.Unlock()
		ar.logger.Printf("⏭️  Skipping already registered handler: %s", handlerType.Name())
		return
	}
	ar.registeredTypes[handlerType] = true
	ar.mu.Unlock()

	// Create instance with dependency injection if needed
	instance, err := ar.createHandlerInstance(handlerType, handlerValue)
	if err != nil {
		result.addError(fmt.Errorf("dependency injection failed: %w", err), handlerType)
		return
	}

	// Try to register as different handler types
	registered := false

	// Try command handler first
	if ar.implementsCommandHandler(handlerType) {
		if err := ar.registerAsCommandHandler(instance); err != nil {
			result.addError(err, handlerType)
		} else {
			result.addSuccess(handlerType, "command handler")
			registered = true
		}
	}

	// Try query handler
	if ar.implementsQueryHandler(handlerType) {
		if err := ar.registerAsQueryHandler(instance); err != nil {
			result.addError(err, handlerType)
		} else {
			result.addSuccess(handlerType, "query handler")
			registered = true
		}
	}

	// Try event handler
	if ar.implementsEventHandler(handlerType) {
		if err := ar.registerAsEventHandler(instance); err != nil {
			result.addError(err, handlerType)
		} else {
			result.addSuccess(handlerType, "event handler")
			registered = true
		}
	}

	// Try validator
	if ar.implementsValidator(handlerType) {
		if err := ar.registerAsValidator(instance); err != nil {
			result.addError(err, handlerType)
		} else {
			result.addSuccess(handlerType, "validator")
			registered = true
		}
	}

	// If nothing was registered, check if it has Handle or Validate methods but wrong signatures
	if !registered {
		if _, hasHandle := handlerType.MethodByName("Handle"); hasHandle {
			result.addError(fmt.Errorf("handler %s has Handle method but invalid signature for command/query/event handler", handlerType.Name()), handlerType)
		} else if _, hasValidate := handlerType.MethodByName("Validate"); hasValidate {
			result.addError(fmt.Errorf("handler %s has Validate method but invalid signature for validator", handlerType.Name()), handlerType)
		} else {
			result.addError(fmt.Errorf("handler %s does not implement any recognized handler interface", handlerType.Name()), handlerType)
		}
	}
}

// createHandlerInstance creates or injects dependencies into a handler instance
func (ar *AutoRegistry) createHandlerInstance(handlerType reflect.Type, handlerValue reflect.Value) (any, error) {
	// If we already have a valid instance, check if it needs dependency injection
	if handlerValue.IsValid() && !handlerValue.IsNil() {
		// Attempt dependency injection on existing instance
		if err := ar.injectDependencies(handlerValue); err != nil {
			return nil, fmt.Errorf("failed to inject dependencies: %w", err)
		}
		return handlerValue.Interface(), nil
	}

	// Create new instance
	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}

	newInstance := reflect.New(handlerType)

	// Inject dependencies into the new instance
	if err := ar.injectDependencies(newInstance); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies into new instance: %w", err)
	}

	return newInstance.Interface(), nil
}

// injectDependencies injects dependencies into handler fields
func (ar *AutoRegistry) injectDependencies(handlerValue reflect.Value) error {
	if ar.dependencyProvider == nil {
		return nil // No dependency injection configured
	}

	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}

	if handlerValue.Kind() != reflect.Struct {
		return nil // Only inject into structs
	}

	handlerType := handlerValue.Type()

	for i := 0; i < handlerValue.NumField(); i++ {
		field := handlerValue.Field(i)
		fieldType := handlerType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Check for dependency injection tag
		if tag := fieldType.Tag.Get("inject"); tag != "" || (field.IsNil() && field.Type().Kind() == reflect.Interface) {
			// Only try to inject if the field type is an interface or pointer to a struct
			fieldKind := field.Type().Kind()

			// Skip primitive types, slices, arrays, maps, channels, and funcs
			if fieldKind == reflect.Slice || fieldKind == reflect.Array ||
				fieldKind == reflect.Map || fieldKind == reflect.Chan ||
				fieldKind == reflect.Func || fieldKind == reflect.String ||
				fieldKind == reflect.Int || fieldKind == reflect.Int8 ||
				fieldKind == reflect.Int16 || fieldKind == reflect.Int32 ||
				fieldKind == reflect.Int64 || fieldKind == reflect.Uint ||
				fieldKind == reflect.Uint8 || fieldKind == reflect.Uint16 ||
				fieldKind == reflect.Uint32 || fieldKind == reflect.Uint64 ||
				fieldKind == reflect.Float32 || fieldKind == reflect.Float64 ||
				fieldKind == reflect.Bool {
				continue
			}

			dependency := ar.dependencyProvider.Resolve(field.Type())
			if dependency != nil {
				field.Set(reflect.ValueOf(dependency))
				ar.logger.Printf("💉 Injected dependency %s into %s.%s",
					field.Type().Name(), handlerType.Name(), fieldType.Name)
			}
		}
	}

	return nil
}

// Check if type implements command handler interface
func (ar *AutoRegistry) implementsCommandHandler(t reflect.Type) bool {
	// Look for Handle(context.Context, *T) error method
	method, found := t.MethodByName("Handle")
	if !found {
		return false
	}

	methodType := method.Type
	// Method signature: func (receiver) Handle(ctx context.Context, cmd *T) error
	if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
		return false
	}

	// Check parameter types
	ctxType := methodType.In(1)
	cmdType := methodType.In(2)
	errType := methodType.Out(0)

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	return ctxType.Implements(contextType) &&
		cmdType.Kind() == reflect.Ptr &&
		errType.Implements(errorType)
}

// Check if type implements query handler interface
func (ar *AutoRegistry) implementsQueryHandler(t reflect.Type) bool {
	method, found := t.MethodByName("Handle")
	if !found {
		return false
	}

	methodType := method.Type
	// Method signature: func (receiver) Handle(ctx context.Context, query T) (R, error)
	if methodType.NumIn() != 3 || methodType.NumOut() != 2 {
		return false
	}

	ctxType := methodType.In(1)
	errType := methodType.Out(1)

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	return ctxType.Implements(contextType) && errType.Implements(errorType)
}

// Check if type implements event handler interface
func (ar *AutoRegistry) implementsEventHandler(t reflect.Type) bool {
	method, found := t.MethodByName("Handle")
	if !found {
		return false
	}

	methodType := method.Type
	// Method signature: func (receiver) Handle(ctx context.Context, event T) error
	if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
		return false
	}

	ctxType := methodType.In(1)
	eventType := methodType.In(2)
	errType := methodType.Out(0)

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	return ctxType.Implements(contextType) &&
		eventType.Kind() != reflect.Ptr && // Events are passed by value
		errType.Implements(errorType)
}

// Check if type implements validator interface
func (ar *AutoRegistry) implementsValidator(t reflect.Type) bool {
	method, found := t.MethodByName("Validate")
	if !found {
		return false
	}

	methodType := method.Type
	// Method signature: func (receiver) Validate(ctx context.Context, cmd *T) error
	if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
		return false
	}

	ctxType := methodType.In(1)
	cmdType := methodType.In(2)
	errType := methodType.Out(0)

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	return ctxType.Implements(contextType) &&
		cmdType.Kind() == reflect.Ptr &&
		errType.Implements(errorType)
}

// Register instance as command handler using reflection
func (ar *AutoRegistry) registerAsCommandHandler(instance any) error {
	handlerType := reflect.TypeOf(instance)
	handlerValue := reflect.ValueOf(instance)

	// Find the Handle method
	method, found := handlerType.MethodByName("Handle")
	if !found {
		return fmt.Errorf("command handler %s missing Handle method", handlerType.Name())
	}

	methodType := method.Type
	// Expected signature: func (receiver) Handle(ctx context.Context, cmd *T) error
	if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
		return fmt.Errorf("command handler %s has invalid Handle method signature", handlerType.Name())
	}

	// Extract command type from second parameter (first is receiver, second is context, third is command)
	cmdType := methodType.In(2)
	if cmdType.Kind() != reflect.Ptr {
		return fmt.Errorf("command handler %s command parameter must be a pointer", handlerType.Name())
	}

	// Get the actual command type (without pointer)
	cmdElemType := cmdType.Elem()

	// Validate return type is error
	errType := methodType.Out(0)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !errType.Implements(errorInterface) {
		return fmt.Errorf("command handler %s Handle method must return error", handlerType.Name())
	}

	// Apply decorators
	ar.manager.decoratorsMu.RLock()
	decoratorsCopy := make([]decorators.HandlerDecorator, len(ar.manager.decorators))
	copy(decoratorsCopy, ar.manager.decorators)
	ar.manager.decoratorsMu.RUnlock()

	// Create a wrapper that matches the command handler interface
	wrapper := &reflectionCommandHandlerWrapper{
		handler:      handlerValue,
		handleMethod: method,
		cmdType:      cmdType,
	}

	// Wrap with decorators using the base decorator function
	base := decorators.HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		if reflect.TypeOf(msg) != cmdType {
			return nil, fmt.Errorf("invalid command type: expected %v, got %v", cmdType, reflect.TypeOf(msg))
		}
		return nil, wrapper.Handle(ctx, msg)
	})

	decorated := decorators.WithDecorators(base, decoratorsCopy...)

	// Register directly with manager's command handlers map
	ar.manager.handlersMu.Lock()
	defer ar.manager.handlersMu.Unlock()
	ar.manager.commandHandlers[cmdType] = &decoratedCommandHandler{
		decorated: decorated,
		cmdType:   cmdType,
	}

	ar.logger.Printf("✅ Registered command handler %s for command type %s", handlerType.Name(), cmdElemType.Name())
	return nil
}

// Register instance as query handler using reflection
func (ar *AutoRegistry) registerAsQueryHandler(instance any) error {
	handlerType := reflect.TypeOf(instance)
	handlerValue := reflect.ValueOf(instance)

	// Find the Handle method
	method, found := handlerType.MethodByName("Handle")
	if !found {
		return fmt.Errorf("query handler %s missing Handle method", handlerType.Name())
	}

	methodType := method.Type
	// Expected signature: func (receiver) Handle(ctx context.Context, query T) (R, error)
	if methodType.NumIn() != 3 || methodType.NumOut() != 2 {
		return fmt.Errorf("query handler %s has invalid Handle method signature", handlerType.Name())
	}

	// Extract query and result types
	queryType := methodType.In(2)
	resultType := methodType.Out(0)

	// Validate return types
	errType := methodType.Out(1)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !errType.Implements(errorInterface) {
		return fmt.Errorf("query handler %s Handle method must return (R, error)", handlerType.Name())
	}

	// Apply decorators
	ar.manager.decoratorsMu.RLock()
	decoratorsCopy := make([]decorators.HandlerDecorator, len(ar.manager.decorators))
	copy(decoratorsCopy, ar.manager.decorators)
	ar.manager.decoratorsMu.RUnlock()

	// Create a wrapper that matches the query handler interface
	wrapper := &reflectionQueryHandlerWrapper{
		handler:      handlerValue,
		handleMethod: method,
		queryType:    queryType,
		resultType:   resultType,
	}

	// Wrap with decorators
	base := decorators.HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		if reflect.TypeOf(msg) != queryType {
			return nil, fmt.Errorf("invalid query type: expected %v, got %v", queryType, reflect.TypeOf(msg))
		}
		return wrapper.Handle(ctx, msg)
	})

	decorated := decorators.WithDecorators(base, decoratorsCopy...)

	// Register directly with manager's query handlers map
	ar.manager.handlersMu.Lock()
	defer ar.manager.handlersMu.Unlock()
	ar.manager.queryHandlers[queryType] = &decoratedQueryHandler{
		decorated:  decorated,
		queryType:  queryType,
		resultType: resultType,
	}

	ar.logger.Printf("✅ Registered query handler %s for query type %s -> %s",
		handlerType.Name(), queryType.Name(), resultType.Name())
	return nil
}

// Register instance as event handler using reflection
func (ar *AutoRegistry) registerAsEventHandler(instance any) error {
	handlerType := reflect.TypeOf(instance)
	handlerValue := reflect.ValueOf(instance)

	// Find the Handle method
	method, found := handlerType.MethodByName("Handle")
	if !found {
		return fmt.Errorf("event handler %s missing Handle method", handlerType.Name())
	}

	methodType := method.Type
	// Expected signature: func (receiver) Handle(ctx context.Context, event T) error
	if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
		return fmt.Errorf("event handler %s has invalid Handle method signature", handlerType.Name())
	}

	// Extract event type from second parameter
	eventType := methodType.In(2)
	if eventType.Kind() == reflect.Ptr {
		return fmt.Errorf("event handler %s event parameter should not be a pointer", handlerType.Name())
	}

	// Validate return type is error
	errType := methodType.Out(0)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !errType.Implements(errorInterface) {
		return fmt.Errorf("event handler %s Handle method must return error", handlerType.Name())
	}

	// Apply decorators
	ar.manager.decoratorsMu.RLock()
	decoratorsCopy := make([]decorators.HandlerDecorator, len(ar.manager.decorators))
	copy(decoratorsCopy, ar.manager.decorators)
	ar.manager.decoratorsMu.RUnlock()

	// Create a wrapper that matches the event handler interface
	wrapper := &reflectionEventHandlerWrapper{
		handler:      handlerValue,
		handleMethod: method,
		eventType:    eventType,
	}

	// Wrap with decorators
	base := decorators.HandlerDecoratorFunc(func(ctx context.Context, msg any) (any, error) {
		if reflect.TypeOf(msg) != eventType {
			return nil, fmt.Errorf("invalid event type: expected %v, got %v", eventType, reflect.TypeOf(msg))
		}
		return nil, wrapper.Handle(ctx, msg)
	})

	decorated := decorators.WithDecorators(base, decoratorsCopy...)

	// Register directly with manager's event handlers map
	ar.manager.handlersMu.Lock()
	defer ar.manager.handlersMu.Unlock()
	if ar.manager.eventHandlers[eventType] == nil {
		ar.manager.eventHandlers[eventType] = make([]any, 0)
	}
	ar.manager.eventHandlers[eventType] = append(ar.manager.eventHandlers[eventType], &decoratedEventHandler{
		decorated: decorated,
		eventType: eventType,
	})

	ar.logger.Printf("✅ Registered event handler %s for event type %s", handlerType.Name(), eventType.Name())
	return nil
}

// Register instance as validator using reflection
func (ar *AutoRegistry) registerAsValidator(instance any) error {
	handlerType := reflect.TypeOf(instance)
	handlerValue := reflect.ValueOf(instance)

	// Find the Validate method
	method, found := handlerType.MethodByName("Validate")
	if !found {
		return fmt.Errorf("validator %s missing Validate method", handlerType.Name())
	}

	methodType := method.Type
	// Expected signature: func (receiver) Validate(ctx context.Context, cmd *T) error
	if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
		return fmt.Errorf("validator %s has invalid Validate method signature", handlerType.Name())
	}

	// Extract command type from second parameter
	cmdType := methodType.In(2)
	if cmdType.Kind() != reflect.Ptr {
		return fmt.Errorf("validator %s command parameter must be a pointer", handlerType.Name())
	}

	// Validate return type is error
	errType := methodType.Out(0)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !errType.Implements(errorInterface) {
		return fmt.Errorf("validator %s Validate method must return error", handlerType.Name())
	}

	// Create a simple reflection-based validator that works with the CQRS system
	simpleValidator := &simpleReflectionValidator{
		handlerValue: handlerValue,
		cmdType:      cmdType,
	}

	// Register directly with manager's validators map using the pointer type
	ar.manager.validatorsMu.Lock()
	defer ar.manager.validatorsMu.Unlock()
	ar.manager.validators[cmdType] = append(ar.manager.validators[cmdType], simpleValidator)

	ar.logger.Printf("✅ Registered validator %s for command type %s",
		handlerType.Name(), cmdType.Elem().Name())
	return nil
}

// Reflection wrapper types for production-ready implementation

type reflectionCommandHandlerWrapper struct {
	handler      reflect.Value
	handleMethod reflect.Method
	cmdType      reflect.Type
}

func (w *reflectionCommandHandlerWrapper) Handle(ctx context.Context, cmd any) error {
	// Validate command type
	if reflect.TypeOf(cmd) != w.cmdType {
		return fmt.Errorf("invalid command type: expected %v, got %v", w.cmdType, reflect.TypeOf(cmd))
	}

	// Call the Handle method using reflection
	results := w.handler.MethodByName("Handle").Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(cmd),
	})

	// Check if error was returned
	if len(results) != 1 {
		return fmt.Errorf("unexpected number of return values from Handle method")
	}

	if !results[0].IsNil() {
		if err, ok := results[0].Interface().(error); ok {
			return err
		}
		return fmt.Errorf("Handle method returned non-error value")
	}

	return nil
}

type reflectionQueryHandlerWrapper struct {
	handler      reflect.Value
	handleMethod reflect.Method
	queryType    reflect.Type
	resultType   reflect.Type
}

func (w *reflectionQueryHandlerWrapper) Handle(ctx context.Context, query any) (any, error) {
	// Validate query type
	if reflect.TypeOf(query) != w.queryType {
		return nil, fmt.Errorf("invalid query type: expected %v, got %v", w.queryType, reflect.TypeOf(query))
	}

	// Call the Handle method using reflection
	results := w.handler.MethodByName("Handle").Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(query),
	})

	// Check return values
	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected number of return values from Handle method")
	}

	result := results[0].Interface()
	errValue := results[1]

	var err error
	if !errValue.IsNil() {
		if e, ok := errValue.Interface().(error); ok {
			err = e
		} else {
			err = fmt.Errorf("Handle method returned non-error in error position")
		}
	}

	return result, err
}

type reflectionEventHandlerWrapper struct {
	handler      reflect.Value
	handleMethod reflect.Method
	eventType    reflect.Type
}

func (w *reflectionEventHandlerWrapper) Handle(ctx context.Context, event any) error {
	// Validate event type
	if reflect.TypeOf(event) != w.eventType {
		return fmt.Errorf("invalid event type: expected %v, got %v", w.eventType, reflect.TypeOf(event))
	}

	// Call the Handle method using reflection
	results := w.handler.MethodByName("Handle").Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(event),
	})

	// Check if error was returned
	if len(results) != 1 {
		return fmt.Errorf("unexpected number of return values from Handle method")
	}

	if !results[0].IsNil() {
		if err, ok := results[0].Interface().(error); ok {
			return err
		}
		return fmt.Errorf("Handle method returned non-error value")
	}

	return nil
}

type simpleReflectionValidator struct {
	handlerValue reflect.Value
	cmdType      reflect.Type
}

func (v *simpleReflectionValidator) Validate(ctx context.Context, cmd any) error {
	// Validate command type
	if reflect.TypeOf(cmd) != v.cmdType {
		return fmt.Errorf("invalid command type: expected %v, got %v", v.cmdType, reflect.TypeOf(cmd))
	}

	// Call the Validate method using reflection
	results := v.handlerValue.MethodByName("Validate").Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(cmd),
	})

	// Check if error was returned
	if len(results) != 1 {
		return fmt.Errorf("unexpected number of return values from Validate method")
	}

	if !results[0].IsNil() {
		if err, ok := results[0].Interface().(error); ok {
			return err
		}
		return fmt.Errorf("Validate method returned non-error value")
	}

	return nil
}

// Decorated handler wrappers for production integration

type decoratedCommandHandler struct {
	decorated decorators.IHandlerDecorator
	cmdType   reflect.Type
}

func (d *decoratedCommandHandler) Handle(ctx context.Context, cmd any) error {
	_, err := d.decorated.Handle(ctx, cmd)
	return err
}

type decoratedQueryHandler struct {
	decorated  decorators.IHandlerDecorator
	queryType  reflect.Type
	resultType reflect.Type
}

func (d *decoratedQueryHandler) Handle(ctx context.Context, query any) (any, error) {
	return d.decorated.Handle(ctx, query)
}

type decoratedEventHandler struct {
	decorated decorators.IHandlerDecorator
	eventType reflect.Type
}

func (d *decoratedEventHandler) Handle(ctx context.Context, event any) error {
	_, err := d.decorated.Handle(ctx, event)
	return err
}

// logRegistrationSummary provides user-friendly summary of registration results
func (ar *AutoRegistry) logRegistrationSummary(result *RegistrationResult) {
	ar.logger.Printf("📊 Registration Summary:")
	ar.logger.Printf("   ✅ Handlers: %d", result.RegisteredHandlers)
	ar.logger.Printf("   ✅ Validators: %d", result.RegisteredValidators)
	ar.logger.Printf("   ❌ Errors: %d", len(result.Errors))

	// Log details
	for _, detail := range result.Details {
		ar.logger.Printf("   %s", detail)
	}

	// Log errors with stack traces for debugging
	for _, err := range result.Errors {
		ar.logger.Printf("   ❌ Error: %v", err)
		if ar.logger.Writer() == log.Writer() { // Only show stack trace in debug mode
			ar.logger.Printf("      Stack: %s", getStackTrace())
		}
	}
}

// getStackTrace returns a simplified stack trace for debugging
func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Clean API functions for easy usage

// AutoRegisterFromPackage provides a clean API for package-based auto-registration
func AutoRegisterFromPackage(packages ...any) *RegistrationResult {
	autoRegistry := NewAutoRegistry(GetManager())
	return autoRegistry.RegisterFromPackage(packages...)
}

// AutoRegisterHandlers provides a clean API for instance-based auto-registration
func AutoRegisterHandlers(handlers ...any) *RegistrationResult {
	autoRegistry := NewAutoRegistry(GetManager())
	return autoRegistry.RegisterHandlerInstances(handlers...)
}

// AutoRegisterWithDependencies combines auto-registration with dependency injection
func AutoRegisterWithDependencies(provider DependencyProvider, packages ...any) *RegistrationResult {
	autoRegistry := NewAutoRegistry(GetManager()).SetDependencyProvider(provider)
	return autoRegistry.RegisterFromPackage(packages...)
}
