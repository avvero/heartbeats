package main

type Event struct {
	service     *Service
	description string
}

type EventConsumer interface {
	consume(event *Event)
}

type EventConsumerStub struct {
}

func (it *EventConsumerStub) consume(event *Event) {
	//do nothing
}
