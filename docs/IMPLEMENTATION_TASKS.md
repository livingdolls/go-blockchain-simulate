# Backend Implementation Tasks

**Last Updated**: 26 Februari 2026

---

## 📊 Status Overview

| Priority  | Total  | Completed | In Progress | Not Started |
| --------- | ------ | --------- | ----------- | ----------- |
| 🔴 HIGH   | 9      | 0         | 0           | 9           |
| 🟡 MEDIUM | 6      | 0         | 0           | 6           |
| 🟢 LOW    | 5      | 0         | 0           | 5           |
| **TOTAL** | **20** | **0**     | **0**       | **20**      |

---

## 🔴 PHASE 1: CRITICAL FEATURES (Sprint 1-2)

### Task 1.1: Transaction History & Status Endpoints

**Priority**: 🔴 HIGH  
**Estimated**: 2-3 days  
**Components**: Handler, Service, Repository, DTOs

- [*] **Handler**: Create `TransactionHistoryHandler`
  - [*] `GetPendingTransactions(address)` - GET /transaction/pending/:address
  - [*] `GetTransactionHistory(address)` - GET /transaction/history/:address
  - [*] `GetConfirmedTransactions(address)` - GET /transaction/confirmed/:address
  - [*] `GetTransactionStatus(id)` - GET /transaction/status/:id

- [*] **Service**: Enhance `TransactionService`
  - [*] Method: `GetPendingTransactionsByAddress(ctx, address)`
  - [*] Method: `GetTransactionHistory(ctx, address, limit, offset)`
  - [*] Method: `GetTransactionStatus(ctx, txID)`

- [*] **Repository**: Add methods to `TransactionRepository`
  - [*] `GetPendingByAddress(ctx, address, limit, offset)`
  - [*] `GetByAddressWithStatus(ctx, address, status, limit, offset)`
  - [*] `GetWithDetails(ctx, txID)`

- [*] **DTOs**: Create response DTOs
  - [*] `TransactionHistoryDTO`
  - [*] `TransactionStatusDTO`
  - [*] `PendingTransactionDTO`

---

### Task 1.2: Block Explorer Enhancement

**Priority**: 🔴 HIGH  
**Estimated**: 2-3 days  
**Components**: Handler, Service, Repository, DTOs

- [ ] **Handler**: Enhance `BlockHandler`
  - [ ] `GetLatestBlock()` - GET /blocks/latest
  - [*] `GetBlockTransactions(number)` - GET /blocks/:number/transactions
  - [ ] `GetBlockDetailedInfo(number)` - GET /blocks/:number/details
  - [ ] `GetBlockStats()` - GET /blocks/stats
  - [*] `SearchBlocks(query)` - GET /blocks/search
  - [*] `GetBlocksInRange(from, to)` - GET /blocks/range/:from/:to

- [ ] **Service**: Create `BlockExplorerService`
  - [ ] Method: `GetLatestBlock(ctx)`
  - [*] Method: `GetBlockWithTransactions(ctx, blockNumber)`
  - [ ] Method: `GetBlockStats(ctx)`
  - [*] Method: `SearchBlocks(ctx, query)`
  - [ ] Method: `GetDifficultyHistory(ctx)`
  - [ ] Method: `CalculateHashRate(ctx, lastN int)`

- [ ] **Repository**: Add methods to `BlockRepository`
  - [ ] `GetLatest(ctx)`
  - [*] `GetTransactionsByBlockNumber(ctx, number, limit, offset)`
  - [*] `SearchByHash(ctx, hash)`
  - [ ] `SearchByMiner(ctx, address, limit, offset)`
  - [*] `GetRangeStats(ctx, from, to)`

- [ ] **Routes**: Add to routes.go
  ```go
  blockGroup.GET("/latest", a.BlockHandler.GetLatestBlock)
  blockGroup.GET("/:number/transactions", a.BlockHandler.GetBlockTransactions)
  blockGroup.GET("/:number/details", a.BlockHandler.GetBlockDetailedInfo)
  blockGroup.GET("/stats", a.BlockHandler.GetBlockStats)
  blockGroup.GET("/search", a.BlockHandler.SearchBlocks)
  blockGroup.GET("/range/:from/:to", a.BlockHandler.GetBlocksInRange)
  ```

