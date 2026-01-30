package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type DiscrepancyRepository interface {
	StoreDiscrepancy(discrepancy models.BalanceDiscrepancy) error
	GetUnresolvedDiscrepancies(limit int) ([]models.BalanceDiscrepancy, error)
	GetDiscrepanciesByAddress(address string, limit int) ([]models.BalanceDiscrepancy, error)
	GetDiscrepanciesByBlockNumber(blockNumber int) ([]models.BalanceDiscrepancy, error)
	MarkAsResolved(id int64, resolutionNote string) error
	GetDiscrepanciesCount() (int, error)
}

type discrepancyRepository struct {
	db *sqlx.DB
}

func NewDiscrepancyRepository(db *sqlx.DB) DiscrepancyRepository {
	return &discrepancyRepository{db: db}
}

// GetDiscrepanciesByAddress implements [DiscrepancyRepository].
func (d *discrepancyRepository) GetDiscrepanciesByAddress(address string, limit int) ([]models.BalanceDiscrepancy, error) {
	query := `
		SELECT id, address, block_number, expected_balance, actual_balance, difference, resolved, resolution_note, timestamp
		FROM balance_discrepancy
		WHERE address = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	var discrepancies []models.BalanceDiscrepancy
	err := d.db.Select(&discrepancies, query, address, limit)
	if err != nil {
		return nil, fmt.Errorf("get discrepancies by address: %w", err)
	}

	return discrepancies, nil
}

// GetDiscrepanciesByBlockNumber implements [DiscrepancyRepository].
func (d *discrepancyRepository) GetDiscrepanciesByBlockNumber(blockNumber int) ([]models.BalanceDiscrepancy, error) {
	query := `
		SELECT id, address, block_number, expected_balance, actual_balance, difference, resolved, resolution_note, timestamp
		FROM balance_discrepancy
		WHERE block_number = ?
		ORDER BY timestamp DESC
	`

	var discrepancies []models.BalanceDiscrepancy
	err := d.db.Select(&discrepancies, query, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("get discrepancies by block number: %w", err)
	}

	return discrepancies, nil
}

// GetDiscrepanciesCount implements [DiscrepancyRepository].
func (d *discrepancyRepository) GetDiscrepanciesCount() (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM balance_discrepancy`
	err := d.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("get discrepancies count: %w", err)
	}

	return count, nil
}

// GetUnresolvedDiscrepancies implements [DiscrepancyRepository].
func (d *discrepancyRepository) GetUnresolvedDiscrepancies(limit int) ([]models.BalanceDiscrepancy, error) {
	var discrepancies []models.BalanceDiscrepancy

	query := `
		SELECT id, address, block_number, expected_balance, actual_balance, difference, resolved, resolution_note, timestamp
		FROM balance_discrepancy
		WHERE resolved = FALSE
		ORDER BY timestamp DESC
		LIMIT ?
	`

	err := d.db.Select(&discrepancies, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get unresolved discrepancies: %w", err)
	}

	return discrepancies, nil
}

// MarkAsResolved implements [DiscrepancyRepository].
func (d *discrepancyRepository) MarkAsResolved(id int64, resolutionNote string) error {
	query := `
		UPDATE balance_discrepancy
		SET resolved = TRUE, resolution_note = ?
		WHERE id = ?
	`
	_, err := d.db.Exec(query, resolutionNote, id)
	if err != nil {
		return fmt.Errorf("mark as resolved: %w", err)
	}
	return nil
}

// StoreDiscrepancy implements [DiscrepancyRepository].
func (d *discrepancyRepository) StoreDiscrepancy(discrepancy models.BalanceDiscrepancy) error {
	query := `
		INSERT INTO balance_discrepancy
		(address, block_number, expected_balance, actual_balance, difference, timestamp)
		VALUES
		(?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(
		query,
		discrepancy.Address,
		discrepancy.BlockNumber,
		discrepancy.ExpectedBalance,
		discrepancy.ActualBalance,
		discrepancy.Difference,
		discrepancy.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("store discrepancy: %w", err)
	}

	return nil
}
