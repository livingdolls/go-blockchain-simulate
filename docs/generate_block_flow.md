┌─────────────────────────────────────────────────────────────────────┐
│                    GENERATE BLOCK FLOW                               │
└─────────────────────────────────────────────────────────────────────┘

START: POST /generate-block
   ↓
┌──────────────────────────────────────────────────────────────┐
│ PHASE 1: PRE-VALIDATION (Read-Only, No Database Locks)      │
└──────────────────────────────────────────────────────────────┘
   ↓
   ├─→ [1.1] Get Last Block (read-only)
   │         SELECT * FROM blocks ORDER BY block_number DESC LIMIT 1
   │         ↓
   │         lastBlock = Block #N
   │         prevHash = lastBlock.CurrentHash
   │
   ├─→ [1.2] Get Pending Transactions (read-only, max 100)
   │         SELECT * FROM transactions WHERE status = 'PENDING' LIMIT 100
   │         ↓
   │         pendingTxs = [Tx1, Tx2, Tx3, ...]
   │         ↓
   │         IF pendingTxs is empty → RETURN ERROR "No pending transactions"
   │
   ├─→ [1.3] Collect Unique Addresses
   │         uniqueAddresses = []
   │         FOR each tx in pendingTxs:
   │             add tx.FromAddress
   │             add tx.ToAddress
   │         ↓
   │         addresses = ["addr1", "addr2", "addr3", ...]
   │
   ├─→ [1.4] Get All Users (bulk query, read-only)
   │         SELECT * FROM users WHERE address IN (addresses)
   │         ↓
   │         users = {addr1: User1, addr2: User2, ...}
   │         ↓
   │         Cache users in memory (userCache)
   │
   ├─→ [1.5] Pre-Validate Balances (in-memory, no DB)
   │         balances = {addr1: 100, addr2: 50, ...}
   │         ↓
   │         FOR each tx in pendingTxs:
   │             IF balances[tx.FromAddress] < tx.Amount:
   │                 RETURN ERROR "Insufficient balance"
   │             
   │             balances[tx.FromAddress] -= tx.Amount
   │             balances[tx.ToAddress] += tx.Amount
   │         ↓
   │         ✅ All transactions valid
   │
┌──────────────────────────────────────────────────────────────┐
│ PHASE 2: MINING (Proof of Work - CPU Intensive)             │
└──────────────────────────────────────────────────────────────┘
   ↓
   ├─→ [2.1] Get All Blocks for Difficulty Calculation
   │         SELECT * FROM blocks ORDER BY block_number
   │         ↓
   │         allBlocks = [Block1, Block2, ..., BlockN]
   │
   ├─→ [2.2] Calculate Next Difficulty
   │         IF blocks count < 10:
   │             difficulty = 4 (default)
   │         ELSE:
   │             last10Blocks = allBlocks[-10:]
   │             actualTime = last10Blocks[-1].Timestamp - last10Blocks[0].Timestamp
   │             expectedTime = 10 seconds × 9 blocks = 90 seconds
   │             
   │             IF actualTime < expectedTime / 2:
   │                 difficulty += 1  (blocks too fast, increase)
   │             ELSE IF actualTime > expectedTime × 2:
   │                 difficulty -= 1  (blocks too slow, decrease)
   │             ELSE:
   │                 difficulty stays same
   │         ↓
   │         difficulty = 4
   │
   ├─→ [2.3] Calculate Merkle Root
   │         merkleTree = BuildMerkleTree(pendingTxs)
   │         merkleRoot = merkleTree.Root
   │         ↓
   │         merkleRoot = "a1b2c3d4e5f6..."
   │
   ├─→ [2.4] Start Mining (Proof of Work)
   │         target = "0000..." (difficulty leading zeros)
   │         nonce = 0
   │         startTime = now()
   │         
   │         LOOP (until valid hash found):
   │             data = blockNumber + prevHash + transactions + nonce + timestamp
   │             hash = SHA256(data)
   │             
   │             IF hash starts with target (e.g., "0000..."):
   │                 ✅ VALID HASH FOUND!
   │                 BREAK
   │             
   │             nonce += 1
   │             
   │             Every 100,000 attempts:
   │                 Print progress (attempts, time, hash rate)
   │             
   │             IF elapsed > 10 minutes:
   │                 RETURN ERROR "Mining timeout"
   │         ↓
   │         Result:
   │         - hash = "0000a1b2c3d4e5f6..."
   │         - nonce = 234567
   │         - duration = 8.5 seconds
   │         - hashRate = 27,600 H/s
   │
