package cqrs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// AutoRegistryConfig holds production configuration for the auto-registry
type AutoRegistryConfig struct {
	// EnableMetrics enables detailed metrics collection
	EnableMetrics bool

	// EnableHealthChecks enables health check endpoints
	EnableHealthChecks bool

	// MaxRegistrationTime sets maximum time allowed for registration
	MaxRegistrationTime time.Duration

	// EnableDeadlockDetection enables goroutine deadlock detection
	EnableDeadlockDetection bool

	// EnableRetry enables retry mechanism for failed registrations
	EnableRetry bool
	MaxRetries  int
	RetryDelay  time.Duration

	// LogLevel controls logging verbosity
	LogLevel LogLevel

	// EnableValidation enables strict validation of handler signatures
	EnableValidation bool

	// RequireDependencyInjection fails registration if dependencies are missing
	RequireDependencyInjection bool

	// CustomLogger allows providing custom logger
	CustomLogger *log.Logger
}

// LogLevel defines logging levels
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// DefaultProductionConfig returns a production-ready configuration
func DefaultProductionConfig() *AutoRegistryConfig {
	return &AutoRegistryConfig{
		EnableMetrics:              true,
		EnableHealthChecks:         true,
		MaxRegistrationTime:        time.Second * 30,
		EnableDeadlockDetection:    true,
		EnableRetry:                true,
		MaxRetries:                 3,
		RetryDelay:                 time.Millisecond * 100,
		LogLevel:                   LogLevelInfo,
		EnableValidation:           true,
		RequireDependencyInjection: false,
	}
}

// DefaultDevelopmentConfig returns a development-friendly configuration
func DefaultDevelopmentConfig() *AutoRegistryConfig {
	return &AutoRegistryConfig{
		EnableMetrics:              true,
		EnableHealthChecks:         false,
		MaxRegistrationTime:        time.Minute * 5,
		EnableDeadlockDetection:    false,
		EnableRetry:                false,
		MaxRetries:                 0,
		RetryDelay:                 0,
		LogLevel:                   LogLevelDebug,
		EnableValidation:           true,
		RequireDependencyInjection: false,
	}
}

// ProductionAutoRegistry is a production-ready version of AutoRegistry
type ProductionAutoRegistry struct {
	*AutoRegistry
	config    *AutoRegistryConfig
	metrics   *RegistryMetrics
	mu        sync.RWMutex
	startTime time.Time
}

// RegistryMetrics holds production metrics
type RegistryMetrics struct {
	TotalRegistrations      int64
	SuccessfulRegistrations int64
	FailedRegistrations     int64
	TotalRegistrationTime   time.Duration
	AverageRegistrationTime time.Duration
	LastRegistrationTime    time.Time

	HandlerTypes map[string]int64 // command, query, event, validator

	DependencyInjections int64
	ValidationFailures   int64
	RetryAttempts        int64

	mu sync.RWMutex
}

// NewProductionAutoRegistry creates a production-ready auto-registry
func NewProductionAutoRegistry(manager *Manager, config *AutoRegistryConfig) *ProductionAutoRegistry {
	if config == nil {
		config = DefaultProductionConfig()
	}

	autoRegistry := NewAutoRegistry(manager)

	// Set custom logger if provided
	if config.CustomLogger != nil {
		autoRegistry.SetLogger(config.CustomLogger)
	}

	return &ProductionAutoRegistry{
		AutoRegistry: autoRegistry,
		config:       config,
		metrics: &RegistryMetrics{
			HandlerTypes: make(map[string]int64),
		},
		startTime: time.Now(),
	}
}

