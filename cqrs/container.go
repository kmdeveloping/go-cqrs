package cqrs

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

// SimpleContainer is a basic dependency injection container
type SimpleContainer struct {
	instances map[reflect.Type]any
	factories map[reflect.Type]func() any
	mu        sync.RWMutex
	logger    *log.Logger
}

// NewSimpleContainer creates a new dependency injection container
func NewSimpleContainer() *SimpleContainer {
	return &SimpleContainer{
		instances: make(map[reflect.Type]any),
		factories: make(map[reflect.Type]func() any),
		logger:    log.New(log.Writer(), "[DI-Container] ", log.LstdFlags),
	}
}

// SetLogger configures a custom logger for the container
func (c *SimpleContainer) SetLogger(logger *log.Logger) *SimpleContainer {
	c.logger = logger
	return c
}

// RegisterInstance registers a singleton instance for a type
func (c *SimpleContainer) RegisterInstance(instance any) *SimpleContainer {
	t := reflect.TypeOf(instance)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.instances[t] = instance
	c.logger.Printf("📝 Registered singleton instance: %s", t.Name())
	return c
}

// RegisterFactory registers a factory function for a type
func (c *SimpleContainer) RegisterFactory(factory any) *SimpleContainer {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	// Validate factory function signature: func() T
	if factoryType.Kind() != reflect.Func || factoryType.NumIn() != 0 || factoryType.NumOut() != 1 {
		c.logger.Printf("❌ Invalid factory function signature for %s", factoryType.String())
		return c
	}

	returnType := factoryType.Out(0)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.factories[returnType] = func() any {
		results := factoryValue.Call(nil)
		return results[0].Interface()
	}

	c.logger.Printf("🏭 Registered factory for: %s", returnType.Name())
	return c
}

// RegisterTransient registers a factory that creates new instances each time
func (c *SimpleContainer) RegisterTransient(constructor any) *SimpleContainer {
	return c.RegisterFactory(constructor)
}

// RegisterScoped registers an instance that is created once per scope (simplified as singleton here)
func (c *SimpleContainer) RegisterScoped(constructor any) *SimpleContainer {
	// For simplicity, treating scoped as singleton
	// In a real implementation, you'd need scope management
	return c.RegisterSingleton(constructor)
}

// RegisterSingleton registers a factory that creates one instance and reuses it
func (c *SimpleContainer) RegisterSingleton(constructor any) *SimpleContainer {
	factoryValue := reflect.ValueOf(constructor)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func || factoryType.NumIn() != 0 || factoryType.NumOut() != 1 {
		c.logger.Printf("❌ Invalid singleton constructor signature for %s", factoryType.String())
		return c
	}

	returnType := factoryType.Out(0)
	var once sync.Once
	var instance any

	c.mu.Lock()
	defer c.mu.Unlock()

	c.factories[returnType] = func() any {
		once.Do(func() {
			results := factoryValue.Call(nil)
			instance = results[0].Interface()
			c.logger.Printf("🔨 Created singleton instance: %s", returnType.Name())
		})
		return instance
	}

	c.logger.Printf("🔗 Registered singleton factory for: %s", returnType.Name())
	return c
}

// Resolve implements the DependencyProvider interface
func (c *SimpleContainer) Resolve(t reflect.Type) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// First check for registered instances
	if instance, exists := c.instances[t]; exists {
		c.logger.Printf("🎯 Resolved instance: %s", t.Name())
		return instance
	}

	// Then check for factories
	if factory, exists := c.factories[t]; exists {
		instance := factory()
		c.logger.Printf("🎯 Resolved via factory: %s", t.Name())
		return instance
	}

	// Try to resolve interface by implementation
	for registeredType, instance := range c.instances {
		if registeredType.Implements(t) || registeredType.AssignableTo(t) {
			c.logger.Printf("🎯 Resolved by interface: %s -> %s", t.Name(), registeredType.Name())
			return instance
		}
	}

	for registeredType, factory := range c.factories {
		if registeredType.Implements(t) || registeredType.AssignableTo(t) {
			instance := factory()
			c.logger.Printf("🎯 Resolved by interface via factory: %s -> %s", t.Name(), registeredType.Name())
			return instance
		}
	}

	c.logger.Printf("❓ Could not resolve: %s", t.Name())
	return nil
}

// MustResolve resolves a dependency or panics if not found
func (c *SimpleContainer) MustResolve(t reflect.Type) any {
	instance := c.Resolve(t)
	if instance == nil {
		panic(fmt.Sprintf("failed to resolve dependency: %s", t.Name()))
	}
	return instance
}

// Clear removes all registrations (useful for testing)
func (c *SimpleContainer) Clear() *SimpleContainer {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.instances = make(map[reflect.Type]any)
	c.factories = make(map[reflect.Type]func() any)
	c.logger.Printf("🧹 Container cleared")
	return c
}

// GetRegistrationCount returns the number of registered types
func (c *SimpleContainer) GetRegistrationCount() (instances, factories int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.instances), len(c.factories)
}

// ListRegistrations logs all registered types for debugging
func (c *SimpleContainer) ListRegistrations() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.logger.Printf("📋 Container registrations:")
	c.logger.Printf("   Instances (%d):", len(c.instances))
	for t := range c.instances {
		c.logger.Printf("     - %s", t.Name())
	}

	c.logger.Printf("   Factories (%d):", len(c.factories))
	for t := range c.factories {
		c.logger.Printf("     - %s", t.Name())
	}
}

// Generic helper functions for type-safe registration

// Register registers an instance with compile-time type safety
func Register[T any](container *SimpleContainer, instance T) *SimpleContainer {
	return container.RegisterInstance(instance)
}

// RegisterFunc registers a factory function with compile-time type safety
func RegisterFunc[T any](container *SimpleContainer, factory func() T) *SimpleContainer {
	return container.RegisterFactory(factory)
}

// RegisterSingletonFunc registers a singleton factory with compile-time type safety
func RegisterSingletonFunc[T any](container *SimpleContainer, factory func() T) *SimpleContainer {
	return container.RegisterSingleton(factory)
}

// Resolve resolves a dependency with compile-time type safety
func Resolve[T any](container *SimpleContainer) T {
	var zero T
	t := reflect.TypeOf(zero)
	instance := container.Resolve(t)
	if instance == nil {
		return zero
	}
	return instance.(T)
}

// MustResolve resolves a dependency with compile-time type safety or panics
func MustResolve[T any](container *SimpleContainer) T {
	var zero T
	t := reflect.TypeOf(zero)
	instance := container.MustResolve(t)
	return instance.(T)
}
