┌─────────────────────────────────────────────────────────────────────┐
│ SEND TRANSACTION FLOW │
└─────────────────────────────────────────────────────────────────────┘

START: POST /send
↓
Request Body: {
"from": "address1",
"to": "address2",
"amount": 10.5,
"private_key": "privatekey123"
}
↓
┌──────────────────────────────────────────────────────────────┐
│ STEP 1: REQUEST VALIDATION │
└──────────────────────────────────────────────────────────────┘
↓
├─→ [1.1] Validate Request Format
│ IF from is empty:
│ RETURN ERROR "from address required"
│ IF to is empty:
│ RETURN ERROR "to address required"
│ IF amount <= 0:
│ RETURN ERROR "amount must be positive"
│ IF private_key is empty:
│ RETURN ERROR "private_key required"
│ ↓
│ ✅ Request format valid
│
├─→ [1.2] Validate Not Self-Transfer
│ IF from == to:
│ RETURN ERROR "Cannot send to yourself"
│ ↓
│ ✅ Different addresses
│
┌──────────────────────────────────────────────────────────────┐
│ STEP 2: VALIDATE SENDER (FROM ADDRESS) │
└──────────────────────────────────────────────────────────────┘
↓
├─→ [2.1] Get Sender from Database
│ SELECT _ FROM users WHERE address = 'address1'
│ ↓
│ IF NOT FOUND:
│ RETURN ERROR "Sender not found"
│ ↓
│ sender = {
│ Address: "address1",
│ Balance: 100.0,
│ PublicKey: "pubkey123",
│ CreatedAt: "2024-11-20"
│ }
│
├─→ [2.2] Verify Private Key Ownership
│ expectedPubKey = DerivePublicKey(private_key)
│ actualPubKey = sender.PublicKey
│  
 │ IF expectedPubKey != actualPubKey:
│ RETURN ERROR "Invalid private key for sender"
│ ↓
│ ✅ Private key verified
│
┌──────────────────────────────────────────────────────────────┐
│ STEP 3: VALIDATE RECEIVER (TO ADDRESS) │
└──────────────────────────────────────────────────────────────┘
↓
├─→ [3.1] Get Receiver from Database
│ SELECT _ FROM users WHERE address = 'address2'
│ ↓
│ IF NOT FOUND:
│ RETURN ERROR "Receiver not found"
│ ↓
│ receiver = {
│ Address: "address2",
│ Balance: 50.0,
│ PublicKey: "pubkey456",
│ CreatedAt: "2024-11-19"
│ }
│ ↓
│ ✅ Receiver exists
│
┌──────────────────────────────────────────────────────────────┐
│ STEP 4: VALIDATE BALANCE (WITH FEE - if implemented) │
└──────────────────────────────────────────────────────────────┘
↓
├─→ [4.1] Calculate Total Required
│ amount = 10.5
│ fee = 0.001 (default, or calculated)
│ totalRequired = amount + fee = 10.501
│
├─→ [4.2] Check Current Balance
│ currentBalance = sender.Balance = 100.0
│  
 │ IF currentBalance < totalRequired:
│ RETURN ERROR {
│ "error": "Insufficient balance",
│ "required": 10.501,
│ "available": 100.0,
│ "shortage": 0 (in this case OK)
│ }
│ ↓
│ ✅ Balance sufficient: 100.0 >= 10.501
│
├─→ [4.3] Check Pending Transactions (Optional)
│ SELECT SUM(amount + fee)
│ FROM transactions
│ WHERE from_address = 'address1'
│ AND status = 'PENDING'
│ ↓
│ pendingAmount = 5.0
│ availableBalance = currentBalance - pendingAmount
│ availableBalance = 100.0 - 5.0 = 95.0
│  
 │ IF availableBalance < totalRequired:
│ RETURN ERROR {
│ "error": "Insufficient available balance",
│ "current_balance": 100.0,
│ "pending_transactions": 5.0,
│ "available": 95.0,
│ "required": 10.501
│ }
│ ↓
│ ✅ Available balance sufficient
│
┌──────────────────────────────────────────────────────────────┐
│ STEP 5: GENERATE SIGNATURE (Cryptographic Proof) │
└──────────────────────────────────────────────────────────────┘
↓
├─→ [5.1] Create Transaction Data String
│ txData = from + to + amount + timestamp
│ txData = "address1" + "address2" + "10.5" + "1732186543"
│
├─→ [5.2] Sign with Private Key
│ signature = SignFake(private_key, to, amount)
│ // In production: ECDSA signature
│ // signature = ECDSA.sign(private_key, txData)
│ ↓
│ signature = "a1b2c3d4e5f6..." (64 char hex string)
│ ↓
│ ✅ Signature generated
│
┌──────────────────────────────────────────────────────────────┐
│ STEP 6: CREATE TRANSACTION (Database Insert) │
└──────────────────────────────────────────────────────────────┘
↓
├─→ [6.1] Prepare Transaction Object
│ transaction = {
│ FromAddress: "address1",
│ ToAddress: "address2",
│ Amount: 10.5,
│ Fee: 0.001,
│ Signature: "a1b2c3d4e5f6...",
│ Status: "PENDING",
│ CreatedAt: now()
│ }
│
├─→ [6.2] Insert into Database
│ INSERT INTO transactions
│ (from_address, to_address, amount, fee, signature, status)
│ VALUES
│ ('address1', 'address2', 10.5, 0.001, 'a1b2c3...', 'PENDING')
│ ↓
│ txID = 123
│ transaction.ID = 123
│ ↓
│ ✅ Transaction saved to database
│
┌──────────────────────────────────────────────────────────────┐
│ STEP 7: RESPONSE │
└──────────────────────────────────────────────────────────────┘
↓
└─→ [7.1] Return Success Response
{
"message": "Transaction created successfully",
"transaction": {
"id": 123,
"from_address": "address1",
"to_address": "address2",
"amount": 10.5,
"fee": 0.001,
"signature": "a1b2c3d4e5f6...",
"status": "PENDING",
"created_at": "2024-11-21 10:30:45"
},
"note": "Transaction is pending. It will be confirmed when included in a block."
}

END

┌──────────────────────────────────────────────────────────────┐
│ TRANSACTION NOW IN PENDING POOL │
│ Waiting to be picked up by miner... │
└──────────────────────────────────────────────────────────────┘
