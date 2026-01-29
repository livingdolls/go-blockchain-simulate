package app

import (
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
)

// getQueueDefinitions returns all queue definitions
func getQueueDefinitions() []models.QueueDef {
	return []models.QueueDef{
		{Name: rabbitmq.TransactionPendingQueue, Durable: true, AutoDelete: false},
		{Name: rabbitmq.TransactionConfirmedQueue, Durable: true, AutoDelete: false},
		{Name: rabbitmq.BlockGenerationQueue, Durable: true, AutoDelete: false},
		{Name: rabbitmq.BlockMinedQueue, Durable: true, AutoDelete: false},
		{Name: rabbitmq.MarketPricingQueue, Durable: true, AutoDelete: false},
		{Name: rabbitmq.MarketVolumeQueue, Durable: true, AutoDelete: false},
		{Name: rabbitmq.LedgerEntriesQueue, Durable: true, AutoDelete: false},
	}
}

// getExchangeDefinitions returns all exchange definitions
func getExchangeDefinitions() []models.ExchangeDef {
	return []models.ExchangeDef{
		{Name: rabbitmq.TransactionExchange, Kind: "topic", Durable: true},
		{Name: rabbitmq.BlockExchange, Kind: "topic", Durable: true},
		{Name: rabbitmq.MarketExchange, Kind: "topic", Durable: true},
		{Name: rabbitmq.LedgerExchange, Kind: "topic", Durable: true},
	}
}

// getBindingDefinitions returns all queue-to-exchange bindings
func getBindingDefinitions() []models.BindDef {
	return []models.BindDef{
		{Queue: rabbitmq.TransactionPendingQueue, Exchange: rabbitmq.TransactionExchange, RoutingKey: rabbitmq.TransactionSubmittedKey},
		{Queue: rabbitmq.TransactionConfirmedQueue, Exchange: rabbitmq.TransactionExchange, RoutingKey: rabbitmq.TransactionConfirmedKey},
		{Queue: rabbitmq.BlockGenerationQueue, Exchange: rabbitmq.BlockExchange, RoutingKey: rabbitmq.BlockGenerateKey},
		{Queue: rabbitmq.BlockMinedQueue, Exchange: rabbitmq.BlockExchange, RoutingKey: rabbitmq.BlockMinedKey},
		{Queue: rabbitmq.MarketPricingQueue, Exchange: rabbitmq.MarketExchange, RoutingKey: rabbitmq.MarketPricingKey},
		{Queue: rabbitmq.MarketVolumeQueue, Exchange: rabbitmq.MarketExchange, RoutingKey: rabbitmq.MarketVolumeUpdateKey},
		{Queue: rabbitmq.LedgerEntriesQueue, Exchange: rabbitmq.LedgerExchange, RoutingKey: rabbitmq.LedgerEntryKey},
	}
}
