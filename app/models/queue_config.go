package models

const (
	// Exchanges
	TransactionExchange  = "transactions"
	BlockExchange        = "blocks"
	MarketExchange       = "market"
	NotificationExchange = "notifications"

	// Queues
	TransactionPendingQueue   = "transaction.pending"
	TransactionConfirmedQueue = "transaction.confirmed"
	BlockGenerationQueue      = "block.generation"
	BlockMinedQueue           = "block.mined"
	MarketPricingQueue        = "market.pricing"
	MarketVolumeQueue         = "market.volume"
	NotificationQueue         = "notification.general"

	// Routing Keys
	TransactionSubmittedKey = "transaction.submitted"
	TransactionConfirmedKey = "transaction.confirmed"
	BlockGenerateKey        = "block.generate"
	BlockMinedKey           = "block.mined"
	MarketPriceUpdateKey    = "market.price.update"
	MarketVolumeUpdateKey   = "market.volume.update"
)

type QueueConfig struct {
	Queue        string
	Exchange     string
	RoutingKey   string
	Durable      bool
	AutoDelete   bool
	ExchangeType string
}

var QueueConfigs = []QueueConfig{
	// Transaction Queues
	{
		Queue:        TransactionPendingQueue,
		Exchange:     TransactionExchange,
		RoutingKey:   TransactionSubmittedKey,
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},
	{
		Queue:        TransactionConfirmedQueue,
		Exchange:     TransactionExchange,
		RoutingKey:   TransactionConfirmedKey,
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},

	// Block Queues
	{
		Queue:        BlockGenerationQueue,
		Exchange:     BlockExchange,
		RoutingKey:   BlockGenerateKey,
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},
	{
		Queue:        BlockMinedQueue,
		Exchange:     BlockExchange,
		RoutingKey:   BlockMinedKey,
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},

	// Market Queues
	{
		Queue:        MarketPricingQueue,
		Exchange:     MarketExchange,
		RoutingKey:   MarketPriceUpdateKey,
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},
	{
		Queue:        MarketVolumeQueue,
		Exchange:     MarketExchange,
		RoutingKey:   MarketVolumeUpdateKey,
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},

	// Notification Queues
	{
		Queue:        NotificationQueue,
		Exchange:     NotificationExchange,
		RoutingKey:   "notification.#",
		Durable:      true,
		AutoDelete:   false,
		ExchangeType: "topic",
	},
}
