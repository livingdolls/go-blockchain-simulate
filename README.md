# Go Blockchain Simulator

A simplified blockchain implementation in Go with REST API for learning cryptocurrency concepts including digital signatures, transaction validation, and block mining.

## 🚀 Features

- **Digital Wallet** - Generate RSA key pairs for wallet creation
- **Digital Signatures** - Sign and verify transactions using RSA encryption
- **Transaction Management** - Create, validate, and track transactions
- **Transaction Fees** - Each transaction includes a fee, paid to the miner
- **Block Rewards** - Miner receives a block reward (coinbase transaction) for each mined block
- **Miner Account** - All transaction fees and block rewards go directly to the miner's wallet
- **Ledger Improvements** - Accurate recording of all balance changes, including coinbase/reward entries (with nullable `tx_id`)
- **Proof of Work Mining** - SHA256-based mining with configurable difficulty (default: 4 leading zeros)
- **Dynamic Difficulty Adjustment** - Automatically adjusts every 10 blocks targeting 10s block time
- **Merkle Tree Verification** - Transaction integrity using Merkle roots for SPV support
- **Block Mining** - Two-phase mining process with read validation + PoW + short write transaction
- **Balance Tracking** - Calculate balances from transaction history via ledger
- **Optimized Database Operations** - Bulk operations achieving 95% performance improvement (<2s transaction time)
- **Blockchain Integrity Validation** - Complete chain verification with PoW and Merkle proof validation
- **REST API** - HTTP endpoints for wallet registration, transactions, block generation, and explorer endpoints
- **Database Persistence** - Store users, transactions, blocks, and ledger entries in MySQL

## 🆕 Recent Improvements

- **Transaction fees and block rewards** are now implemented and paid directly to the miner
- **FEE_POOL logic removed**: all fees go to the miner, matching modern blockchain standards
- **Ledger entries**: coinbase/reward entries use `NULL` for `tx_id` (no foreign key error)
- **Decimal precision**: amounts and fees now use `DECIMAL(32,8)` for accuracy
- **Bug fixes**: transaction repository, balance service, and panic fixes

## �️ Roadmap

1. **Block Explorer API**: Endpoints for querying blocks, transactions, and addresses
2. **Block Explorer Web UI**: Simple web interface for browsing blockchain data
3. Wallet management improvements
4. Analytics dashboard (block/tx stats)
5. Websocket for real-time updates

See the bottom of this file for more details on planned features.

## �📋 Prerequisites

- Go 1.23.0 or higher
- MySQL 8.0 or higher
- Make (optional, for using Makefile commands)

## 🛠️ Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/livingdolls/go-blockchain-simulate.git
   cd go-blockchain-simulate
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Setup database**

   ```sql
   CREATE DATABASE blockchain_db;
   USE blockchain_db;

   -- Users table
   CREATE TABLE users (
       id INT AUTO_INCREMENT PRIMARY KEY,
       address VARCHAR(255) UNIQUE NOT NULL,
       public_key TEXT NOT NULL,
       private_key TEXT NOT NULL,
       balance DOUBLE DEFAULT 1000.0,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );

   -- Transactions table
   CREATE TABLE transactions (
       id BIGINT AUTO_INCREMENT PRIMARY KEY,
       from_address VARCHAR(255) NOT NULL,
       to_address VARCHAR(255) NOT NULL,
       amount DOUBLE NOT NULL,
       signature TEXT NOT NULL,
       status VARCHAR(20) DEFAULT 'PENDING',
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       INDEX idx_status (status)
   );

   -- Ledger table
   CREATE TABLE ledger (
       id BIGINT AUTO_INCREMENT PRIMARY KEY,
       tx_id BIGINT NOT NULL,
       address VARCHAR(255) NOT NULL,
       change_amount DOUBLE NOT NULL,
       balance_after DOUBLE NOT NULL,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       FOREIGN KEY (tx_id) REFERENCES transactions(id)
   );

   -- Blocks table
   CREATE TABLE blocks (
       id BIGINT AUTO_INCREMENT PRIMARY KEY,
       block_number INT NOT NULL,
       timestamp BIGINT NOT NULL,
       previous_hash VARCHAR(64) NOT NULL,
       current_hash VARCHAR(64) NOT NULL,
       nonce BIGINT NOT NULL DEFAULT 0,
       difficulty INT NOT NULL DEFAULT 4,
       merkle_root VARCHAR(64) DEFAULT '',
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       UNIQUE KEY unique_block_number (block_number)
   );

   -- Block-Transaction junction table
   CREATE TABLE block_transactions (
       block_id BIGINT NOT NULL,
       transaction_id BIGINT NOT NULL,
       PRIMARY KEY (block_id, transaction_id),
       FOREIGN KEY (block_id) REFERENCES blocks(id),
       FOREIGN KEY (transaction_id) REFERENCES transactions(id)
   );
   ```

