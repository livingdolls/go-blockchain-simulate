### Task Message Broker

Queue: block.generation
Purpose: Decouple block mining dari HTTP request
Benefits:

- Mining tidak blocking user API
- Distributed mining workers
- Retry mechanism otomatis jika mining gagal

[DONE]
Queue: transaction.pending
Queue: transaction.confirmed
Purpose: Async processing untuk transaksi
Benefits:

- Send transaction API return immediately
- Background validation & processing
- Prevent double-spending dengan queue serialization
- Transaction confirmation updates via WebSocket

Queue: market.pricing
Queue: market.volume.updates
Purpose: Async price updates berdasarkan block data
Benefits:

- Market updates tidak delay block creation
- Multiple consumers (WebSocket, storage)
- Historical price tracking

Queue: ledger.entries
Purpose: Audit trail & balance reconciliation
Benefits:

- Decoupled dari block transaction
- Guaranteed persistence of ledger entries
- Multi-consumer for analytics

Queue: notifications
Routing Keys:

- notification.user.{address}
- notification.transaction.{tx_id}
  Purpose: Replace current WebSocket-only notifications
  Benefits:
- Persistent notification queue
- Offline message queue
- Multiple notification channels (email, SMS)

Queue: reward.calculation
Queue: reward.distribution
Purpose: Async reward processing for users & miner
Benefits:

- Heavy calculation tidak block main flow
- Scheduled reward distribution

### Architekture Pattern

HTTP Request
↓
Handler → Service → RabbitMQ (Immediate Return)
↓
Worker Pool
↓
Database/WebSocket
↓
Real-time Updates