// RegisterHandlerInstancesWithConfig registers handlers with production configuration
func (pr *ProductionAutoRegistry) RegisterHandlerInstancesWithConfig(handlers ...any) *RegistrationResult {
	start := time.Now()

	// Create timeout context if configured
	ctx := context.Background()
	if pr.config.MaxRegistrationTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pr.config.MaxRegistrationTime)
		defer cancel()
	}

	// Use a channel to handle timeout
	resultChan := make(chan *RegistrationResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("registration panic: %v", r)
			}
		}()

		result := pr.registerWithRetry(handlers...)
		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		pr.updateMetrics(result, time.Since(start))
		return result
	case err := <-errorChan:
		result := &RegistrationResult{
			Errors:  []error{err},
			Details: []string{fmt.Sprintf("❌ Registration error: %v", err)},
		}
		pr.updateMetrics(result, time.Since(start))
		return result
	case <-ctx.Done():
		result := &RegistrationResult{
			Errors:  []error{fmt.Errorf("registration timeout after %v", pr.config.MaxRegistrationTime)},
			Details: []string{fmt.Sprintf("❌ Registration timeout after %v", pr.config.MaxRegistrationTime)},
		}
		pr.updateMetrics(result, time.Since(start))
		return result
	}
}

// registerWithRetry implements retry logic for failed registrations
func (pr *ProductionAutoRegistry) registerWithRetry(handlers ...any) *RegistrationResult {
	var lastResult *RegistrationResult

	for attempt := 0; attempt <= pr.config.MaxRetries; attempt++ {
		if attempt > 0 && pr.config.EnableRetry {
			pr.metrics.mu.Lock()
			pr.metrics.RetryAttempts++
			pr.metrics.mu.Unlock()

			time.Sleep(pr.config.RetryDelay)
		}

		result := pr.AutoRegistry.RegisterHandlerInstances(handlers...)

		// If no errors or retry is disabled, return result
		if len(result.Errors) == 0 || !pr.config.EnableRetry {
			return result
		}

		lastResult = result

		// Log retry attempt
		if pr.config.LogLevel >= LogLevelWarn {
			pr.AutoRegistry.logger.Printf("⚠️  Registration attempt %d failed, retrying...", attempt+1)
		}
	}

	return lastResult
}

// updateMetrics updates internal metrics
func (pr *ProductionAutoRegistry) updateMetrics(result *RegistrationResult, duration time.Duration) {
	if !pr.config.EnableMetrics {
		return
	}

	pr.metrics.mu.Lock()
	defer pr.metrics.mu.Unlock()

	pr.metrics.TotalRegistrations++
	pr.metrics.TotalRegistrationTime += duration
	pr.metrics.AverageRegistrationTime = pr.metrics.TotalRegistrationTime / time.Duration(pr.metrics.TotalRegistrations)
	pr.metrics.LastRegistrationTime = time.Now()

	if len(result.Errors) == 0 {
		pr.metrics.SuccessfulRegistrations++
	} else {
		pr.metrics.FailedRegistrations++
	}

	// Count handler types
	pr.metrics.HandlerTypes["command"] += int64(result.RegisteredHandlers)
	pr.metrics.HandlerTypes["validator"] += int64(result.RegisteredValidators)
}

// GetMetrics returns current metrics
func (pr *ProductionAutoRegistry) GetMetrics() *RegistryMetrics {
	pr.metrics.mu.RLock()
	defer pr.metrics.mu.RUnlock()

	// Return a copy to avoid data races
	metricsCopy := &RegistryMetrics{
		TotalRegistrations:      pr.metrics.TotalRegistrations,
		SuccessfulRegistrations: pr.metrics.SuccessfulRegistrations,
		FailedRegistrations:     pr.metrics.FailedRegistrations,
		TotalRegistrationTime:   pr.metrics.TotalRegistrationTime,
		AverageRegistrationTime: pr.metrics.AverageRegistrationTime,
		LastRegistrationTime:    pr.metrics.LastRegistrationTime,
		DependencyInjections:    pr.metrics.DependencyInjections,
		ValidationFailures:      pr.metrics.ValidationFailures,
		RetryAttempts:           pr.metrics.RetryAttempts,
		HandlerTypes:            make(map[string]int64),
	}

	for k, v := range pr.metrics.HandlerTypes {
		metricsCopy.HandlerTypes[k] = v
	}

	return metricsCopy
}

// HealthCheck performs a health check of the auto-registry
type HealthCheck struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Uptime    string           `json:"uptime"`
	Metrics   *RegistryMetrics `json:"metrics,omitempty"`
	Errors    []string         `json:"errors,omitempty"`
}