---

### Task 1.3: Admin Dashboard Routes & Handlers

**Priority**: 🔴 HIGH  
**Estimated**: 2-3 days  
**Components**: Handler, Service, DTOs, Middleware

- [ ] **Middleware**: Create Admin Role Check
  - [ ] `AdminMiddleware()` - Check admin role from JWT
  - [ ] Add to handlers

- [ ] **Handler**: Create `AdminHandler`
  - [ ] `GetSystemStats()` - GET /admin/stats
  - [ ] `GetNetworkStats()` - GET /admin/network-stats
  - [ ] `GetMiningStats()` - GET /admin/mining-stats
  - [ ] `HealthCheck()` - GET /admin/health
  - [ ] `StartMining()` - POST /admin/mining/start
  - [ ] `StopMining()` - POST /admin/mining/stop
  - [ ] `GetMiningStatus()` - GET /admin/mining/status
  - [ ] `VerifyBlockchain()` - POST /admin/verify-blockchain

- [ ] **Service**: Create `AdminService`
  - [ ] Method: `GetSystemStats(ctx)`
  - [ ] Method: `GetNetworkMetrics(ctx)`
  - [ ] Method: `GetMiningMetrics(ctx)`
  - [ ] Method: `VerifyBlockchainIntegrity(ctx)`

- [ ] **DTOs**: Create response DTOs
  - [ ] `SystemStatsDTO`
  - [ ] `NetworkStatsDTO`
  - [ ] `MiningStatsDTO`
  - [ ] `HealthCheckDTO`

- [ ] **Routes**: Add protected routes to routes.go
  ```go
  admin := r.Group("/admin")
  admin.Use(handler.JWTMiddleware(a.JWT), handler.AdminMiddleware())
  {
    admin.GET("/stats", a.AdminHandler.GetSystemStats)
    admin.GET("/network-stats", a.AdminHandler.GetNetworkStats)
    admin.GET("/mining-stats", a.AdminHandler.GetMiningStats)
    admin.GET("/health", a.AdminHandler.HealthCheck)
    admin.POST("/mining/start", a.AdminHandler.StartMining)
    admin.POST("/mining/stop", a.AdminHandler.StopMining)
    admin.GET("/mining/status", a.AdminHandler.GetMiningStatus)
    admin.POST("/verify-blockchain", a.AdminHandler.VerifyBlockchain)
  }
  ```

---

### Task 1.4: User & Address Profile Routes

**Priority**: 🔴 HIGH  
**Estimated**: 2-3 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Create `AddressProfileHandler`
  - [ ] `GetPublicProfile(address)` - GET /user/:address
  - [ ] `GetAssetHoldings(address)` - GET /user/:address/assets
  - [ ] `GetTradingStats(address)` - GET /user/:address/trading-stats
  - [ ] `GetLeaderboard()` - GET /user/leaderboard

- [ ] **Handler**: Enhance `ProfileHandler` (Protected)
  - [ ] `GetMyProfile()` - GET /profile/me
  - [ ] `GetMyWallets()` - GET /profile/wallets
  - [ ] `UpdateProfile(data)` - PUT /profile/me

- [ ] **Service**: Create `ProfileService`
  - [ ] Method: `GetUserPublicProfile(ctx, address)`
  - [ ] Method: `GetUserAssets(ctx, address)`
  - [ ] Method: `GetTradingStatistics(ctx, address)`
  - [ ] Method: `GetLeaderboard(ctx, limit, offset)`

- [ ] **Service**: Enhance `UserService`
  - [ ] Method: `GetMyDetailedProfile(ctx, userID)`
  - [ ] Method: `GetMyWallets(ctx, userID)`

- [ ] **DTOs**: Create response DTOs
  - [ ] `PublicProfileDTO`
  - [ ] `UserAssetsDTO`
  - [ ] `TradingStatsDTO`
  - [ ] `LeaderboardEntryDTO`
  - [ ] `MyProfileDTO`

