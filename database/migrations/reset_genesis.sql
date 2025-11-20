-- Script to reset and create proper genesis block
-- WARNING: This will delete all blockchain data!

-- Delete all data (in order to respect foreign keys)
DELETE FROM block_transactions;
DELETE FROM blocks;
DELETE FROM ledger;
UPDATE transactions SET status = 'PENDING';

-- Reset auto increment
ALTER TABLE blocks AUTO_INCREMENT = 1;

-- Create genesis block with proper hash
-- Hash for genesis block with no transactions: SHA256("0") = 5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9
INSERT INTO blocks (block_number, previous_hash, current_hash) 
VALUES (1, '0', '5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9');

-- Verify
SELECT * FROM blocks;
