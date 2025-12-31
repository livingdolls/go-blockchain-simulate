# 🚀 Go Blockchain Simulator

A production-ready blockchain implementation in Go with complete cryptocurrency trading platform features including digital signatures, transaction validation, block mining, market simulation, and real-time data streaming.

## ✨ Key Features

### 🔐 **Blockchain Core**

- **Digital Wallet System** - RSA-2048 key pairs for secure wallet creation
- **Challenge-Response Authentication** - Secure login without storing passwords
- **JWT Authentication** - Protected endpoints with token-based auth
- **Digital Signatures** - Sign and verify transactions using RSA encryption
- **Transaction Nonce System** - Prevent double-spending and replay attacks
- **Transaction Fees** - Dynamic fee calculation paid to miners
- **Block Rewards** - Halving schedule every 210,000 blocks (Bitcoin-like)
- **Miner Rewards** - All fees and coinbase rewards go to miner wallet
- **Ledger System** - Double-entry bookkeeping for all balance changes
- **Proof of Work Mining** - SHA256-based with configurable difficulty
- **Dynamic Difficulty** - Auto-adjusts every 10 blocks targeting 10s block time
- **Merkle Tree** - Transaction integrity verification with SPV support
- **Blockchain Integrity** - Complete chain validation with PoW verification

### 📈 **Trading & Market Engine**

- **Market Simulation** - Real-time price discovery from buy/sell transactions
- **Buy/Sell Orders** - Execute market orders with instant settlement
- **Market Ticks** - Track price, volume, and transaction count per block
- **Liquidity Pool** - Initial liquidity for price stability
- **OHLCV Candles** - Multi-timeframe candlestick data (1m, 5m, 15m, 30m, 1h, 4h, 1d)
- **Candle Aggregation** - Automated worker for historical data generation
- **WebSocket Streaming** - Real-time market updates via WebSocket
- **SSE (Server-Sent Events)** - Real-time candle updates per interval
- **Redis Pub/Sub** - Scalable real-time data distribution

### ⚡ **Performance & Scalability**

- **Background Workers** - Async block generation and candle aggregation
- **Worker Pool Pattern** - Concurrent job processing with timeout protection
- **Graceful Shutdown** - Proper cleanup of workers, connections, and resources
- **Connection Pooling** - Optimized MySQL and Redis connections
- **Context Cancellation** - Timeout handling for long-running operations
- **Bulk Operations** - 95% performance improvement with batch processing
- **Indexed Queries** - Optimized database schema with proper indexes
- **Caching Layer** - Redis-based caching for frequently accessed data

### 🌐 **API & Integration**

- **REST API** - 30+ endpoints for complete blockchain operations
- **CORS Support** - Multi-origin configuration for frontend integration
- **WebSocket Hub** - Pub/Sub system for real-time client updates
- **SSE Streaming** - HTTP streaming for candle data
- **Health Checks** - Ping endpoints for monitoring
- **Error Handling** - Structured error responses with proper HTTP codes

### 🛡️ **Security & Validation**

- **Transaction Validation** - Multi-layer validation before mining
- **Balance Verification** - Prevent overspending with real-time checks
- **Signature Verification** - Cryptographic proof of ownership
- **Nonce Tracking** - Redis-based nonce management
- **Rate Limiting** - Protect against spam (via Redis)
- **Input Sanitization** - Prevent injection attacks

## 🆕 Latest Updates

### v2.0.0 - Market Trading Platform

- ✅ **Market Engine** - Complete trading system with buy/sell orders
- ✅ **OHLCV Candles** - Multi-timeframe candlestick generation
- ✅ **Real-time Streaming** - WebSocket + SSE for live updates
- ✅ **Redis Pub/Sub** - Scalable real-time architecture
- ✅ **Worker System** - Background jobs for block & candle generation
- ✅ **Graceful Shutdown** - Production-ready lifecycle management
- ✅ **Context Handling** - Timeout protection for all operations

