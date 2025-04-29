package cqrs

type ICqrsConfiguration interface {
	UseLoggingDecorator() ICqrsConfiguration
	UseMetricsDecorator() ICqrsConfiguration
	UseErrorHandlerDecorator() ICqrsConfiguration
}

type Configuration struct {
	enableErrorHandlerDecorator bool
	enableLoggingDecorator      bool
	enableMetricsDecorator      bool
}

func (c *Configuration) UseLoggingDecorator() ICqrsConfiguration {
	c.enableLoggingDecorator = true
	return c
}

func (c *Configuration) UseMetricsDecorator() ICqrsConfiguration {
	c.enableMetricsDecorator = true
	return c
}

func (c *Configuration) UseErrorHandlerDecorator() ICqrsConfiguration {
	c.enableErrorHandlerDecorator = true
	return c
}
