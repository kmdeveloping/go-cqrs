package event

type IEvent interface{}

type BaseEventHandler[TEvent IEvent] struct {
	Publish func(*TEvent) error
}
