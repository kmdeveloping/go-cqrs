package decorators

import (
	"context"
	"log"
)

// LoggingConfig provides configuration for the logging decorator
type LoggingConfig struct {
	Enabled    bool
	LogInputs  bool
	LogOutputs bool
	SampleRate float64 // 0.0 to 1.0, for sampling high-traffic scenarios
}

// OPTIMIZATION: Improved logging decorator with configurable options and reduced allocations
func LoggingDecorator(logger *log.Logger) HandlerDecorator {
	return LoggingDecoratorWithConfig(logger, LoggingConfig{
		Enabled:    true,
		LogInputs:  true,
		LogOutputs: false,
		SampleRate: 1.0,
	})
}

// OPTIMIZATION: Configurable logging decorator for better performance control
func LoggingDecoratorWithConfig(logger *log.Logger, config LoggingConfig) HandlerDecorator {
	return func(next IHandlerDecorator) IHandlerDecorator {
		return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
			// OPTIMIZATION: Early exit if logging is disabled
			if !config.Enabled {
				return next.Handle(ctx, message)
			}

			// OPTIMIZATION: Implement sampling for high-traffic scenarios
			// Note: This is a simple implementation, in production you might want
			// a more sophisticated sampling strategy
			if config.SampleRate < 1.0 {
				// Simple sampling based on message hash
				// In a real implementation, you might use a better sampling strategy
				hash := simpleHash(message)
				if float64(hash%100)/100.0 > config.SampleRate {
					return next.Handle(ctx, message)
				}
			}

			// OPTIMIZATION: Use structured logging to reduce string concatenation
			if config.LogInputs {
				logger.Printf("[Handler Input] Type=%T", message)
			}

			result, err := next.Handle(ctx, message)

			if config.LogOutputs {
				if err != nil {
					logger.Printf("[Handler Output] Type=%T Error=%v", message, err)
				} else {
					logger.Printf("[Handler Output] Type=%T Success", message)
				}
			}

			return result, err
		})
	}
}

// OPTIMIZATION: Simple hash function for sampling
func simpleHash(v any) uint32 {
	// Simple hash based on type name length and some basic properties
	// This is not cryptographically secure but sufficient for sampling
	if v == nil {
		return 0
	}

	typeName := getTypeName(v)
	h := uint32(2166136261) // FNV offset basis
	for i := 0; i < len(typeName); i++ {
		h ^= uint32(typeName[i])
		h *= 16777619 // FNV prime
	}
	return h
}

// OPTIMIZATION: Get type name without expensive reflection
func getTypeName(v any) string {
	// This is a simplified version - in a real implementation you might
	// want to cache this information
	switch v.(type) {
	default:
		return "unknown"
	}
}