4. **Configure database connection**

   Update database credentials in `database/conn.go`:

   ```go
   dsn := "username:password@tcp(localhost:3306)/blockchain_db?parseTime=true"
   ```

## 🚦 Running the Application

### Using Make:

```bash
make run
```

### Using Go directly:

```bash
go run main.go
```

The API will start on `http://localhost:3010`

## 📡 API Endpoints

### 1. Register Wallet

Create a new wallet with initial balance.

```http
POST /register
Content-Type: application/json

{
    "initial_balance": 1000
}
```

**Response:**

```json
{
  "address": "0xabc123...",
  "public_key": "-----BEGIN PUBLIC KEY-----...",
  "private_key": "-----BEGIN PRIVATE KEY-----...",
  "balance": 1000
}
```

### 2. Send Transaction

Transfer funds between wallets.

```http
POST /send
Content-Type: application/json

{
    "from_address": "0xabc123...",
    "to_address": "0xdef456...",
    "private_key": "-----BEGIN PRIVATE KEY-----...",
    "amount": 50
}
```

**Response:**

```json
{
  "id": 1,
  "from_address": "0xabc123...",
  "to_address": "0xdef456...",
  "amount": 50,
  "signature": "encrypted_signature",
  "status": "PENDING"
}
```

### 3. Get Balance

Check wallet balance.

```http
GET /balance/:address
```

**Response:**

```json
{
  "address": "0xabc123...",
  "balance": 950
}
```

### 4. Generate Block

Mine a new block with pending transactions using Proof of Work.

```http
POST /generate-block
```

**Response:**

```json
{
  "block_number": 1,
  "current_hash": "00001a2b3c4d...",
  "previous_hash": "0000000000...",
  "nonce": 142536,
  "difficulty": 4,
  "merkle_root": "abc123def456...",
  "timestamp": 1732089600,
  "transactions_count": 5,
  "message": "Block mined successfully with PoW"
}
```

**Mining Process:**

- Collects all PENDING transactions
- Validates balances and wallets
- Calculates Merkle root from transactions
- Finds valid nonce through PoW mining
- Stores block with nonce, difficulty, and merkle_root
- Updates all transactions to CONFIRMED status

**Note:** Mining may take 5-15 seconds depending on difficulty and CPU power.

### 5. Get All Blocks

Retrieve the complete blockchain with transaction details.

```http
GET /blocks
```

**Response:**

```json
{
  "blocks": [
    {
      "id": 1,
      "block_number": 1,
      "previous_hash": "0",
      "current_hash": "00001a2b3c4d...",
      "nonce": 142536,
      "difficulty": 4,
      "merkle_root": "abc123def456...",
      "timestamp": 1732089600,
      "transactions": [
        {
          "id": 1,
          "from_address": "0xabc123...",
          "to_address": "0xdef456...",
          "amount": 50,
          "signature": "...",
          "status": "CONFIRMED"
        }
      ]
    }
  ],
  "message": "Blockchain retrieved successfully"
}
```

## 🏗️ Project Structure

