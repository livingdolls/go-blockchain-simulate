package services

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

type MarketEngineService interface {
	GetState() (models.MarketEngine, error)
	ApplyBlockPricingWithTx(tx *sqlx.Tx, blockID int64, buyVolume, sellVolume float64, txCount int) (models.MarketEngine, error)
}

type marketEngineService struct {
	repo     repository.MarketRepository
	base     float64
	slope    float64
	minPrice float64
}

func NewMarketEngineService(repo repository.MarketRepository) MarketEngineService {
	return &marketEngineService{
		repo:     repo,
		base:     1000.0,
		slope:    0.0001,
		minPrice: 1.0,
	}
}

// ApplyBlockPricingWithTx implements MarketEngineService.
func (m *marketEngineService) ApplyBlockPricingWithTx(tx *sqlx.Tx, blockID int64, buyVolume float64, sellVolume float64, txCount int) (models.MarketEngine, error) {
	state, err := m.repo.GetStateForUpdateWithTx(tx)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return models.MarketEngine{}, err
		}

		state = models.MarketEngine{
			ID:        1,
			Price:     100.0,
			Liquidity: 0,
			LastBlock: 0,
		}
	}

	deltaVolume := buyVolume - sellVolume
	newLiquidity := state.Liquidity + deltaVolume

	if newLiquidity < 0 {
		newLiquidity = 0
	}

	newPrice := state.Price + m.slope*deltaVolume

	if newPrice < m.minPrice {
		newPrice = m.minPrice
	}

	state.Price = newPrice
	state.Liquidity = newLiquidity
	state.LastBlock = blockID

	if err := m.repo.UpdateStateWithTx(tx, state); err != nil {
		return models.MarketEngine{}, err
	}

	if _, err := m.repo.InsertTickWithTx(tx, models.MarketTick{
		BlockID:    blockID,
		Price:      state.Price,
		BuyVolume:  buyVolume,
		SellVolume: sellVolume,
		TxCount:    txCount,
	}); err != nil {
		return models.MarketEngine{}, err
	}

	return state, nil
}

// GetState implements MarketEngineService.
func (m *marketEngineService) GetState() (models.MarketEngine, error) {
	return m.repo.GetState()
}
