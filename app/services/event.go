package services

import (
	"context"

	"github.com/livingdolls/go-blockchain-simulate/app/port"
	"github.com/livingdolls/go-blockchain-simulate/logger"
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
	logger.LogInfo("event published : channel=" + channel + " payload=" + string(payload))
	return e.broker.Publish(ctx, channel, payload)
}

// Subscribe implements [port.MessageBroker].
func (e *EventService) Subscribe(ctx context.Context, channel string, callback func([]byte) error) error {
	return e.broker.Subscribe(
		ctx, channel, func(msg []byte) error {
			logger.LogInfo("event received : channel=" + channel + " payload=" + string(msg))
			return callback(msg)
		})
}
