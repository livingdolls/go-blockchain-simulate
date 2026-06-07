package services

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBlockRepo adalah stub BlockRepository yang hanya GetLastBlock aktif.
// Method lain akan panic jika dipanggil (tujuannya agar test jujur).
type fakeBlockRepo struct {
	lastBlock models.Block
	err       error
}

func (f *fakeBlockRepo) GetLastBlock() (models.Block, error) {
	return f.lastBlock, f.err
}

// method lain tidak relevan untuk test reward, isi stub kosong
func (f *fakeBlockRepo) BeginTx() (*sqlx.Tx, error)                       { return nil, nil }
func (f *fakeBlockRepo) CreateWithTx(*sqlx.Tx, models.Block) (int64, error) { return 0, nil }
func (f *fakeBlockRepo) GetLastBlockForUpdateWithTx(*sqlx.Tx) (models.Block, error) {
	return models.Block{}, nil
}
func (f *fakeBlockRepo) InsertBlockTransactionWithTx(*sqlx.Tx, int64, int64) error { return nil }
func (f *fakeBlockRepo) BulkInsertBlockTransactionsWithTx(*sqlx.Tx, int64, []int64) error {
	return nil
}
func (f *fakeBlockRepo) GetBlocks(int, int) ([]models.Block, error)  { return nil, nil }
func (f *fakeBlockRepo) GetBlockByID(int64) (models.Block, error)    { return models.Block{}, nil }
func (f *fakeBlockRepo) GetAllBlocks() ([]models.Block, error)       { return nil, nil }
func (f *fakeBlockRepo) GetTotalFeesInBlock(int64) (float64, error)  { return 0, nil }
func (f *fakeBlockRepo) GetBlockByBlockNumber(int64) (models.Block, error) {
	return models.Block{}, nil
}
func (f *fakeBlockRepo) GetTransactionsByBlockNumber(context.Context, int64) ([]models.Transaction, error) {
	return nil, nil
}
func (f *fakeBlockRepo) SearchByHash(context.Context, string) ([]models.Block, error) {
	return nil, nil
}
func (f *fakeBlockRepo) GetBlocksInRange(context.Context, int64, int64) ([]models.Block, error) {
	return nil, nil
}
func (f *fakeBlockRepo) GetBlockStats(context.Context) (models.BlockStats, error) {
	return models.BlockStats{}, nil
}
func (f *fakeBlockRepo) GetLatestBlockInfo(context.Context) (models.Block, error) {
	return models.Block{}, nil
}
func (f *fakeBlockRepo) GetBlockCountLastHour(context.Context) (int64, error) { return 0, nil }
func (f *fakeBlockRepo) SearchByMinerAddress(context.Context, string, int, int) ([]models.Block, error) {
	return nil, nil
}

var _ repository.BlockRepository = (*fakeBlockRepo)(nil)

func TestRewardInfo_EraZero(t *testing.T) {
	// Block 50 (era 0): reward 50, next reward 50, current supply 2500
	repo := &fakeBlockRepo{lastBlock: models.Block{BlockNumber: 50}}
	svc := NewRewardHandler(repo)

	info, err := svc.RewardInfo()
	require.NoError(t, err)
	assert.Equal(t, int64(50), info.CurrentBlockNumber)
	assert.Equal(t, 50.0, info.CurrentReward)
	assert.Equal(t, 50.0, info.NextReward) // block 51 juga era 0
	assert.Equal(t, int64(100), info.NextHalvingBlock)
	assert.Equal(t, int64(50), info.BlocksUntilHalving)
	assert.Equal(t, 2500.0, info.CurrentSupply) // 50 * 50
	assert.InDelta(t, 25.0, info.SupplyPercentage, 0.01) // 2500/10000
}

func TestRewardInfo_EraTransition(t *testing.T) {
	// Block 99 (last of era 0): next reward (block 100) HARUS 25
	repo := &fakeBlockRepo{lastBlock: models.Block{BlockNumber: 99}}
	svc := NewRewardHandler(repo)

	info, err := svc.RewardInfo()
	require.NoError(t, err)
	assert.Equal(t, 50.0, info.CurrentReward)
	assert.Equal(t, 25.0, info.NextReward, "block 100 mulai era 1, reward jadi 25")
	assert.Equal(t, int64(1), info.BlocksUntilHalving)
}

func TestRewardInfo_RepoError(t *testing.T) {
	repo := &fakeBlockRepo{err: errors.New("db down")}
	svc := NewRewardHandler(repo)

	_, err := svc.RewardInfo()
	require.Error(t, err)
	assert.Equal(t, "db down", err.Error())
}

func TestGetRewardSchedule_HalvingMarker(t *testing.T) {
	// Block terakhir = 95. Schedule 10 block (96-105). Halving di 100.
	repo := &fakeBlockRepo{lastBlock: models.Block{BlockNumber: 95}}
	svc := NewRewardHandler(repo)

	sched, err := svc.GetRewardSchedule(10)
	require.NoError(t, err)
	assert.Equal(t, int64(95), sched.CurrentBlockNumber)
	require.Len(t, sched.Schedule, 10)

	// Block 96-99: era 0 (reward 50)
	for i := 0; i < 4; i++ {
		assert.Equal(t, 50.0, sched.Schedule[i].Reward, "block %d harus era 0", 96+i)
		assert.False(t, sched.Schedule[i].IsHalving)
	}
	// Block 100: halving
	assert.Equal(t, 25.0, sched.Schedule[4].Reward)
	assert.True(t, sched.Schedule[4].IsHalving, "block 100 harus ditandai isHalving")
	// Block 101-105: era 1 (reward 25)
	for i := 5; i < 10; i++ {
		assert.Equal(t, 25.0, sched.Schedule[i].Reward)
		assert.False(t, sched.Schedule[i].IsHalving)
	}
}

func TestGetRewardSchedule_RepoError(t *testing.T) {
	repo := &fakeBlockRepo{err: errors.New("db error")}
	svc := NewRewardHandler(repo)

	_, err := svc.GetRewardSchedule(5)
	require.Error(t, err)
}