┌──────────────────────────────────────────────────────────────┐
│ PHASE 3: DATABASE WRITE (Short Transaction < 2 seconds)     │
└──────────────────────────────────────────────────────────────┘
   ↓
   ├─→ [3.1] Begin Transaction
   │         tx = BEGIN TRANSACTION
   │         defer tx.Rollback() (safety net)
   │
   ├─→ [3.2] Lock Last Block (verify no concurrent mining)
   │         SELECT * FROM blocks 
   │         WHERE block_number = (SELECT MAX(block_number) FROM blocks)
   │         FOR UPDATE
   │         ↓
   │         lastBlockLocked = Block #N
   │         
   │         IF lastBlockLocked.BlockNumber != lastBlock.BlockNumber:
   │             ROLLBACK
   │             RETURN ERROR "New block created while mining, retry"
   │
   ├─→ [3.3] Create New Block
   │         newBlock = {
   │             BlockNumber: N + 1,
   │             PreviousHash: prevHash,
   │             CurrentHash: hash (from mining),
   │             Nonce: nonce,
   │             Difficulty: difficulty,
   │             Timestamp: now(),
   │             MerkleRoot: merkleRoot
   │         }
   │         
   │         INSERT INTO blocks (...) VALUES (...)
   │         ↓
   │         blockID = 123
   │
   ├─→ [3.4] Lock All Users (bulk lock)
   │         SELECT * FROM users 
   │         WHERE address IN (addresses)
   │         FOR UPDATE
   │         ↓
   │         🔒 All users locked
   │
   ├─→ [3.5] Prepare Bulk Operations (in-memory)
   │         ledgerEntries = []
   │         txIDs = []
   │         currentBalances = {addr1: 100, addr2: 50, ...}
   │         
   │         FOR each tx in pendingTxs:
   │             // Update balances
   │             currentBalances[tx.FromAddress] -= tx.Amount
   │             currentBalances[tx.ToAddress] += tx.Amount
   │             
   │             // Prepare ledger entries
   │             ledgerEntries.append({
   │                 TxID: tx.ID,
   │                 Address: tx.FromAddress,
   │                 Amount: -tx.Amount,
   │                 BalanceAfter: currentBalances[tx.FromAddress]
   │             })
   │             ledgerEntries.append({
   │                 TxID: tx.ID,
   │                 Address: tx.ToAddress,
   │                 Amount: +tx.Amount,
   │                 BalanceAfter: currentBalances[tx.ToAddress]
   │             })
   │             
   │             txIDs.append(tx.ID)
   │
   ├─→ [3.6] Bulk Insert Ledger Entries (1 query)
   │         INSERT INTO ledger (tx_id, address, amount, balance_after)
   │         VALUES 
   │             (1, 'addr1', -10, 90),
   │             (1, 'addr2', +10, 60),
   │             (2, 'addr1', -5, 85),
   │             (2, 'addr3', +5, 5),
   │             ...
   │         ↓
   │         ✅ 200 ledger entries inserted (100 txs × 2 entries)
   │
   ├─→ [3.7] Bulk Link Transactions to Block (1 query)
   │         INSERT INTO block_transactions (block_id, transaction_id)
   │         VALUES 
   │             (123, 1),
   │             (123, 2),
   │             (123, 3),
   │             ...
   │         ↓
   │         ✅ 100 block-transaction links created
   │
   ├─→ [3.8] Bulk Mark Transactions as Confirmed (1 query)
   │         UPDATE transactions 
   │         SET status = 'CONFIRMED' 
   │         WHERE id IN (1, 2, 3, ...)
   │         ↓
   │         ✅ 100 transactions confirmed
   │
   ├─→ [3.9] Bulk Update User Balances (1 query)
   │         UPDATE users 
   │         SET balance = CASE address
   │             WHEN 'addr1' THEN 85
   │             WHEN 'addr2' THEN 60
   │             WHEN 'addr3' THEN 5
   │             ...
   │         END
   │         WHERE address IN ('addr1', 'addr2', 'addr3', ...)
   │         ↓
   │         ✅ 50 user balances updated
   │
   ├─→ [3.10] Commit Transaction
   │         COMMIT
   │         ↓
   │         🔓 All locks released
   │         ✅ Block added to blockchain!
   │
┌──────────────────────────────────────────────────────────────┐
│ PHASE 4: RESPONSE                                            │
└──────────────────────────────────────────────────────────────┘
   ↓
   └─→ [4.1] Return Success Response
         {
             "message": "Block generated successfully",
             "block": {
                 "id": 123,
                 "block_number": N+1,
                 "hash": "0000a1b2c3d4...",
                 "nonce": 234567,
                 "difficulty": 4,
                 "merkle_root": "a1b2c3...",
                 "transactions": 100,
                 "mining_time": "8.5s"
             }
         }

END