// GetHealthCheck returns the current health status
func (pr *ProductionAutoRegistry) GetHealthCheck() *HealthCheck {
	health := &HealthCheck{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(pr.startTime).String(),
		Errors:    make([]string, 0),
	}

	if pr.config.EnableHealthChecks {
		health.Metrics = pr.GetMetrics()

		// Check for health issues
		if health.Metrics.FailedRegistrations > 0 {
			failureRate := float64(health.Metrics.FailedRegistrations) / float64(health.Metrics.TotalRegistrations)
			if failureRate > 0.1 { // More than 10% failure rate
				health.Status = "degraded"
				health.Errors = append(health.Errors, fmt.Sprintf("High failure rate: %.2f%%", failureRate*100))
			}
		}

		// Check if average registration time is too high
		if health.Metrics.AverageRegistrationTime > time.Second {
			health.Status = "degraded"
			health.Errors = append(health.Errors, fmt.Sprintf("Slow registration time: %v", health.Metrics.AverageRegistrationTime))
		}
	}

	return health
}

// IsHealthy returns true if the registry is healthy
func (pr *ProductionAutoRegistry) IsHealthy() bool {
	return pr.GetHealthCheck().Status == "healthy"
}

// LogSummary logs a summary of registry status
func (pr *ProductionAutoRegistry) LogSummary() {
	if pr.config.LogLevel < LogLevelInfo {
		return
	}

	metrics := pr.GetMetrics()
	health := pr.GetHealthCheck()

	pr.AutoRegistry.logger.Printf("📊 Auto-Registry Summary:")
	pr.AutoRegistry.logger.Printf("   Status: %s", health.Status)
	pr.AutoRegistry.logger.Printf("   Uptime: %s", health.Uptime)
	pr.AutoRegistry.logger.Printf("   Total Registrations: %d", metrics.TotalRegistrations)
	pr.AutoRegistry.logger.Printf("   Success Rate: %.2f%%",
		float64(metrics.SuccessfulRegistrations)/float64(metrics.TotalRegistrations)*100)
	pr.AutoRegistry.logger.Printf("   Average Registration Time: %v", metrics.AverageRegistrationTime)

	for handlerType, count := range metrics.HandlerTypes {
		if count > 0 {
			pr.AutoRegistry.logger.Printf("   %s handlers: %d", handlerType, count)
		}
	}

	if len(health.Errors) > 0 {
		pr.AutoRegistry.logger.Printf("   Issues:")
		for _, err := range health.Errors {
			pr.AutoRegistry.logger.Printf("     - %s", err)
		}
	}
}

// ValidateConfiguration validates the production configuration
func ValidateConfiguration(config *AutoRegistryConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if config.MaxRegistrationTime < 0 {
		return fmt.Errorf("MaxRegistrationTime cannot be negative")
	}

	if config.EnableRetry && config.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries cannot be negative when retry is enabled")
	}

	if config.EnableRetry && config.RetryDelay < 0 {
		return fmt.Errorf("RetryDelay cannot be negative when retry is enabled")
	}

	if config.LogLevel < LogLevelSilent || config.LogLevel > LogLevelDebug {
		return fmt.Errorf("invalid LogLevel: %d", config.LogLevel)
	}

	return nil
}

// Production-ready clean API functions

// NewProductionAutoRegistryWithDefaults creates a production auto-registry with default config
func NewProductionAutoRegistryWithDefaults(manager *Manager) *ProductionAutoRegistry {
	return NewProductionAutoRegistry(manager, DefaultProductionConfig())
}

// AutoRegisterProductionHandlers provides a clean API for production handler registration
func AutoRegisterProductionHandlers(config *AutoRegistryConfig, handlers ...any) *RegistrationResult {
	prodRegistry := NewProductionAutoRegistry(GetManager(), config)
	return prodRegistry.RegisterHandlerInstancesWithConfig(handlers...)
}

// AutoRegisterWithHealthCheck registers handlers and returns health status
func AutoRegisterWithHealthCheck(handlers ...any) (*RegistrationResult, *HealthCheck) {
	prodRegistry := NewProductionAutoRegistryWithDefaults(GetManager())
	result := prodRegistry.RegisterHandlerInstancesWithConfig(handlers...)
	health := prodRegistry.GetHealthCheck()
	return result, health
}