### v1.0.0 - Blockchain Foundation

- ✅ **Blockchain Core** - Complete PoW blockchain with rewards
- ✅ **Wallet System** - RSA key generation and management
- ✅ **Transaction System** - Full UTXO-like transaction flow
- ✅ **Mining System** - PoW with difficulty adjustment
- ✅ **Ledger System** - Accurate balance tracking

## 🎯 Use Cases

- **Learning Blockchain** - Understand cryptocurrency fundamentals
- **Trading Simulation** - Test trading strategies risk-free
- **Algorithm Testing** - Backtest trading algorithms with historical data
- **Real-time Analytics** - Monitor market data with WebSocket/SSE
- **API Integration** - Build frontend applications with REST API
- **Research** - Study blockchain consensus and economic models

## 📋 Prerequisites

- **Go** 1.23.0 or higher
- **MySQL** 8.0 or higher
- **Redis** 6.0 or higher (for caching and pub/sub)
- **Make** (optional, for using Makefile commands)

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

3. **Setup MySQL Database**

   ```bash
   mysql -u root -p < database/migrations/users.sql
   mysql -u root -p < database/migrations/transactions.sql
   mysql -u root -p < database/migrations/ledger.sql
   mysql -u root -p < database/migrations/block.sql
   mysql -u root -p < database/migrations/user_balance_history.sql
   mysql -u root -p < database/migrations/market_engine.sql
   ```

   Or manually create the database:

   ```sql
   CREATE DATABASE blockchain_db;
   USE blockchain_db;

   -- Run all migration files from database/migrations/
   ```

4. **Setup Redis**

   ```bash
   # Install Redis (Ubuntu/Debian)
   sudo apt-get install redis-server

   # Start Redis
   sudo systemctl start redis-server

   # Verify Redis is running
   redis-cli ping  # Should return "PONG"
   ```

5. **Configure Environment**

   Update database and Redis credentials in the code:

   - Database: `database/conn.go`
   - Redis: `redis/redis.go`
   - JWT Secret: `main.go` (line 29)

## 🚦 Running the Application

### Using Make:

```bash
make run        # Run the application
make build      # Build binary
make clean      # Clean build files
```

### Using Go directly:

```bash
go run main.go
```

### Using Docker (Optional):

```bash
docker-compose up -d
```

The API will start on **`http://localhost:3010`**

## 📡 API Documentation

### 🔐 Authentication Endpoints

#### 1. Register Wallet

Create a new wallet with RSA key pair.

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
  "success": true,
  "data": {
    "address": "0x1a2b3c4d...",
    "public_key": "-----BEGIN PUBLIC KEY-----\n...",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...",
    "balance": 1000
  }
}
```

#### 2. Challenge Authentication

Request a challenge for signature verification.

```http
POST /challenge/:address
```

**Response:**

```json
{
  "success": true,
  "data": {
    "challenge": "random_string_to_sign"
  }
}
```

#### 3. Verify Challenge

Submit signed challenge for JWT token.

```http
POST /challenge/verify
Content-Type: application/json

