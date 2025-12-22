CREATE TABLE transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    signature TEXT NOT NULL,
    status ENUM('PENDING', 'SUCCESS', 'FAILED', 'CONFIRMED') DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_from_address (from_address),
    INDEX idx_to_address (to_address)
);


-- Add Fee Column to Transactions
ALTER TABLE transactions
ADD COLUMN fee DECIMAL(20, 8) NOT NULL DEFAULT 0.00100000 AFTER amount;

-- Update existing transactions to set default fee
UPDATE transactions SET fee = 0.00100000 WHERE fee = 0 OR fee IS NULL;

-- Add index for based query
CREATE INDEX idx_transactions_fee ON transactions(fee DESC);

ALTER TABLE transactions MODIFY COLUMN amount DECIMAL(20, 8) NOT NULL;
ALTER TABLE transactions MODIFY COLUMN fee DECIMAL(20, 8) NOT NULL;

-- ADD type column to transactions
ALTER TABLE transactions
ADD COLUMN type ENUM('TRANSFER', 'BUY', 'SELL') NOT NULL DEFAULT 'TRANSFER' AFTER fee;