# API Documentation

Base URL (contoh):

- `http://localhost:8080`

## Auth

### POST /register

Register user baru.

Request body (contoh):

```json
{
  "username": "alice",
  "email": "alice@mail.com",
  "password": "secret"
}
```

Response:

- `200 OK` / `201 Created`: registrasi berhasil
- `400 Bad Request`: payload tidak valid

### POST /challenge/:address

Minta challenge message untuk verifikasi wallet address.

Path params:

- `address` (string): alamat wallet

Response:

- `200 OK`: challenge berhasil dibuat
- `404 Not Found`: address tidak ditemukan (tergantung implementasi)

### POST /challenge/verify

Verifikasi challenge + signature.

Request body (contoh):

```json
{
  "address": "0x123...",
  "signature": "0xabc...",
  "message": "challenge-message"
}
```

Response:

- `200 OK`: verifikasi sukses (biasanya return token/session)
- `401 Unauthorized`: signature tidak valid

## Transaction

### POST /transaction/send

Kirim transaksi antar wallet.

Request body (contoh):

```json
{
  "from": "0xaaa...",
  "to": "0xbbb...",
  "amount": 1.25,
  "nonce": 10
}
```

Response:

- `200 OK`: transaksi diterima
- `400 Bad Request`: data transaksi tidak valid

### GET /transaction/:id

Ambil detail transaksi berdasarkan ID.

Path params:

- `id` (string): ID transaksi

Response:

- `200 OK`: detail transaksi
- `404 Not Found`: transaksi tidak ditemukan

### POST /transaction/buy

Beli aset crypto (market buy sesuai implementasi).

Request body (contoh):

```json
{
  "address": "0xabc...",
  "symbol": "BTC",
  "usd_amount": 100
}
```

Response:

- `200 OK`: order berhasil diproses
- `400 Bad Request`: request invalid / saldo tidak cukup

### POST /transaction/sell

Jual aset crypto.

Request body (contoh):

```json
{
  "address": "0xabc...",
  "symbol": "BTC",
  "amount": 0.01
}
```

Response:

- `200 OK`: order berhasil diproses
- `400 Bad Request`: request invalid / saldo aset tidak cukup

### GET /generate-tx-nonce/:address

Generate / ambil nonce transaksi terakhir untuk address.

Path params:

- `address` (string): alamat wallet

Response:

- `200 OK`: nonce saat ini
- `404 Not Found`: address tidak ditemukan (tergantung implementasi)

## Balance & Wallet

### GET /balance/:address

Ambil balance user + USD balance.

Path params:

- `address` (string): alamat wallet

Response:

- `200 OK`: data balance user
- `404 Not Found`: wallet/user tidak ditemukan

### POST /balance/topup

Top up saldo USD user.

Request body (contoh):

```json
{
  "address": "0xabc...",
  "amount": 500
}
```

Response:

- `200 OK`: top up berhasil
- `400 Bad Request`: nominal invalid

### GET /wallet/:address

Ambil wallet balance (aset crypto).

Path params:

- `address` (string): alamat wallet

Response:

- `200 OK`: detail saldo wallet
- `404 Not Found`: wallet tidak ditemukan

## Blocks

Prefix: `/blocks`

### POST /blocks/generate

Generate block baru (manual trigger).

Response:

- `200 OK`: block berhasil dibuat
- `500 Internal Server Error`: gagal generate block

### GET /blocks

Ambil list blocks.

Query params (opsional, tergantung implementasi):

- pagination / filter

Response:

- `200 OK`: daftar block

### GET /blocks/:id

Ambil block berdasarkan ID database.

Path params:

- `id` (string|number)

Response:

- `200 OK`: detail block
- `404 Not Found`: block tidak ditemukan

### GET /blocks/detail/:number

Ambil block berdasarkan block number.

Path params:

- `number` (number)

Response:

- `200 OK`: detail block
- `404 Not Found`: block number tidak ditemukan

### GET /blocks/integrity

Cek integritas blockchain.

Response:

- `200 OK`: status integritas chain

### GET /blocks/transaction/:number

Ambil transaksi pada block number tertentu.

Path params:

- `number` (number): block number

Response:

- `200 OK`: list transaksi dalam block
- `404 Not Found`: block tidak ditemukan

### GET /blocks/search

Cari block berdasarkan hash.

Query params:

- `hash` (string): hash block

Response:

- `200 OK`: hasil pencarian block

### GET /blocks/range

Ambil blocks dalam rentang number tertentu.

Query params (contoh):

- `from` (number)
- `to` (number)

Response:

- `200 OK`: daftar blocks dalam range

### GET /blocks/stats

Ambil statistik blockchain/blocks.

Response:

- `200 OK`: ringkasan statistik

### GET /blocks/search/miner/

Cari blocks berdasarkan miner address.

Query params (contoh):

- `address` (string): alamat miner

Response:

- `200 OK`: daftar block oleh miner tersebut

Catatan:

- Endpoint ini memiliki trailing slash sesuai route yang terdaftar.

## Reward

Prefix: `/reward`

### GET /reward/schedule/:number

Ambil jadwal reward berdasarkan block number.

Path params:

- `number` (number)

Response:

- `200 OK`: detail reward schedule

### GET /reward/block/:number

Ambil reward untuk block tertentu.

Path params:

- `number` (number)

Response:

- `200 OK`: detail reward block

### GET /reward/info

Ambil info umum reward system.

Response:

- `200 OK`: metadata reward

## Market

### GET /market

Ambil state market engine saat ini.

Response:

- `200 OK`: data market engine state

## Candles

Prefix: `/candles`

### GET /candles

Ambil candle terkini / berdasarkan default interval.

Query params:

- interval/symbol (tergantung implementasi `GetCandle`)

Response:

- `200 OK`: data candle

### GET /candles/range

Ambil candle dalam rentang waktu tertentu.

Query params (contoh):

- `from` (timestamp)
- `to` (timestamp)
- `interval` (string)

Response:

- `200 OK`: list candle range

## Streaming

### GET /sse/candles

Server-Sent Events stream untuk update candle real-time.

Headers:

- `Accept: text/event-stream`

Response:

- Stream event candle secara kontinu

### GET /sse/ping

Endpoint health-check untuk SSE.

Response:

- `200 OK`: pong/status aktif

## WebSocket

### GET /ws/market

WebSocket endpoint untuk market stream.

Protocol:

- WebSocket

Catatan:

- Endpoint ini menggunakan auth JWT via handler websocket (sesuai implementasi server).

## Protected Profile

Prefix: `/profile`  
Middleware:

- JWT required

### GET /profile

Ambil profile user yang sedang login.

Headers:

- `Authorization: Bearer <token>`

Response:

- `200 OK`: data profil user
- `401 Unauthorized`: token tidak ada/invalid

## Error Format (Umum)

Format error mengikuti implementasi handler, umumnya:

```json
{
  "message": "error message"
}
```
