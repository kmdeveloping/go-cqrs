package decorators

import (
	"context"
	"log"
	"time"
)

// MetricsConfig provides configuration for the metrics decorator
type MetricsConfig struct {
	Enabled            bool
	LogAllExecutions   bool
	SlowThreshold      time.Duration // Only log executions slower than this
	EnableDetailedLogs bool          // Whether to log detailed timing info
}

// OPTIMIZATION: Improved metrics decorator with better performance and configurability
func MetricsDecorator() HandlerDecorator {
	return MetricsDecoratorWithConfig(MetricsConfig{
		Enabled:            true,
		LogAllExecutions:   false,
		SlowThreshold:      100 * time.Millisecond,
		EnableDetailedLogs: false,
	})
}

// OPTIMIZATION: Configurable metrics decorator for better performance control
func MetricsDecoratorWithConfig(config MetricsConfig) HandlerDecorator {
	return func(next IHandlerDecorator) IHandlerDecorator {
		return HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
			// OPTIMIZATION: Early exit if metrics are disabled
			if !config.Enabled {
				return next.Handle(ctx, message)
			}

			start := time.Now()

			// OPTIMIZATION: Only log start time if detailed logging is enabled
			if config.EnableDetailedLogs && config.LogAllExecutions {
				log.Printf("Execution started @ %s", start.Format(time.RFC3339Nano))
			}

			result, err := next.Handle(ctx, message)

			duration := time.Since(start)

			// OPTIMIZATION: Only log if it exceeds threshold or if all executions should be logged
			shouldLog := config.LogAllExecutions || (config.SlowThreshold > 0 && duration > config.SlowThreshold)

			if shouldLog {
				if config.EnableDetailedLogs {
					stop := time.Now()
					if err != nil {
						log.Printf("Execution failed @ %s\t duration: %s\t type: %T\t error: %v",
							stop.Format(time.RFC3339Nano), duration, message, err)
					} else {
						log.Printf("Execution completed @ %s\t duration: %s\t type: %T",
							stop.Format(time.RFC3339Nano), duration, message)
					}
				} else {
					// OPTIMIZATION: Simple logging for better performance
					if err != nil {
						log.Printf("Handler failed in %s for %T: %v", duration, message, err)
					} else {
						log.Printf("Handler completed in %s for %T", duration, message)
					}
				}
			}

			return result, err
		})
	}
}

// OPTIMIZATION: New function for high-performance scenarios with minimal logging
func MetricsDecoratorMinimal() HandlerDecorator {
	return MetricsDecoratorWithConfig(MetricsConfig{
		Enabled:            true,
		LogAllExecutions:   false,
		SlowThreshold:      1 * time.Second, // Only log very slow operations
		EnableDetailedLogs: false,
	})
}
