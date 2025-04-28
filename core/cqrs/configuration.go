package cqrs

import (
	"github.com/kmdeveloping/go-cqrs/core/registry"
)

type CqrsConfiguration struct {
	Handlers                    registry.Registry
	enableErrorHandlerDecorator bool
	enableLoggingDecorator      bool
	enableMetricsDecorator      bool
}
