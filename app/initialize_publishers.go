package app

import "github.com/livingdolls/go-blockchain-simulate/app/services"

// InitializePublishers menginisialisasi publisher untuk mengirim event
// ke RabbitMQ. Dipakai oleh service layer (Block, Market, Reward).
func (a *AppConfig) InitializePublishers() {
	a.PricingPublisher = services.NewMarketPricingPublisher(a.RMQClient)
	a.LedgerPublisher = services.NewLedgerPublisher(a.RMQClient)
	a.RewardPublisher = services.NewRewardPublisher(a.RMQClient)
}
