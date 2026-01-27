package rabbitmq

const (
	// Exchange Names
	TransactionExchange = "transactions"
	BlockExchange       = "blocks"
	MarketExchange      = "market"

	// Queue Names
	TransactionPendingQueue   = "transaction.pending"
	TransactionConfirmedQueue = "transaction.confirmed"
	BlockGenerationQueue      = "block.generation"
	BlockMinedQueue           = "block.mined"
	MarketPricingQueue        = "market.pricing"

	// Routing Keys
	TransactionSubmittedKey = "transaction.submitted"
	TransactionConfirmedKey = "transaction.confirmed"
	BlockGenerateKey        = "block.generate"
	BlockMinedKey           = "block.mined"
	MarketPricingKey        = "market.pricing"
)
