package websocket

import "github.com/livingdolls/go-blockchain-simulate/app/entity"

type Message struct {
	Type entity.MessageType `json:"type"`
	Data any                `json:"data,omitempty"`
}

type SubscribeRequest struct {
	Events []entity.MessageType `json:"events"`
}

type SubscribeResponse struct {
	Success bool                 `json:"success"`
	Events  []entity.MessageType `json:"events"`
}