- [ ] **Routes**: Add to routes.go

  ```go
  // Public routes
  r.GET("/user/:address", a.AddressProfileHandler.GetPublicProfile)
  r.GET("/user/:address/assets", a.AddressProfileHandler.GetAssetHoldings)
  r.GET("/user/:address/trading-stats", a.AddressProfileHandler.GetTradingStats)
  r.GET("/user/leaderboard", a.AddressProfileHandler.GetLeaderboard)

  // Protected routes
  protected.GET("/me", a.ProfileHandler.GetMyProfile)
  protected.GET("/wallets", a.ProfileHandler.GetMyWallets)
  protected.PUT("/me", a.ProfileHandler.UpdateProfile)
  ```

---

### Task 1.5: Market Data & History Routes

**Priority**: 🔴 HIGH  
**Estimated**: 2 days  
**Components**: Handler, Service, Repository, DTOs

- [ ] **Handler**: Create `MarketAnalyticsHandler`
  - [ ] `GetPriceHistory()` - GET /market/price/history
  - [ ] `GetVolumeHistory()` - GET /market/volume/history
  - [ ] `GetMarketStats()` - GET /market/stats
  - [ ] `GetOrderBook()` - GET /market/orderbook

- [ ] **Service**: Enhance `MarketService`
  - [ ] Method: `GetPriceHistory(ctx, from, to, interval)`
  - [ ] Method: `GetVolumeHistory(ctx, from, to)`
  - [ ] Method: `GetMarketStatistics(ctx)`
  - [ ] Method: `CalculateMarketMetrics(ctx)`

- [ ] **Repository**: Add methods to `MarketRepository`
  - [ ] `GetHistoricalPrices(ctx, from, to, limit)`
  - [ ] `GetVolumeData(ctx, from, to, limit)`

- [ ] **DTOs**: Create response DTOs
  - [ ] `PriceHistoryDTO`
  - [ ] `VolumeHistoryDTO`
  - [ ] `MarketStatsDTO`

- [ ] **Routes**: Add to routes.go
  ```go
  market := r.Group("/market")
  {
    market.GET("/price/history", a.MarketAnalyticsHandler.GetPriceHistory)
    market.GET("/volume/history", a.MarketAnalyticsHandler.GetVolumeHistory)
    market.GET("/stats", a.MarketAnalyticsHandler.GetMarketStats)
    market.GET("/orderbook", a.MarketAnalyticsHandler.GetOrderBook)
  }
  ```

---

## 🟡 PHASE 2: IMPORTANT FEATURES (Sprint 3-4)

### Task 2.1: Enhanced Wallet/Balance Routes

**Priority**: 🟡 MEDIUM  
**Estimated**: 2 days  
**Components**: Handler, Service, Repository, DTOs

- [ ] **Handler**: Enhance `BalanceHandler` or create `WalletExplorerHandler`
  - [ ] `GetWalletHistory(address)` - GET /wallet/:address/history
  - [ ] `GetWalletTransactions(address)` - GET /wallet/:address/transactions
  - [ ] `GetBalanceHistory(address)` - GET /balance/:address/history
  - [ ] `GetNetworkBalanceStats()` - GET /balance/stats

- [ ] **Service**: Create `WalletExplorerService`
  - [ ] Method: `GetWalletHistory(ctx, address, limit)`
  - [ ] Method: `GetWalletTransactions(ctx, address, limit)`
  - [ ] Method: `GetBalanceHistory(ctx, address)`

- [ ] **Repository**: Add methods to related repositories
  - [ ] `GetBalanceHistory(ctx, address, limit)`
  - [ ] `GetNetworkStats(ctx)`

- [ ] **DTOs**: Create response DTOs
  - [ ] `WalletHistoryDTO`
  - [ ] `BalanceHistoryDTO`
  - [ ] `NetworkBalanceStatsDTO`

---

### Task 2.2: Search & Filter Functionality

