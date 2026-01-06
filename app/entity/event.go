package entity

type MessageType string

const (
	EventMarketUpdate      MessageType = "market.update"
	EventTypeBlockMined    MessageType = "block.mined"
	EventTypeSubscribe     MessageType = "subscribe"
	EventTypeUnsubscribe   MessageType = "unsubscribe"
	EventTransactionUpdate MessageType = "transaction.update"
	EventBalanceUpdate     MessageType = "balance.update"
)