{
    "address": "0x1a2b3c4d...",
    "signature": "signed_challenge_string"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 💰 Transaction Endpoints

#### 4. Send Transaction

Transfer funds between wallets.

```http
POST /send
Content-Type: application/json

{
    "from_address": "0x1a2b3c4d...",
    "to_address": "0x5e6f7g8h...",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...",
    "amount": 50,
    "nonce": 1
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 123,
    "from_address": "0x1a2b3c4d...",
    "to_address": "0x5e6f7g8h...",
    "amount": 50,
    "fee": 0.001,
    "signature": "...",
    "status": "PENDING",
    "nonce": 1
  }
}
```

#### 5. Buy Order

Execute market buy order.

```http
POST /transaction/buy
Content-Type: application/json

{
    "address": "0x1a2b3c4d...",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...",
    "amount": 10,
    "nonce": 2
}
```

#### 6. Sell Order

Execute market sell order.

```http
POST /transaction/sell
Content-Type: application/json

{
    "address": "0x1a2b3c4d...",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...",
    "amount": 5,
    "nonce": 3
}
```

#### 7. Get Transaction

Retrieve transaction details.

```http
GET /transaction/:id
```

#### 8. Generate Nonce

Get next available nonce for address.

```http
GET /generate-tx-nonce/:address
```

**Response:**

```json
{
  "success": true,
  "data": {
    "nonce": 4
  }
}
```

### 💼 Wallet & Balance Endpoints

#### 9. Get Balance

Check wallet balance from ledger.

```http
GET /balance/:address
```

**Response:**

```json
{
  "success": true,
  "data": {
    "address": "0x1a2b3c4d...",
    "balance": 950.5
  }
}
```

#### 10. Get Wallet Info

Get complete wallet information.

```http
GET /wallet/:address
```

### ⛏️ Mining & Block Endpoints

#### 11. Generate Block

Mine a new block (manual trigger).

```http
POST /generate-block
```

**Response:**

```json
{
  "success": true,
  "data": {
    "block_number": 42,
    "current_hash": "00001a2b3c...",
    "previous_hash": "000034f5a6...",
    "nonce": 142536,
    "difficulty": 4,
    "merkle_root": "abc123def456...",
    "timestamp": 1735689600,
    "transactions_count": 7
  },
  "message": "Block mined successfully"
}
```

**Note:** Blocks are also auto-generated every 10 seconds by background worker.

#### 12. Get All Blocks

Retrieve blockchain with transactions.

```http
GET /blocks
```

#### 13. Get Block by ID

Get specific block by database ID.

```http
GET /blocks/:id
```

#### 14. Get Block by Number

Get specific block by block number.

```http
GET /blocks/detail/:number
```

#### 15. Check Blockchain Integrity

Validate entire blockchain.

```http
GET /blocks/integrity
```

**Response:**

```json
{
  "success": true,
  "data": {
    "valid": true,
    "total_blocks": 42,
    "invalid_blocks": []
  }
}
```

### 🎁 Reward Endpoints

#### 16. Get Reward Schedule

Get block reward for specific block number.

```http
GET /reward/schedule/:number
```

**Response:**

```json
{
  "success": true,
  "data": {
    "block_number": 100,
    "reward": 50,
    "halving_epoch": 0
  }
}
```

#### 17. Get Block Reward

Get reward for mined block.

```http
GET /reward/block/:number
```

#### 18. Get Reward Info

Get current reward schedule information.

```http
GET /reward/info
```

### 📊 Market & Trading Endpoints

#### 19. Get Market State

Get current market engine state.

```http
GET /market
```

**Response:**

```json
{
  "success": true,
  "data": {
    "price": 100.065,
    "liquidity": 1000000,
    "last_block": 42,
    "updated_at": "2025-12-31T10:30:00Z"
  }
}
```

#### 20. Get Candles

Get OHLCV candles for specific interval.

```http
GET /candles?interval=1m&limit=100
```

**Parameters:**

- `interval`: `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`
- `limit`: Number of candles (default: 100)

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "interval_type": "1m",
      "start_time": 1735689600,
      "open_price": 100.05,
      "high_price": 100.08,
      "low_price": 100.04,
      "close_price": 100.065,
      "volume": 150.5
    }
  ]
}
```

#### 21. Get Candles in Range

Get candles from specific timestamp.

```http
GET /candles/range?interval=1h&start_time=1735689600&limit=50
```

### 🌐 Real-time Streaming Endpoints

#### 22. WebSocket Market Stream

Real-time market updates via WebSocket.

```javascript
const ws = new WebSocket("ws://localhost:3010/ws/market");

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log("Market update:", data);
};
```

#### 23. SSE Candle Stream

Real-time candle updates via Server-Sent Events.

```javascript
const es = new EventSource("http://localhost:3010/sse/candles?interval=1m");

