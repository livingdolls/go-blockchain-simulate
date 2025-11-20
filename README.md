# Go Blockchain Simulator

A simplified blockchain implementation in Go with REST API for learning cryptocurrency concepts including digital signatures, transaction validation, and block mining.

## 🚀 Features

- **Digital Wallet** - Generate RSA key pairs for wallet creation
- **Digital Signatures** - Sign and verify transactions using RSA encryption
- **Transaction Management** - Create, validate, and track transactions
- **Block Mining** - Mine blocks with pending transactions
- **Balance Tracking** - Calculate balances from transaction history via ledger
- **REST API** - HTTP endpoints for wallet registration, transactions, and block generation
- **Database Persistence** - Store users, transactions, blocks, and ledger entries in MySQL

## 📋 Prerequisites

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
       block_index INT NOT NULL,
       timestamp BIGINT NOT NULL,
       prev_hash VARCHAR(255) NOT NULL,
       hash VARCHAR(255) NOT NULL,
       transactions_count INT DEFAULT 0,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

Mine a new block with pending transactions.

```http
POST /generate-block
```

**Response:**

```json
{
  "block_index": 1,
  "hash": "abc123...",
  "prev_hash": "0",
  "timestamp": 1732089600,
  "transactions_count": 5,
  "message": "Block generated successfully"
}
```

## 🏗️ Project Structure

```
go-blockchain-simulate/
├── app/
│   ├── handler/          # HTTP request handlers
│   ├── models/           # Data models
│   ├── repository/       # Database layer
│   └── services/         # Business logic
├── block/                # Block structure and hashing
├── blockchain/           # Blockchain core logic
├── database/             # Database connection
├── signature/            # Digital signature utilities
├── transaction/          # Transaction models
├── utils/                # Helper functions
├── wallet/               # Wallet generation
├── main.go               # Application entry point
├── Makefile              # Build commands
└── README.md
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

### 3. **Block Mining**

```
Collect Pending Transactions
    ↓
Create Block with:
  - Index
  - Timestamp
  - Transactions
  - Previous Hash
    ↓
Calculate SHA256 Hash
    ↓
Append to Blockchain
    ↓
Update Transaction Status
```

### 4. **Balance Calculation**

```
Query Ledger Table
    ↓
Sum all change_amount for address
    ↓
Return current balance
```

## 🧪 Testing with cURL

```bash
# Register two wallets
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
- No proof-of-work or consensus mechanism
- Single-node implementation (no network)

### Database Transactions

The system uses database transactions to ensure ACID compliance:

- All balance updates are atomic
- Rollback on any error
- Prevents double-spending and inconsistent states

### Transaction Validation

Every transaction is validated:

1. ✓ Sender wallet exists
2. ✓ Private key matches sender
3. ✓ Sufficient balance
4. ✓ Valid digital signature

## 📚 Key Concepts Demonstrated

- **Asymmetric Cryptography** (RSA key pairs)
- **Digital Signatures** (Sign & Verify)
- **Hash Functions** (SHA256 for blocks)
- **Transaction Validation**
- **Immutable Ledger**
- **Database Transactions** (ACID)
- **REST API Design**

## 🐛 Troubleshooting

### Error: "wallet sender not found"

**Solution:** Register the wallet using `/register` endpoint before sending transactions.

### Error: "sql: no rows in result set"

**Solution:** Check if database has data and connection is configured correctly.

### Error: "invalid private key"

**Solution:** Use the exact private key returned from `/register` endpoint.

### Error: MySQL reserved keyword (change)

**Solution:** Use backticks around column names or rename columns to avoid reserved keywords.

## 📄 License

This project is open source and available under the MIT License.

## 👤 Author

**livingdolls**

- GitHub: [@livingdolls](https://github.com/livingdolls)

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

---

**Note:** This is a simplified blockchain implementation for educational purposes. It demonstrates core blockchain concepts but lacks many features required for a production cryptocurrency system (consensus mechanisms, network layer, advanced cryptography, etc.).