```
go-blockchain-simulate/
├── app/
│   ├── handler/          # HTTP request handlers
│   │   ├── block.go      # Block generation endpoints
│   │   ├── transaction.go # Transaction endpoints
│   │   └── user.go       # Wallet registration and balance
│   ├── models/           # Data models
│   │   ├── block.go      # Block structure (with PoW fields)
│   │   ├── transaction.go # Transaction structure
│   │   └── user.go       # User/Wallet structure
│   ├── repository/       # Database layer (with bulk operations)
│   │   ├── block.go      # Block CRUD + BulkInsertBlockTransactions
│   │   ├── ledger.go     # Ledger CRUD + BulkCreateWithTx
│   │   ├── transaction.go # Transaction CRUD + BulkMarkConfirmed
│   │   └── user.go       # User CRUD + BulkUpdateBalances + LockMultiple
│   └── services/         # Business logic
│       ├── block.go      # Mining logic with PoW and Merkle trees
│       └── transaction.go # Transaction validation and creation
├── database/             # Database connection and migrations
│   ├── conn.go           # MySQL connection setup
│   └── migrations/       # SQL schema files
│       ├── 001_create_users.sql
│       ├── 002_create_transactions.sql
│       ├── 003_create_ledger.sql
│       ├── 004_create_blocks.sql
│       └── 005_create_block_transactions.sql
├── utils/                # Helper functions
│   └── fake-crypto.go    # PoW mining, Merkle trees, signatures, integrity checks
├── main.go               # Application entry point
├── go.mod                # Go module dependencies
├── Makefile              # Build and run commands
└── README.md             # This file
```

## 🔐 How It Works

### 1. **Wallet Creation**

```
Generate RSA Key Pair (2048-bit)
    ↓
Private Key → Public Key
    ↓
Hash Public Key → Address
```

### 2. **Transaction Flow**

```
Create Transaction
    ↓
Sign with Private Key → Signature
    ↓
Verify Signature with Public Key
    ↓
Add to Mempool (PENDING)
    ↓
Mine Block → Status: CONFIRMED
```

### 3. **Proof of Work Mining**

The system implements SHA256-based Proof of Work mining:

```
Fetch Pending Transactions
    ↓
Phase 1: Read Validation (No locks)
  - Verify sender/receiver wallets exist
  - Check balances in-memory
  - Pre-validate all transactions
    ↓
Phase 2: Mining (Outside DB transaction)
  - Calculate Merkle Root from transactions
  - Adjust difficulty if needed (every 10 blocks)
  - Find valid nonce (hash with N leading zeros)
  - Monitor mining progress
    ↓
Phase 3: Database Write (<2s transaction)
  - Bulk lock all involved users
  - Final balance validation
  - Bulk update balances (single CASE query)
  - Insert block with PoW data
  - Bulk mark transactions CONFIRMED
  - Bulk create ledger entries
  - Commit transaction
```

**Mining Algorithm:**

- Target: Hash with N leading zeros (difficulty = 4 means "0000...")
- Method: Increment nonce until valid hash found
- Hash: SHA256(block_number + previous_hash + merkle_root + timestamp + nonce)
- Average time: 5-15 seconds per block at difficulty 4

**Difficulty Adjustment:**

- Triggered every 10 blocks
- Target block time: 10 seconds
- Formula: `new_difficulty = old_difficulty ± 1` based on actual vs target time
- Range: 1 to 10 leading zeros

**Merkle Tree:**

- Binary hash tree of all transactions
- Root stored in block header
- Enables SPV (Simplified Payment Verification)
- Transaction format: `from_address|to_address|amount`

### 4. **Balance Calculation**

```
Query Ledger Table
    ↓
Sum all change_amount for address
    ↓
Return current balance
```

## 🚀 Performance Optimizations

The system implements several critical optimizations:

### Bulk Database Operations

- **Problem:** N+1 query pattern caused 50-second lock timeouts
- **Solution:** Single bulk queries for all operations
  - `GetMultipleByAddress`: Fetch all users in one query with `IN` clause
  - `LockMultipleUsersWithTx`: Lock all users atomically with `FOR UPDATE`
  - `BulkUpdateBalancesWithTx`: Update all balances in single `CASE` statement
  - `BulkMarkConfirmedWithTx`: Mark transactions in one query
  - `BulkCreateWithTx`: Insert all ledger entries at once
  - `BulkInsertBlockTransactionsWithTx`: Create junction records in one query

### Two-Phase Mining

