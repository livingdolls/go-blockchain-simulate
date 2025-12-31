package port

import "context"

type MessageBroker interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channel string, callback func([]byte) error) error
}