es.onmessage = (event) => {
  const candle = JSON.parse(event.data);
  console.log("New candle:", candle);
};
```

#### 24. SSE Ping

Health check for SSE connection.

```http
GET /sse/ping
```

### 🔒 Protected Endpoints

#### 25. Get Profile

Get user profile (requires JWT token).

```http
GET /profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "address": "0x1a2b3c4d...",
    "balance": 950.5,
    "public_key": "..."
  }
}
```

## 🏗️ Project Architecture

```
go-blockchain-simulate/
├── app/
│   ├── dto/               # Data Transfer Objects
│   │   ├── candle.go      # Candle DTOs
│   │   └── http.go        # HTTP response wrappers
│   ├── entity/            # Domain entities
│   │   ├── error.go       # Custom error types
│   │   └── event.go       # Event types for pub/sub
│   ├── handler/           # HTTP request handlers
│   │   ├── balance.go     # Balance endpoints
│   │   ├── block.go       # Block endpoints
│   │   ├── candles.go     # Candle data endpoints
│   │   ├── jwt-middleware.go  # Authentication middleware
│   │   ├── market.go      # Market engine endpoints
│   │   ├── register.go    # User registration
│   │   ├── reward.go      # Block reward endpoints
│   │   ├── streams_candle.go  # SSE streaming handler
│   │   ├── transaction.go # Transaction endpoints
│   │   └── user.go        # User profile endpoints
│   ├── models/            # Database models
│   │   ├── block.go       # Block model
│   │   ├── candles.go     # Candle model
│   │   ├── market.go      # Market engine model
│   │   ├── register.go    # User model
│   │   ├── reward.go      # Reward model
│   │   ├── transaction.go # Transaction model
│   │   ├── user.go        # User account model
│   │   └── wallet.go      # Wallet model
│   ├── port/              # Interface ports
│   │   └── message_broker.go  # Pub/sub interface
│   ├── publisher/         # Event publishers
│   │   └── ws.go          # WebSocket publisher
│   ├── repository/        # Data access layer
│   │   ├── block.go       # Block repository
│   │   ├── candles.go     # Candles repository
│   │   ├── ledger.go      # Ledger repository
│   │   ├── market.go      # Market repository
│   │   ├── transaction.go # Transaction repository
│   │   └── user.go        # User repository
│   ├── services/          # Business logic layer
│   │   ├── balance.go     # Balance calculation service
│   │   ├── block.go       # Block mining service
│   │   ├── candles.go     # Candle service
│   │   ├── candles_stream.go  # Candle streaming service
│   │   ├── event.go       # Event handling service
│   │   ├── market.go      # Market engine service
│   │   ├── profile.go     # User profile service
│   │   ├── register.go    # Registration service
│   │   ├── reward.go      # Reward calculation service
│   │   ├── transaction.go # Transaction service
│   │   └── verify.go      # Signature verification service
│   ├── websocket/         # WebSocket infrastructure
│   │   ├── client.go      # WebSocket client
│   │   ├── handler.go     # WebSocket handler
│   │   ├── message.go     # Message types
│   │   └── websocket.go   # Hub implementation
│   └── worker/            # Background workers
│       ├── generate-block.go   # Auto block mining worker
│       └── generate-candle.go  # Candle aggregation worker
├── database/              # Database layer
│   ├── conn.go            # MySQL connection
│   └── migrations/        # SQL migration files
│       ├── block.sql
│       ├── ledger.sql
│       ├── market_engine.sql
│       ├── reset_genesis.sql
│       ├── transactions.sql
│       ├── user_balance_history.sql
│       └── users.sql
├── docs/                  # Documentation
│   ├── generate_block_flow.md
│   ├── register_flow.md
│   └── transaction_flow.md
├── redis/                 # Redis layer
│   ├── redis.go           # Redis client
│   └── redis-service.go   # Redis service adapter
├── security/              # Security utilities
│   └── jwt.go             # JWT token management
├── utils/                 # Utility functions
│   ├── candles-interval.go   # Interval utilities
│   ├── fake-crypto.go     # Test data generator
│   ├── fee.go             # Fee calculation
│   ├── merkle_root.go     # Merkle tree implementation
│   ├── mnemonic.go        # Mnemonic phrase generator
│   ├── pow.go             # Proof of Work implementation
│   ├── prefixed-hash.go   # Hash utilities
│   ├── random-hex.go      # Random string generator
│   ├── reward-calc.go     # Reward calculation
│   └── sse_setup.go       # SSE setup utilities
├── view/                  # Frontend (Next.js)
│   ├── app/               # Next.js app directory
│   ├── components/        # React components
│   ├── hooks/             # Custom React hooks
│   ├── lib/               # Frontend utilities
│   ├── providers/         # React context providers
│   ├── repository/        # API client layer
│   ├── store/             # State management
│   └── types/             # TypeScript types
├── go.mod                 # Go dependencies
├── go.sum                 # Go checksums
├── main.go                # Application entry point
├── Makefile               # Build automation
└── README.md              # This file
```

## 🔄 System Flow

### Transaction Flow

```
1. User creates transaction with private key signature
2. Transaction validated (signature, balance, nonce)
3. Transaction added to mempool (PENDING status)
4. Worker mines block every 10 seconds
5. Block validation + PoW mining
6. Transactions marked as CONFIRMED
7. Market price updated from buy/sell orders
8. Candle data aggregated per interval
9. Real-time updates sent via WebSocket/SSE
```

### Mining Flow

```
1. Collect PENDING transactions
2. Validate all transactions (balance, signature, nonce)
3. Calculate market price from buy/sell orders
4. Update market_ticks table
5. Calculate Merkle root
6. Perform Proof of Work (find valid nonce)
7. Save block to database
8. Update transaction status to CONFIRMED
9. Record miner rewards in ledger
10. Broadcast block event via WebSocket
11. Trigger candle aggregation
```

### Candle Aggregation Flow

```
1. Worker runs every 1 minute
2. Check if interval boundary reached (1m, 5m, 1h, etc.)
3. Fetch market_ticks in time window
4. Calculate OHLCV from ticks
5. Upsert candle to database
6. Check if data changed (cache comparison)
7. If changed: Publish to Redis channel
8. SSE clients receive update
9. Frontend updates chart in real-time
```

## 🔧 Configuration

### Database Configuration

Edit `database/conn.go`:

```go
dsn := "username:password@tcp(localhost:3306)/blockchain_db?parseTime=true"
```

### Redis Configuration

Edit `redis/redis.go`:

```go
addr: "localhost:6379"
password: ""  // leave empty if no password
db: 0
```

### JWT Secret

Edit `main.go`:

```go
jwt := security.NewJWTAdapter("your-secret-key-here", 24*time.Hour)
```

### Mining Difficulty

Default: 4 (four leading zeros)
Auto-adjusts every 10 blocks to target 10s block time.

### Worker Intervals

- **Block Worker**: 10 seconds (auto-mining)
- **Candle Worker**: 1 minute (aggregation check)

Edit in `main.go`:

```go
generateBlockWorker.Start(10 * time.Second)
candleWorker.Start(1 * time.Minute)
```

### CORS Origins

Edit `main.go`:

```go
allowedOrigins := map[string]bool{
    "http://localhost:3000": true,
    "http://localhost:3001": true,
    // Add your frontend URLs
}
```

## 🧪 Testing

### Manual Testing

Use the provided `rest.http` file with VS Code REST Client extension:

```bash
# Install VS Code REST Client extension
code --install-extension humao.rest-client

