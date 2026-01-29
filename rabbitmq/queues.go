package rabbitmq

const (
	// Exchange Names
	TransactionExchange = "transactions"
	BlockExchange       = "blocks"
	MarketExchange      = "market"
	LedgerExchange      = "ledger"
	LedgerEntriesQueue  = "ledger.entries"
	RewardExchange      = "rewards"

	// Queue Names
	TransactionPendingQueue   = "transaction.pending"
	TransactionConfirmedQueue = "transaction.confirmed"
	BlockGenerationQueue      = "block.generation"
	BlockMinedQueue           = "block.mined"
	MarketPricingQueue        = "market.pricing"
	MarketVolumeQueue         = "market.volume.updates"
	RewardCalculationQueue    = "reward.calculation"
	RewardDistributionQueue   = "reward.distribution"

	// Routing Keys
	TransactionSubmittedKey = "transaction.submitted"
	TransactionConfirmedKey = "transaction.confirmed"
	BlockGenerateKey        = "block.generate"
	BlockMinedKey           = "block.mined"
	MarketPricingKey        = "market.pricing"
	MarketVolumeUpdateKey   = "market.volume.update"
	LedgerBatchKey          = "ledger.batch"
	LedgerEntryKey          = "ledger.entry"
	RewardCalculationKey    = "reward.calculation"
	RewardDistributionKey   = "reward.distribution"
)
