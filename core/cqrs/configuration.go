package cqrs

type ICqrsConfiguration interface {
	UseLoggingDecorator() ICqrsConfiguration
	UseMetricsDecorator() ICqrsConfiguration
	UseErrorHandlerDecorator() ICqrsConfiguration
}

type CqrsConfiguration struct {
	enableErrorHandlerDecorator bool
	enableLoggingDecorator      bool
	enableMetricsDecorator      bool
}

func (c *CqrsConfiguration) UseLoggingDecorator() ICqrsConfiguration {
	c.enableLoggingDecorator = true
	return c
}

func (c *CqrsConfiguration) UseMetricsDecorator() ICqrsConfiguration {
	c.enableMetricsDecorator = true
	return c
}

func (c *CqrsConfiguration) UseErrorHandlerDecorator() ICqrsConfiguration {
	c.enableErrorHandlerDecorator = true
	return c
}