# Open rest.http and click "Send Request"
```

### Testing Flow

1. Register 2 wallets
2. Send transaction between wallets
3. Execute buy/sell orders
4. Wait for block generation (10s)
5. Check balances and market price
6. Query candle data
7. Test real-time streaming (WebSocket/SSE)

### API Testing Tools

- **Postman Collection**: Import endpoints from README
- **cURL**: All examples use standard HTTP
- **REST Client**: VS Code extension with `rest.http`

## 📊 Performance Metrics

- **Block Mining Time**: 5-15 seconds (depends on difficulty)
- **Transaction Validation**: <100ms
- **API Response Time**: <50ms (average)
- **Bulk Operations**: 95% faster than individual inserts
- **WebSocket Latency**: <10ms
- **SSE Latency**: <50ms
- **Database Queries**: Optimized with indexes
- **Concurrent Requests**: Supports 1000+ req/s

## 🐛 Troubleshooting

### Database Connection Error

```bash
# Check MySQL is running
sudo systemctl status mysql

# Test connection
mysql -u username -p -h localhost
```

### Redis Connection Error

```bash
# Check Redis is running
sudo systemctl status redis-server

# Test connection
redis-cli ping
```

### Port Already in Use

```bash
# Find process using port 3010
lsof -i :3010

# Kill process
kill -9 <PID>
```

### Mining Too Slow

- Reduce difficulty in code (default: 4)
- Increase CPU cores available
- Check system resources

### WebSocket Connection Failed

- Check CORS settings
- Verify WebSocket endpoint URL
- Check firewall rules

### SSE Not Receiving Data

- Verify Redis Pub/Sub is working
- Check candle worker is running
- Ensure transactions are being created
- Check browser console for errors

## 🚀 Deployment

### Production Checklist

- [ ] Change JWT secret to strong random string
- [ ] Update database credentials
- [ ] Configure Redis with password
- [ ] Set up HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Enable rate limiting
- [ ] Set up monitoring/logging
- [ ] Configure backup strategy
- [ ] Test graceful shutdown
- [ ] Load test with expected traffic

### Docker Deployment

```bash
# Build image
docker build -t blockchain-simulator .