**Priority**: 🟡 MEDIUM  
**Estimated**: 2 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Create `SearchHandler`
  - [ ] `UnifiedSearch(query)` - GET /search?q=query
  - [ ] `SearchBlocks(query)` - GET /search/blocks?q=query
  - [ ] `SearchTransactions(query)` - GET /search/transactions?q=query
  - [ ] `SearchAddresses(query)` - GET /search/addresses?q=query

- [ ] **Service**: Create `SearchService`
  - [ ] Method: `SearchAcrossAll(ctx, query, limit)`
  - [ ] Method: `SearchBlocks(ctx, query)`
  - [ ] Method: `SearchTransactions(ctx, query)`
  - [ ] Method: `SearchAddresses(ctx, query)`

- [ ] **DTOs**: Create response DTOs
  - [ ] `SearchResultDTO`
  - [ ] `SearchResultsAggregateDTO`

---

### Task 2.3: Order Management Routes (Advanced Trading)

**Priority**: 🟡 MEDIUM  
**Estimated**: 2-3 days  
**Components**: Handler, Service, Repository, DTOs

- [ ] **Handler**: Create `OrderHandler`
  - [ ] `PlaceOrder(data)` - POST /order/place
  - [ ] `GetPendingOrders(address)` - GET /order/pending/:address
  - [ ] `GetOrderHistory(address)` - GET /order/history/:address
  - [ ] `CancelOrder(id)` - DELETE /order/:id
  - [ ] `GetOrderDetails(id)` - GET /order/:id

- [ ] **Service**: Create `OrderService`
  - [ ] Method: `PlaceOrder(ctx, order)`
  - [ ] Method: `GetPendingOrders(ctx, address)`
  - [ ] Method: `GetOrderHistory(ctx, address)`
  - [ ] Method: `CancelOrder(ctx, orderID)`

- [ ] **Repository**: Create `OrderRepository` if not exists
  - [ ] CRUD methods for orders

- [ ] **DTOs**: Create request/response DTOs
  - [ ] `PlaceOrderRequestDTO`
  - [ ] `OrderResponseDTO`
  - [ ] `OrderHistoryDTO`

---

### Task 2.4: Ledger & Accounting Routes

**Priority**: 🟡 MEDIUM  
**Estimated**: 2 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Create `LedgerHandler`
  - [ ] `GetUserLedger(address)` - GET /ledger/:address
  - [ ] `ExportLedger(address)` - GET /ledger/:address/export
  - [ ] `GetBlockLedger(number)` - GET /ledger/block/:number

- [ ] **Service**: Enhance `LedgerPublisher` or create `LedgerService`
  - [ ] Method: `GetUserLedger(ctx, address, limit, offset)`
  - [ ] Method: `ExportLedger(ctx, address, format)` - CSV/JSON
  - [ ] Method: `GetBlockLedger(ctx, blockNumber)`

- [ ] **Repository**: Ensure `LedgerRepository` has all needed methods

- [ ] **DTOs**: Create response DTOs
  - [ ] `LedgerEntryDTO` (if not exists)
  - [ ] `UserLedgerDTO`

---

### Task 2.5: Transaction Statistics Routes

**Priority**: 🟡 MEDIUM  
**Estimated**: 1-2 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Enhance `TransactionHandler`
  - [ ] `GetUserTransactionStats(address)` - GET /transaction/stats/:address
  - [ ] `GetNetworkTransactionStats()` - GET /transaction/stats

- [ ] **Service**: Enhance `TransactionService`
  - [ ] Method: `GetUserTransactionStatistics(ctx, address)`
  - [ ] Method: `GetNetworkTransactionStatistics(ctx)`

- [ ] **DTOs**: Create response DTOs
  - [ ] `TransactionStatsDTO`
  - [ ] `NetworkTransactionStatsDTO`

---

## 🟢 PHASE 3: NICE TO HAVE FEATURES (Sprint 5+)

### Task 3.1: Advanced Analytics Dashboard

**Priority**: 🟢 LOW  
**Estimated**: 3-4 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Create `AnalyticsHandler`
  - [ ] `GetDashboardData()` - GET /analytics/dashboard
  - [ ] `GetMarketAnalytics()` - GET /analytics/market
  - [ ] `GetUserAnalytics()` - GET /analytics/users
  - [ ] `ExportAnalytics(type)` - GET /analytics/export/:type

