# Frontend API Documentation

Dokumentasi ini ditujukan untuk frontend developer yang akan mengintegrasikan
aplikasi dengan backend blockchain. Semua endpoint mengembalikan JSON.

---

## Daftar Isi

- [Base URL & CORS](#base-url--cors)
- [Authentication](#authentication)
- [Standard Response Format](#standard-response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Auth](#1-register)
  - [Profile](#6-get-profile)
  - [Balance](#7-get-balance)
  - [Transaction](#10-send-transaction)
  - [Block](#14-get-blocks)
  - [Reward](#22-get-reward-info)
  - [Market](#25-get-market-state)
  - [Candle](#26-get-candle)
  - [SSE Streaming](#28-sse-candles)
  - [WebSocket](#30-websocket-market)
  - [Admin](#admin-endpoints)
- [WebSocket Events](#websocket-events)
- [Signature Flow](#signature-flow)
- [Idempotency](#idempotency)
- [Error Codes Reference](#error-codes-reference)
- [Data Models](#data-models)

---

## Base URL & CORS

```
Base URL: http://localhost:3010
```

CORS diaktifkan untuk origin yang diizinkan (tergantung konfigurasi server).
Request cross-origin dari domain yang tidak diizinkan akan ditolak.

---

## Authentication

Backend menggunakan **JWT cookie-based authentication**. Setelah login/register
berhasil, token disimpan di cookie `auth_token` (HttpOnly, SameSite=Strict).

**Cara kerja:**
1. Frontend mengirim request ke `/register` atau `/challenge/verify`
2. Backend mengembalikan `Set-Cookie: auth_token=...` (HttpOnly)
3. Browser otomatis mengirim cookie ini di setiap request berikutnya
4. Tidak perlu menambahkan header Authorization secara manual

**Cookie properties:**
- `HttpOnly: true` — tidak bisa diakses via JavaScript (XSS protection)
- `SameSite: Strict` — tidak dikirim untuk cross-site request (CSRF protection)
- `Secure: false` (dev) / `true` (production via HTTPS)
- `Max-Age: 86400` (24 jam)

**Logout:** Kirim POST ke `/admin/auth/logout` (admin) — cookie akan dihapus.

---

## Standard Response Format

Semua response menggunakan format konsisten:

### Success
```json
{
  "success": true,
  "data": { ... }
}
```

### Error
```json
{
  "success": false,
  "error": "pesan error"
}
```

### Validation Error (400)
```json
{
  "success": false,
  "code": 400,
  "error": "validation failed",
  "fields": [
    {"field": "Amount", "message": "harus lebih besar dari 0"},
    {"field": "Address", "message": "format address Ethereum tidak valid"}
  ]
}
```

---

## Error Handling

| HTTP Code | Arti | Frontend Action |
|-----------|------|-----------------|
| `200` | Success | Tampilkan data |
| `201` | Created (register) | Tampilkan success, redirect |
| `202` | Accepted (transaction submitted) | Tampilkan "sedang diproses" |
| `400` | Bad request / validation error | Tampilkan pesan error ke user |
| `401` | Unauthorized (token invalid/expired) | Redirect ke login |
| `403` | Forbidden (self-only check gagal) | Tampilkan pesan error |
| `404` | Not found | Tampilkan "tidak ditemukan" |
| `429` | Too many requests | Tunggu, retry dengan backoff |
| `500` | Internal server error | Tampilkan "server error, coba lagi" |

---

## Rate Limiting

| Endpoint Group | Limit | Window | Identifier |
|---------------|-------|--------|------------|
| `/register`, `/challenge/*` | 10 req | 1 menit | Client IP |
| `/transaction/*` | 10 req | 1 menit | Address dari body |
| `/transaction/*` | 30 req | 1 menit | Client IP |
| `/admin/auth/*` | 10 req | 1 menit | Client IP |

Response 429:
```json
{
  "success": false,
  "code": 429,
  "error": "Terlalu banyak request untuk address ini, coba lagi nanti"
}
```

---

## Endpoints

### 1. Register

Buat akun baru. Mengembalikan JWT cookie.

```
POST /register
```

**Request Body:**
```json
{
  "username": "alice",
  "address": "0x1234567890abcdef1234567890abcdef12345678",
  "public_key": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}
```

**Validation:**
- `username`: required, 3-50 karakter
- `address`: required, Ethereum address format (0x + 40 hex)
- `public_key`: required, hex string

**Response (201):**
```json
{
  "success": true,
  "data": {
    "username": "alice",
    "address": "0x1234567890abcdef1234567890abcdef12345678",
    "yte_balance": 0,
    "usd_balance": 0,
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 2. Challenge (Dapatkan Nonce)

Mendapatkan nonce untuk challenge-response authentication.

```
POST /challenge/:address
```

**Path Parameter:**
- `address` — Ethereum address (0x + 40 hex)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "challenge": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

### 3. Verify (Login)

Verifikasi signature dan dapatkan JWT cookie.

```
POST /challenge/verify
```

**Request Body:**
```json
{
  "address": "0x1234567890abcdef1234567890abcdef12345678",
  "nonce": "550e8400-e29b-41d4-a716-446655440000",
  "signature": "0x1234...abc",
  "username": "alice"
}
```

**Validation:**
- `address`: required, Ethereum address format
- `nonce`: required
- `signature`: required, 132 karakter (0x + 130 hex = 65 byte)
- `username`: required, 3-50 karakter

**Response (200):**
```json
{
  "success": true,
  "data": {
    "valid": true
  }
}
```

---

### 4. Generate Transaction Nonce

Mendapatkan nonce untuk transaksi (digunakan sebelum send/buy/sell).

```
GET /generate-tx-nonce/:address
```

**Path Parameter:**
- `address` — Ethereum address

**Response (200):**
```json
{
  "nonce": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

---

### 5. Send Transaction

Mengirim transaksi transfer. Endpoint ini menggunakan **signature-based authentication** (bukan JWT). Frontend harus menandatangani pesan canonical dengan private key user.

```
POST /transaction/send
```

**Request Body:**
```json
{
  "from_address": "0x1234567890abcdef1234567890abcdef12345678",
  "to_address": "0xabcdef1234567890abcdef1234567890abcdef12",
  "amount": 10.5,
  "nonce": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "signature": "0x1234...abc"
}
```

**Validation:**
- `from_address`: required, Ethereum address
- `to_address`: required, Ethereum address
- `amount`: required, > 0
- `nonce`: required
- `signature`: required, 132 karakter

**Pesan yang ditandatangani:**
```
Send 10.50 to 0xabcdef...nonce:a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Response (202):**
```json
{
  "success": true,
  "data": {
    "message": "Transaction submitted successfully and is being processed"
  }
}
```

**Idempotency:** Kirim header `Idempotency-Key` untuk mencegah double-submit.

---

### 6. Buy Transaction

Membeli YTE dari sistem (miner). Membayar dalam USD.

```
POST /transaction/buy
```

**Request Body:**
```json
{
  "address": "0x1234567890abcdef1234567890abcdef12345678",
  "amount": 5.0,
  "nonce": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
  "signature": "0x1234...abc"
}
```

**Pesan yang ditandatangani:**
```
 BUY 5.00 nonce:b2c3d4e5-f6a7-8901-bcde-f12345678901
```

---

### 7. Sell Transaction

Menjual YTE ke sistem. Menerima USD.

```
POST /transaction/sell
```

**Request Body:**
```json
{
  "address": "0x1234567890abcdef1234567890abcdef12345678",
  "amount": 3.0,
  "nonce": "c3d4e5f6-a7b8-9012-cdef-123456789012",
  "signature": "0x1234...abc"
}
```

**Pesan yang ditandatangani:**
```
 SELL 3.00 nonce:c3d4e5f6-a7b8-9012-cdef-123456789012
```

---

### 8. Get Transaction

```
GET /transaction/:id
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "transaction": {
      "id": 42,
      "from_address": "0x1234...",
      "to_address": "0xabcd...",
      "amount": 10.5,
      "fee": 0.01,
      "type": "TRANSFER",
      "signature": "0x...",
      "status": "confirmed",
      "created_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

---

### 9. Get Balance (USD + YTE)

```
GET /balance/:address
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "address": "0x1234...",
    "name": "alice",
    "yte_balance": 150.5,
    "usd_balance": 2500.00
  }
}
```

---

### 10. Top Up USD Balance

```
POST /balance/topup
```

**Requires:** JWT authentication (cookie `auth_token`).
**Self-only:** Hanya bisa topup saldo sendiri (address di body harus match JWT).

**Request Body:**
```json
{
  "address": "0x1234567890abcdef1234567890abcdef12345678",
  "amount": 1000.0,
  "reference_id": "TXN-2024-001",
  "description": "Deposit via bank transfer"
}
```

**Validation:**
- `address`: required, Ethereum address, harus match JWT address
- `amount`: required, > 0
- `reference_id`: optional, max 100 karakter
- `description`: optional, max 500 karakter

**Response (200):**
```json
{
  "success": true,
  "data": {
    "address": "0x1234...",
    "amount": 1000.0,
    "balance_before": 500.0,
    "balance_after": 1500.0,
    "reference_id": "TXN-2024-001"
  }
}
```

---

### 11. Get Wallet Balance (dengan transaksi)

```
GET /wallet/:address?type=all&page=1&limit=10
```

**Query Parameters:**
| Parameter | Default | Values |
|-----------|---------|--------|
| `type` | `all` | `all`, `send`, `received`, `buy`, `sell` |
| `status` | `all` | `all`, `pending`, `confirmed` |
| `page` | `1` | ≥ 1 |
| `limit` | `10` | 1-100 |
| `sort_by` | `id` | `id`, `amount` |
| `order` | `desc` | `asc`, `desc` |

---

### 12. Get Blocks

```
GET /blocks?limit=10&offset=0
```

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "block_number": 1,
      "previous_hash": "0000...",
      "current_hash": "0000abc...",
      "nonce": 12345,
      "difficulty": 4,
      "timestamp": 1705312200,
      "merkle_root": "def456...",
      "miner_address": "MINER_ACCOUNT",
      "block_reward": 50.0,
      "total_fees": 0.5,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

---

### 13. Get Block by ID

```
GET /blocks/:id
```

---

### 14. Get Block by Block Number

```
GET /blocks/detail/:number
```

---

### 15. Get Transactions by Block Number

```
GET /blocks/transaction/:number
```

---

### 16. Search Blocks by Hash

```
GET /blocks/search?hash=0000abc
```

---

### 17. Get Blocks in Range

```
GET /blocks/range?from=1&to=100
```

**Validation:** `to - from ≤ 1000`

---

### 18. Get Block Stats

```
GET /blocks/stats
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_blocks": 150,
    "average_difficulty": 4.2,
    "total_transactions": 450,
    "total_fees": 12.5,
    "average_block_reward": 50.0,
    "avg_tx_per_block": 3.0,
    "latest_block_number": 150
  }
}
```

---

### 19. Check Blockchain Integrity

```
GET /blocks/integrity
```

**Response (200):**
```json
{
  "status": "valid"
}
```

Atau jika invalid:
```json
{
  "status": "invalid",
  "error": "block 42: hash mismatch"
}
```

---

### 20. Search Blocks by Miner Address

```
GET /blocks/search/miner/?address=0x1234...&limit=10&offset=0
```

---

### 21. Generate Block (Admin Only)

```
POST /blocks/generate
```

**Requires:** Admin JWT authentication.
**Note:** Block generation otomatis terjadi setiap 10 detik via worker.
Endpoint ini hanya untuk force-mine manual.

---

### 22. Get Reward Info

```
GET /reward/info
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "current_block_number": 150,
    "current_reward": 50.0,
    "next_reward": 50.0,
    "next_halving_block": 200,
    "blocks_until_halving": 50,
    "current_supply": 7500.0,
    "max_supply": 10000.0,
    "supply_percentage": 75.0
  }
}
```

---

### 23. Get Block Reward

```
GET /reward/block/:number
```

---

### 24. Get Reward Schedule

```
GET /reward/schedule/:number?blocks=10
```

**Query Parameters:**
- `blocks` — jumlah block yang ingin ditampilkan (default 10, max 1000)

---

### 25. Get Market State

```
GET /market
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "price": 200.50,
    "volume_24h": 15000.0,
    "high_24h": 210.0,
    "low_24h": 195.0,
    "change_24h": 2.5
  }
}
```

---

### 26. Get Candle

```
GET /candles?interval=1m&type=latest
```

**Query Parameters:**
| Parameter | Values |
|-----------|--------|
| `interval` | `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d` |
| `type` | `latest`, `all` |

---

### 27. Get Candle Range

```
GET /candles/range?interval=1h&from=1705312200&to=1705398600
```

---

### 28. SSE Candles

Server-Sent Events stream untuk data candle real-time.

```
GET /sse/candles?interval=1m
```

**Response:** SSE stream
```
data: {"open":200.0,"high":205.0,"low":198.0,"close":203.0,"volume":1500.0}

data: {"open":203.0,"high":207.0,"low":201.0,"close":206.0,"volume":1200.0}
```

**Frontend implementation:**
```javascript
const eventSource = new EventSource('/sse/candles?interval=1m');
eventSource.onmessage = (event) => {
  const candle = JSON.parse(event.data);
  updateChart(candle);
};
eventSource.onerror = () => {
  // Reconnect setelah delay
  setTimeout(() => eventSource.close(), 5000);
};
```

---

### 29. SSE Ping

```
GET /sse/ping
```

Response: SSE heartbeat setiap beberapa detik.

---

### 30. WebSocket Market

Real-time market data, balance updates, dan transaksi events.

```
GET /ws/market
```

**Requires:** JWT authentication (cookie `auth_token`).
**Origin validation:** Hanya origin yang diizinkan yang bisa connect.

**Frontend implementation:**
```javascript
const ws = new WebSocket('ws://localhost:3010/ws/market');

ws.onopen = () => {
  // Subscribe ke event types
  ws.send(JSON.stringify({
    type: 'subscribe',
    data: {
      events: ['market_price', 'balance_update', 'new_transaction']
    }
  }));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  switch (msg.type) {
    case 'market_price':
      updatePrice(msg.data);
      break;
    case 'balance_update':
      updateBalance(msg.data);
      break;
    case 'new_transaction':
      showNotification(msg.data);
      break;
  }
};

ws.onclose = () => {
  // Reconnect setelah delay
  setTimeout(connectWebSocket, 5000);
};
```

---

## WebSocket Events

### Subscribe

```json
{
  "type": "subscribe",
  "data": {
    "events": ["market_price", "balance_update"]
  }
}
```

### Unsubscribe

```json
{
  "type": "unsubscribe",
  "data": {
    "events": ["market_price"]
  }
}
```

### Available Event Types

| Event Type | Deskripsi |
|-----------|-----------|
| `market_price` | Perubahan harga market |
| `balance_update` | Perubahan saldo user |
| `new_transaction` | Transaksi baru dikonfirmasi |
| `new_block` | Block baru ditambang |
| `candle_update` | Update candle data |

---

## Signature Flow

Untuk transaksi (send/buy/sell), frontend harus menandatangani pesan
canonical menggunakan private key user.

### Step-by-step:

1. **Dapatkan nonce:**
   ```
   GET /generate-tx-nonce/:address
   ```
   Response: `{ "nonce": "abc-123" }`

2. **Bangun pesan canonical:**

   Untuk SEND:
   ```
   Send 10.50 to 0xRecipientAddress nonce:abc-123
   ```

   Untuk BUY:
   ```
    BUY 5.00 nonce:abc-123
   ```

   Untuk SELL:
   ```
    SELL 3.00 nonce:abc-123
   ```

3. **Hash pesan dengan Ethereum personal_sign prefix:**
   ```javascript
   // Menggunakan ethers.js
   const message = `Send 10.50 to ${toAddress} nonce:${nonce}`;
   const signature = await signer.signMessage(message);
   ```

   Atau secara manual:
   ```javascript
   // Prefixed hash: keccak256("\x19Ethereum Signed Message:\n" + len + message)
   const hash = ethers.hashMessage(message);
   const signature = await signer.signMessage(message);
   ```

4. **Kirim transaksi:**
   ```javascript
   const response = await fetch('/transaction/send', {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json',
       'Idempotency-Key': generateUUID() // untuk mencegah double-submit
     },
     body: JSON.stringify({
       from_address: myAddress,
       to_address: recipientAddress,
       amount: 10.5,
       nonce: nonce,
       signature: signature
     })
   });
   ```

---

## Idempotency

Untuk mencegah double-submit saat client retry, gunakan header
`Idempotency-Key` pada POST request:

```javascript
const idempotencyKey = crypto.randomUUID(); // atau UUID library

fetch('/transaction/send', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Idempotency-Key': idempotencyKey
  },
  body: JSON.stringify(transactionData)
});
```

**Behavior:**
- Request pertama dengan key tersebut: diproses normal
- Request berikutnya dengan key yang sama (dalam 24 jam): mengembalikan
  response yang sama persis tanpa memproses ulang
- Header `Idempotent-Replayed: true` ditambahkan pada replay

**Scope:** Hanya berlaku untuk `POST /transaction/send`, `/buy`, `/sell`.

---

## Error Codes Reference

### Validation Errors

| Field | Error Message | Arti |
|-------|--------------|------|
| `Address` | `format address Ethereum tidak valid (0x + 40 hex)` | Address bukan format ETH |
| `FromAddress` | `field wajib diisi` | Address pengirim kosong |
| `ToAddress` | `field wajib diisi` | Address penerima kosong |
| `Amount` | `harus lebih besar dari 0` | Amount harus positif |
| `Amount` | `field wajib diisi` | Amount kosong |
| `Nonce` | `field wajib diisi` | Nonce kosong |
| `Signature` | `panjang harus 132` | Signature bukan 65 byte hex |
| `Username` | `panjang harus >= 3` | Username terlalu pendek |
| `Username` | `panjang harus <= 50` | Username terlalu panjang |
| `PublicKey` | `harus hex string valid` | Public key bukan hex |
| `Role` | `harus salah satu dari: admin moderator support` | Role tidak valid |
| `Status` | `harus salah satu dari: active inactive suspended` | Status tidak valid |
| `Amount` | `harus lebih besar dari 0` | Topup amount harus positif |

### Business Logic Errors

| Error | Arti |
|-------|------|
| `invalid nonce` | Nonce tidak ditemukan atau sudah expired |
| `address mismatch` | Signature tidak cocok dengan address |
| `insufficient balance` | Saldo tidak cukup |
| `transaction not found` | Transaksi tidak ditemukan |
| `invalid Ethereum address format` | Format address salah |
| `Invalid username or password` | Login gagal (generic) |
| `forbidden: hanya bisa topup saldo sendiri` | Mencoba topup address lain |

---

## Data Models

### Block

```typescript
interface Block {
  id: number;
  block_number: number;
  previous_hash: string;
  current_hash: string;
  nonce: number;
  difficulty: number;
  timestamp: number;        // Unix timestamp
  merkle_root: string;
  miner_address: string;
  block_reward: number;
  total_fees: number;
  created_at: string;       // ISO 8601
  transactions?: Transaction[];
}
```

### Transaction

```typescript
interface Transaction {
  id: number;
  from_address: string;
  to_address: string;
  amount: number;
  fee: number;
  type: 'TRANSFER' | 'BUY' | 'SELL';
  signature: string;
  status: 'pending' | 'confirmed';
  created_at: string;
}
```

### BlockStats

```typescript
interface BlockStats {
  total_blocks: number;
  average_difficulty: number;
  total_transactions: number;
  total_fees: number;
  average_block_reward: number;
  avg_tx_per_block: number;
  latest_block_number: number;
}
```

### RewardInfo

```typescript
interface RewardInfo {
  current_block_number: number;
  current_reward: number;
  next_reward: number;
  next_halving_block: number;
  blocks_until_halving: number;
  current_supply: number;
  max_supply: number;
  supply_percentage: number;
}
```

### TopUpResult

```typescript
interface TopUpResult {
  address: string;
  amount: number;
  balance_before: number;
  balance_after: number;
  reference_id?: string;
  description?: string;
}
```

### MarketState

```typescript
interface MarketState {
  price: number;
  volume_24h: number;
  high_24h: number;
  low_24h: number;
  change_24h: number;
}
```

### Candle

```typescript
interface Candle {
  interval: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  timestamp: number;
}
```

---

## Admin Endpoints

### Admin Login

```
POST /admin/auth/login
```

**Request Body:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "username": "admin",
    "role": "admin",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### Admin Logout

```
POST /admin/auth/logout
```

### Admin Dashboard

```
GET /admin/dashboard
```

### List Admins

```
GET /admin/admins?limit=10&offset=0
```

### Create Admin

```
POST /admin/admins
```

**Request Body:**
```json
{
  "user_id": 1,
  "role": "admin",
  "permissions": ["*"]
}
```

**Validation:**
- `user_id`: required, > 0
- `role`: required, salah satu dari `admin`, `moderator`, `support`
- `permissions`: optional

### Update Admin Role

```
PUT /admin/admins/:id/role
```

**Request Body:**
```json
{
  "role": "moderator",
  "permissions": ["read", "write"]
}
```

### Update Admin Status

```
PUT /admin/admins/:id/status
```

**Request Body:**
```json
{
  "status": "active"
}
```

**Validation:**
- `status`: required, salah satu dari `active`, `inactive`, `suspended`

### Delete Admin

```
DELETE /admin/admins/:id
```

### Get Activity Logs

```
GET /admin/activity-logs?admin_id=1&action=login&limit=50&offset=0
```

### Recent Activity Logs

```
GET /admin/activity-logs/recent
```

---

## Health Check

```
GET /healthz
```

Response: `200 OK` — server hidup (liveness probe).

```
GET /readyz
```

Response: `200 OK` jika DB dan Redis terkoneksi, `503 Service Unavailable` jika tidak.
