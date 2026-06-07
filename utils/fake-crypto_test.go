package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeBlock membangun block dengan hash yang sudah benar-benar dihitung
// via RecalculateBlockHash (bukan MineBlock) sehingga test deterministik.
// CATATAN: RecalculateBlockHash tidak menghasilkan hash dengan leading
// zeros, jadi helper ini default ke difficulty=0. Test yang butuh
// difficulty > 0 harus men-set CurrentHash secara eksplisit setelah
// memanggil helper.
func makeBlock(prevHash string, blockNum int, difficulty int, timestamp int64, txs []models.Transaction) models.Block {
	b := models.Block{
		BlockNumber:  blockNum,
		PreviousHash: prevHash,
		Timestamp:    timestamp,
		Difficulty:   difficulty,
		Transactions: txs,
	}
	b.CurrentHash = RecalculateBlockHash(b)
	return b
}

func TestCheckBlockchainIntegrity_HappyPath(t *testing.T) {
	// Bangun 2-block chain. Setelah fix (MineBlock pakai timestamp yang
	// sama dengan block.Timestamp, dan step 5 dihapus), chain apapun
	// yang konsisten dengan makeBlock harus lulus.
	baseTime := time.Now().Unix()
	genesis := makeBlock("", 1, 0, baseTime, nil)
	block2 := makeBlock(genesis.CurrentHash, 2, 0, baseTime+10, nil)

	err := CheckBlockchainIntegrity([]models.Block{genesis, block2})
	assert.NoError(t, err)
}

func TestCheckBlockchainIntegrity_ThreeBlocks(t *testing.T) {
	// 3-block chain juga harus lulus (dulu gagal di step 5 CalculateBlockHash
	// yang formatnya beda). Setelah fix, RecalculateBlockHash adalah satu-
	// satunya sumber kebenaran untuk validasi hash.
	baseTime := time.Now().Unix()
	genesis := makeBlock("", 1, 0, baseTime, nil)
	block2 := makeBlock(genesis.CurrentHash, 2, 0, baseTime+10, nil)
	block3 := makeBlock(block2.CurrentHash, 3, 0, baseTime+20, nil)

	err := CheckBlockchainIntegrity([]models.Block{genesis, block2, block3})
	assert.NoError(t, err)
}

func TestCheckBlockchainIntegrity_BrokenLink(t *testing.T) {
	// Putuskan previousHash link di block 2.
	baseTime := time.Now().Unix()
	genesis := makeBlock("", 1, 0, baseTime, nil)
	block2 := makeBlock(genesis.CurrentHash, 2, 0, baseTime+10, nil)
	block3 := makeBlock("WRONG_HASH", 3, 0, baseTime+20, nil)

	err := CheckBlockchainIntegrity([]models.Block{genesis, block2, block3})
	require.Error(t, err)
	// step 1 (previous hash link) harus trigger sebelum step 5 (calculate hash)
	assert.Contains(t, err.Error(), "previous hash mismatch")
}

func TestCheckBlockchainIntegrity_InvalidPoW(t *testing.T) {
	// Block dengan difficulty 4 tapi hash tidak punya 4 leading zero.
	baseTime := time.Now().Unix()
	genesis := makeBlock("", 1, 0, baseTime, nil)
	block2 := makeBlock(genesis.CurrentHash, 2, 0, baseTime+10, nil)
	// Naikkan difficulty ke 4; RecalculateBlockHash belum tentu punya 4 leading zeros
	block2.Difficulty = 4
	block2.CurrentHash = "ffff" + strings.Repeat("a", 60) // tidak ada leading zero

	err := CheckBlockchainIntegrity([]models.Block{genesis, block2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid proof of work")
}

func TestCheckBlockchainIntegrity_RecalcHashMismatch(t *testing.T) {
	// Sabotase: CurrentHash = RecalculateBlockHash(block2) TAPI difficulty dinaikkan
	// sehingga leading zero tidak cukup -> PoW gagal dulu.
	// Tes ini membuktikan PoW check terjadi sebelum RecalcHash check.
	baseTime := time.Now().Unix()
	genesis := makeBlock("", 1, 0, baseTime, nil)
	block2 := makeBlock(genesis.CurrentHash, 2, 0, baseTime+10, nil)
	// Naikkan difficulty TANPA mengubah hash (hash tidak punya leading zero cukup)
	block2.Difficulty = 10

	err := CheckBlockchainIntegrity([]models.Block{genesis, block2})
	require.Error(t, err)
	// PoW check akan trigger duluan karena RecalculateBlockHash output
	// tidak punya 10 leading zeros
	assert.Contains(t, err.Error(), "invalid proof of work")
}

func TestCheckBlockchainIntegrity_TimestampNotMonotonic(t *testing.T) {
	// Block 2 timestamp mundur dari block 1.
	baseTime := time.Now().Unix()
	genesis := makeBlock("", 1, 0, baseTime, nil)
	block2 := makeBlock(genesis.CurrentHash, 2, 0, baseTime-1, nil) // mundur

	err := CheckBlockchainIntegrity([]models.Block{genesis, block2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timestamp not greater")
}

func TestGenerateFakeKey_Unique(t *testing.T) {
	// Generate 10 key, harus semua unik (probabilitas tabrakan astronomis kecil).
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		priv, pub := GenerateFakeKey()
		assert.Len(t, priv, 64, "private key hex harus 64 char (32 byte)")
		assert.Len(t, pub, 64, "public key hex harus 64 char (32 byte)")
		assert.False(t, seen[priv], "private key duplikat")
		seen[priv] = true
	}
}

func TestGenerateAddressFromPublicKey_Deterministic(t *testing.T) {
	addr1 := GenerateAddressFromPublicKey("pubkey123")
	addr2 := GenerateAddressFromPublicKey("pubkey123")
	assert.Equal(t, addr1, addr2)
	assert.True(t, strings.HasPrefix(addr1, "FAKE-"), "address harus prefix 'FAKE-'")
	assert.Len(t, addr1, len("FAKE-")+32, "address harus 32 char hex setelah prefix")
}

func TestGenerateAddressFromPublicKey_DifferentInput(t *testing.T) {
	a := GenerateAddressFromPublicKey("pubkey1")
	b := GenerateAddressFromPublicKey("pubkey2")
	assert.NotEqual(t, a, b)
}

func TestSignFake_Deterministic(t *testing.T) {
	// Payload sama -> signature sama.
	s1 := SignFake("privkey1", "0xAddr", 100.50)
	s2 := SignFake("privkey1", "0xAddr", 100.50)
	assert.Equal(t, s1, s2)
	assert.Len(t, s1, 64)
}

func TestSignFake_AmountMatters(t *testing.T) {
	// Mutasi amount harus ubah signature.
	s1 := SignFake("priv", "0xAddr", 100)
	s2 := SignFake("priv", "0xAddr", 100.01)
	assert.NotEqual(t, s1, s2, "amount berbeda harus signature berbeda")
}
