package cqrs

import "github.com/kmdeveloping/go-cqrs/core/registry"

type CqrsConfiguration struct {
	Registry               *registry.Registry
	EnableLoggingDecorator bool
}
