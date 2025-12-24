package publisher

import (
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	ws "github.com/livingdolls/go-blockchain-simulate/app/websocket"
)

type PublisherWS struct {
	hub *ws.Hub
}

func NewPublisherWS(hub *ws.Hub) *PublisherWS {
	return &PublisherWS{
		hub: hub,
	}
}

func (p *PublisherWS) Publish(eventType entity.MessageType, message interface{}) {
	p.hub.BroadCast(eventType, message)
}