- [ ] **Service**: Create `AnalyticsService`
  - [ ] Aggregate data from multiple sources
  - [ ] Calculate metrics and KPIs

---

### Task 3.2: Notification/Event Routes

**Priority**: 🟢 LOW  
**Estimated**: 2-3 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Create `NotificationHandler`
  - [ ] `GetUserEvents()` - GET /events/user/:address
  - [ ] `SubscribeToEvents()` - POST /events/subscribe
  - [ ] `UnsubscribeFromEvents()` - POST /events/unsubscribe

---

### Task 3.3: Difficulty & Hash Rate History

**Priority**: 🟢 LOW  
**Estimated**: 1-2 days  
**Components**: Handler, Service, DTOs

- [ ] **Handler**: Enhance `BlockHandler`
  - [ ] `GetDifficultyHistory()` - GET /blocks/difficulty/history
  - [ ] `GetHashRateHistory()` - GET /blocks/hash-rate/history

---

### Task 3.4: Advanced Filtering & Pagination

**Priority**: 🟢 LOW  
**Estimated**: 2-3 days  
**Components**: DTOs, Middleware, Repository

- [ ] Implement pagination for all list endpoints
- [ ] Add filtering capabilities (by date, type, status, etc.)
- [ ] Implement sorting options

---

### Task 3.5: Data Export Features

**Priority**: 🟢 LOW  
**Estimated**: 2 days  
**Components**: Service, Handler

- [ ] CSV export for transactions
- [ ] CSV export for blocks
- [ ] JSON export for ledger
- [ ] PDF export for reports

---

## 📝 Implementation Notes

### Code Structure Template

```
For each task, create:
1. DTOs (if new) → app/dto/<feature>.go
2. Handler → app/handler/<feature>.go
3. Service (if new) → app/services/<feature>.go
4. Repository (if new) → app/repository/<feature>.go
5. Update routes.go with new endpoints
```

### Testing Requirements

- [ ] Unit tests for each handler
- [ ] Unit tests for each service
- [ ] Integration tests for each route
- [ ] Mock repositories for testing

### Response Format Standard

```go
{
  "success": true,
  "status_code": 200,
  "message": "Success",
  "data": { ... },
  "pagination": {
    "limit": 10,
    "offset": 0,
    "total": 100
  },
  "error": null
}
```

### Error Handling Standard

All endpoints must return standardized error responses

```go
{
  "success": false,
  "status_code": 400,
  "message": "Failed to fetch data",
  "data": null,
  "error": {
    "code": "INVALID_REQUEST",
    "details": "..."
  }
}
```

---

## 🔗 Related Documentation

- [Generate Block Flow](generate_block_flow.md)
- [Register Flow](register_flow.md)
- [Transaction Flow](transaction_flow.md)

---

## 📅 Timeline Estimate

- **Phase 1** (Critical): 2-3 weeks (9 tasks × 2-3 days each)
- **Phase 2** (Important): 2 weeks (6 tasks × 1-2 days each)
- **Phase 3** (Nice to have): 2-3 weeks (5 tasks × 1-3 days each)

**Total**: 6-8 weeks for full implementation

---

## ✅ Completion Checklist

Use this as progress tracker:

- [ ] Phase 1.1 - Transaction History
- [ ] Phase 1.2 - Block Explorer
- [ ] Phase 1.3 - Admin Dashboard
- [ ] Phase 1.4 - User Profiles
- [ ] Phase 1.5 - Market Analytics
- [ ] Phase 2.1 - Wallet Explorer
- [ ] Phase 2.2 - Search
- [ ] Phase 2.3 - Order Management
- [ ] Phase 2.4 - Ledger
- [ ] Phase 2.5 - Transaction Stats
- [ ] Phase 3.1 - Analytics
- [ ] Phase 3.2 - Notifications
- [ ] Phase 3.3 - Difficulty History
- [ ] Phase 3.4 - Advanced Filtering
- [ ] Phase 3.5 - Data Export
