CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address VARCHAR(255) UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    balance DECIMAL(18,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Create special FEE_POOL account
INSERT INTO users(name, address, balance, public_key, private_key)
VALUES ('FEE_POOL', 'FEE_POOL', 0, 'SYSTEM_ACCOUNT', 'SYSTEM_ACCOUNT')
ON DUPLICATE KEY UPDATE balance = balance;

-- Create Minner Account
INSERT INTO users (name, address, balance, public_key, private_key)
VALUES ('MINER_ACCOUNT', 'MINER_ACCOUNT', 0.00000000, 'SYSTEM_MINER', 'SYSTEM_MINER')
ON DUPLICATE KEY UPDATE address = address;

ALTER TABLE users MODIFY COLUMN balance DECIMAL(20, 8) NOT NULL;

ALTER TABLE users DROP COLUMN private_key;

-- Migrate data dari users.balance ke user_wallets
INSERT INTO user_wallets (user_address, yte_balance, last_transaction_at)
SELECT address, balance, created_at
FROM users
ON DUPLICATE KEY UPDATE yte_balance = VALUES(yte_balance);

-- Drop kolom balance dari tabel users
ALTER TABLE users DROP COLUMN balance;