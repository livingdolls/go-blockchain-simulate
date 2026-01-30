package rabbitmq

const (
	// Exchange Names
	TransactionExchange = "transactions"
	BlockExchange       = "blocks"
	MarketExchange      = "market"
	LedgerExchange      = "ledger"
	RewardExchange      = "rewards"

	// Queue Names
	TransactionPendingQueue   = "transaction.pending"
	TransactionConfirmedQueue = "transaction.confirmed"
	BlockGenerationQueue      = "block.generation"
	BlockMinedQueue           = "block.mined"
	LedgerEntriesQueue        = "ledger.entries"
	MarketPricingQueue        = "market.pricing"
	MarketVolumeQueue         = "market.volume.updates"
	RewardCalculationQueue    = "reward.calculation"
	RewardDistributionQueue   = "reward.distribution"
	LedgerPresistenceQueue    = "ledger.persistence"
	LedgerAuditQueue          = "ledger.audit"
	LedgerReconcileQueue      = "ledger.reconcile"

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
