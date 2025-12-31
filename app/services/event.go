package services

import (
	"context"
	"log"

	"github.com/livingdolls/go-blockchain-simulate/app/port"
)

type EventService struct {
	broker port.MessageBroker
}

func NewEventService(broker port.MessageBroker) port.MessageBroker {
	return &EventService{
		broker: broker,
	}
}

// Publish implements [port.MessageBroker].
func (e *EventService) Publish(ctx context.Context, channel string, payload []byte) error {
	log.Println("event publish to = ", channel)
	return e.broker.Publish(ctx, channel, payload)
}

// Subscribe implements [port.MessageBroker].
func (e *EventService) Subscribe(ctx context.Context, channel string, callback func([]byte) error) error {
	return e.broker.Subscribe(
		ctx, channel, func(msg []byte) error {
			log.Println("event received :", string(msg))
			return callback(msg)
		})
}