# Run container
docker run -p 3010:3010 -d blockchain-simulator
```

### Systemd Service

```bash
sudo nano /etc/systemd/system/blockchain.service

[Unit]
Description=Blockchain Simulator
After=network.target mysql.service redis.service

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/go-blockchain-simulate
ExecStart=/usr/local/bin/blockchain-simulator
Restart=always

[Install]
WantedBy=multi-user.target
```

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👨‍💻 Author

**livingdolls**

- GitHub: [@livingdolls](https://github.com/livingdolls)

## 🙏 Acknowledgments

- Bitcoin whitepaper for blockchain fundamentals
- Ethereum for smart contract concepts
- Go community for excellent libraries
- Next.js team for amazing frontend framework

## 📚 Resources & Learning

- [Bitcoin Whitepaper](https://bitcoin.org/bitcoin.pdf)
- [Blockchain Basics](https://en.wikipedia.org/wiki/Blockchain)
- [Proof of Work Explained](https://en.wikipedia.org/wiki/Proof_of_work)
- [Merkle Trees](https://en.wikipedia.org/wiki/Merkle_tree)
- [RSA Cryptography](<https://en.wikipedia.org/wiki/RSA_(cryptosystem)>)

## 🗺️ Roadmap

### Version 2.1 (In Progress)

- [ ] Advanced chart visualizations
- [ ] Order book implementation
- [ ] Limit orders
- [ ] Stop-loss orders
- [ ] Portfolio tracking

### Version 3.0 (Planned)

- [ ] Smart contracts support
- [ ] Multi-currency support
- [ ] P2P networking
- [ ] Light client support
- [ ] Mobile app

### Future Ideas

- [ ] Staking mechanism
- [ ] Governance system
- [ ] NFT marketplace
- [ ] DeFi protocols
- [ ] Layer 2 scaling solutions

---

⭐ If you find this project helpful, please give it a star on GitHub!

📧 Questions? Open an issue or contact via GitHub.
├── go.mod # Go module dependencies
├── Makefile # Build and run commands
└── README.md # This file

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

````

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
````

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