- **Phase 1 (Read):** Pre-validation with no database locks
- **Phase 2 (Mine):** CPU-intensive mining outside transaction
- **Phase 3 (Write):** Ultra-short transaction (<2s) for final persistence

### Results

- **Before:** 50+ seconds per block (lock timeout errors)
- **After:** <2 seconds database transaction time
- **Improvement:** 95% faster with zero lock timeout errors
- **Mining time:** 5-15 seconds (separate from DB transaction)

## 🏗️ Architecture

### Clean Architecture Pattern

```
Handler Layer (HTTP)
    ↓
Service Layer (Business Logic)
    ↓
Repository Layer (Database)
    ↓
Database (MySQL)
```

### Key Components

- **handlers/**: HTTP request/response handling
- **services/**: Core blockchain logic (mining, validation)
- **repository/**: Database operations with bulk support
- **models/**: Data structures (Block, Transaction, User)
- **utils/**: Cryptographic utilities (PoW, Merkle, signatures)
- **database/**: Connection and migration management

## 🧪 Testing with cURL

### Complete Workflow Example

```bash
# 1. Register first wallet
WALLET1=$(curl -s -X POST http://localhost:3010/register \
  -H "Content-Type: application/json" \
  -d '{"initial_balance": 1000}')

echo "Wallet 1: $WALLET1"
ADDR1=$(echo $WALLET1 | jq -r '.address')
PRIVKEY1=$(echo $WALLET1 | jq -r '.private_key')

# 2. Register second wallet
WALLET2=$(curl -s -X POST http://localhost:3010/register \
  -H "Content-Type: application/json" \
  -d '{"initial_balance": 500}')

echo "Wallet 2: $WALLET2"
ADDR2=$(echo $WALLET2 | jq -r '.address')

# 3. Send multiple transactions
curl -X POST http://localhost:3010/send \
  -H "Content-Type: application/json" \
  -d "{
    \"from_address\": \"$ADDR1\",
    \"to_address\": \"$ADDR2\",
    \"private_key\": \"$PRIVKEY1\",
    \"amount\": 50
  }"

curl -X POST http://localhost:3010/send \
  -H "Content-Type: application/json" \
  -d "{
    \"from_address\": \"$ADDR1\",
    \"to_address\": \"$ADDR2\",
    \"private_key\": \"$PRIVKEY1\",
    \"amount\": 100
  }"

# 4. Check balances (transactions still PENDING)
curl http://localhost:3010/balance/$ADDR1
curl http://localhost:3010/balance/$ADDR2

# 5. Mine a block (this will take 5-15 seconds)
echo "Mining block with PoW..."
curl -X POST http://localhost:3010/generate-block

# 6. Check balances again (transactions now CONFIRMED)
curl http://localhost:3010/balance/$ADDR1  # Should be 850
curl http://localhost:3010/balance/$ADDR2  # Should be 650

# 7. View complete blockchain
curl http://localhost:3010/blocks | jq
```

### Individual Command Examples

```bash
# Register wallet
curl -X POST http://localhost:3010/register \
  -H "Content-Type: application/json" \
  -d '{"initial_balance": 1000}'

# Send transaction
curl -X POST http://localhost:3010/send \
  -H "Content-Type: application/json" \
  -d '{
    "from_address": "SENDER_ADDRESS",
    "to_address": "RECEIVER_ADDRESS",
    "private_key": "SENDER_PRIVATE_KEY",
    "amount": 50
  }'

# Check balance
curl http://localhost:3010/balance/WALLET_ADDRESS

# Mine block
curl -X POST http://localhost:3010/generate-block

# Get all blocks
curl http://localhost:3010/blocks
```

## 🔧 Makefile Commands

```bash
make run      # Run the application
make build    # Build binary to bin/app
make clean    # Remove build artifacts
```

## ⚠️ Important Notes

### Security

- **DO NOT use in production** - This is for educational purposes only
- Private keys are stored in plain text in the database
- Simplified RSA signatures (real blockchains use ECDSA)
- Single-node implementation (no network/peer-to-peer)
- Transaction fees and block rewards are implemented, but consensus and distributed mining are not
- Centralized mining (no consensus between multiple nodes)

### Performance

The system is optimized for educational demonstration:

- **Database:** Bulk operations prevent N+1 query issues
- **Mining:** Two-phase approach keeps database transactions short (<2s)
- **Concurrency:** Retry mechanism handles concurrent mining attempts
- **Scalability:** Suitable for hundreds of transactions per block

### Database Transactions

The system uses database transactions to ensure ACID compliance:

- All balance updates are atomic
- Rollback on any error
- Prevents double-spending and inconsistent states
- Bulk operations for performance (single query vs N queries)

### Transaction Validation

Every transaction is validated:

1. ✓ Sender wallet exists
2. ✓ Receiver wallet exists
3. ✓ Private key matches sender
4. ✓ Sufficient balance (checked twice: pre-validation + final validation)
5. ✓ Valid digital signature
6. ✓ Amount > 0

### Proof of Work

The PoW implementation is simplified but demonstrates core concepts:

- **Target:** N leading zeros in hash (configurable difficulty)
- **Range:** Difficulty 1-10 (production Bitcoin uses ~19-20 leading zeros)
- **Adjustment:** Every 10 blocks (Bitcoin adjusts every 2016 blocks)
- **Algorithm:** SHA256 (Bitcoin uses double SHA256)
- **Nonce:** 64-bit integer (Bitcoin uses 32-bit + extraNonce in coinbase)

## 📚 Key Concepts Demonstrated

- **Asymmetric Cryptography** (RSA key pairs)
- **Digital Signatures** (Sign & Verify)
- **Proof of Work** (SHA256 mining with nonce)
- **Dynamic Difficulty Adjustment** (Target block time maintenance)
- **Merkle Trees** (Transaction verification and SPV)
- **Hash Functions** (SHA256 for blocks and Merkle nodes)
- **Transaction Validation** (Balance checks, signature verification)
- **Transaction Fees & Block Rewards** (Miner incentives, direct fee-to-miner logic)
- **Miner Account** (All rewards and fees go to miner wallet)
- **Ledger Improvements** (Coinbase/reward entries, nullable `tx_id`)
- **Immutable Ledger** (Append-only transaction history)
- **Database Transactions** (ACID compliance with bulk operations)
- **Blockchain Integrity** (Chain validation with PoW verification)
- **REST API Design** (Clean HTTP interface)
- **Clean Architecture** (Separation of concerns: Handler → Service → Repository)

## 🐛 Troubleshooting

### Error: "wallet sender not found"

**Solution:** Register the wallet using `/register` endpoint before sending transactions.

### Error: "sql: no rows in result set"

**Solution:** Check if database has data and connection is configured correctly.

### Error: "invalid private key"

**Solution:** Use the exact private key returned from `/register` endpoint.

### Error: "Lock wait timeout exceeded; try restarting transaction"

**Cause:** Long-running database transaction holding locks while mining.

**Solution:** This has been fixed with two-phase mining approach. Mining now happens outside the database transaction. If you still encounter this:

1. Ensure you're using the latest code with bulk operations
2. Check for concurrent block generation attempts
3. Use the retry mechanism with exponential backoff

### Error: "insufficient balance"

**Cause:** Sender doesn't have enough balance for the transaction.

**Solution:**

1. Check balance using `/balance/:address` endpoint
2. Ensure all previous transactions are confirmed (block mined)
3. Register wallet with sufficient initial balance

### Error: MySQL reserved keyword (change)

**Solution:** Use backticks around column names or rename columns. In this project, `change` was renamed to `change_amount` in the ledger table.

### Mining takes too long (>30 seconds)

**Cause:** Difficulty too high for your CPU.

**Solution:**

1. Reduce difficulty in `utils/fake-crypto.go` (change `DefaultDifficulty` to 3 or 2)
2. Note: Difficulty auto-adjusts every 10 blocks, so wait for adjustment cycle
3. Check CPU usage - mining should use 100% of one core

### Blockchain integrity check fails

**Possible causes:**

1. Genesis block (block_number=1) has special handling - check if genesis validation is skipped
2. Merkle root NULL for existing blocks - update NULL values to empty string ''
3. Manual database changes - verify hash calculations match stored values

**Solution:** Run integrity check endpoint and review specific error messages. Genesis block should always pass validation.

## � Technical Deep Dive

### Proof of Work Implementation

The mining algorithm is implemented in `utils/fake-crypto.go`:

```go
func MineBlock(block *models.Block, difficulty int) error {
    target := strings.Repeat("0", difficulty)
    block.Nonce = 0

    for {
        hash := CalculateBlockHash(block)
        if strings.HasPrefix(hash, target) {
            block.CurrentHash = hash
            return nil
        }
        block.Nonce++

        // Progress indicator every 10,000 attempts
        if block.Nonce%10000 == 0 {
            log.Printf("Mining... Nonce: %d, Hash: %s", block.Nonce, hash[:10])
        }
    }
}
```

**Key aspects:**

- Brute force search incrementing nonce
- Hash must match difficulty target (N leading zeros)
- Average iterations at difficulty 4: ~65,536 (16^4)
- CPU-bound operation (intentionally expensive)

### Merkle Tree Construction

Merkle root is calculated from transaction hashes:

```go
func CalculateMerkleRoot(transactions []models.Transaction) string {
    if len(transactions) == 0 {
        return ""
    }

    var hashes []string
    for _, tx := range transactions {
        txString := fmt.Sprintf("%s|%s|%f", tx.FromAddress, tx.ToAddress, tx.Amount)
        hashes = append(hashes, HashData(txString))
    }

    // Build tree bottom-up
    for len(hashes) > 1 {
        var newLevel []string
        for i := 0; i < len(hashes); i += 2 {
            if i+1 < len(hashes) {
                combined := hashes[i] + hashes[i+1]
                newLevel = append(newLevel, HashData(combined))
            } else {
                newLevel = append(newLevel, hashes[i]) // Duplicate if odd
            }
        }
        hashes = newLevel
    }

    return hashes[0]
}
```

**Benefits:**

- Efficient verification of transaction inclusion
- Tamper-evident (any change invalidates root)
- Enables SPV for lightweight clients

### Bulk Database Operations

Example of bulk balance update (avoids N queries):

```go
func (u *userRepository) BulkUpdateBalancesWithTx(tx *sqlx.Tx, updates []BalanceUpdate) error {
    query := `UPDATE users SET balance = CASE address `

    var args []interface{}
    for _, update := range updates {
        query += `WHEN ? THEN ? `
        args = append(args, update.Address, update.NewBalance)
    }

    query += `END WHERE address IN (?)`
    addresses := make([]string, len(updates))
    for i, u := range updates {
        addresses[i] = u.Address
    }

    query, args, _ := sqlx.In(query, args..., addresses)
    _, err := tx.Exec(query, args...)
    return err
}
```

**Performance impact:**

- Before: N UPDATE queries (N transactions \* 2 users = 2N queries)
- After: 1 UPDATE with CASE statement
- Reduction: 95% faster for typical block with 5+ transactions

### Blockchain Integrity Validation

Complete chain validation checks:

1. **Genesis Block:** Special handling (block_number=1)
2. **Hash Chain:** Each block's previous_hash matches prior block's current_hash
3. **Proof of Work:** Each hash meets difficulty requirement (N leading zeros)
4. **Merkle Root:** Recalculated root matches stored root
5. **Transaction Integrity:** All transactions referenced in block exist and are CONFIRMED

## �📄 License

This project is open source and available under the MIT License.

## 👤 Author

**livingdolls**

- GitHub: [@livingdolls](https://github.com/livingdolls)

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

---

**Note:** This is a simplified blockchain implementation for educational purposes. It demonstrates core blockchain concepts but lacks many features required for a production cryptocurrency system (consensus mechanisms, network layer, advanced cryptography, etc.).

---

## 🧭 Planned Features

- **Block Explorer API**: Query blocks, transactions, and addresses (coming soon)
- **Block Explorer Web UI**: Browse blockchain data in a web interface
- **Wallet Management**: Improved wallet features and security
- **Analytics Dashboard**: Block and transaction statistics
- **Websocket Support**: Real-time updates for new blocks and transactions

See the [Roadmap](#-roadmap) section above for priorities.
