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
INSERT INTO users(address, balance, public_key, private_key)
VALUES ('FEE_POOL', 0, 'SYSTEM_ACCOUNT', 'SYSTEM_ACCOUNT')
ON DUPLICATE KEY UPDATE balance = balance;

ALTER TABLE users MODIFY COLUMN balance DECIMAL(20, 8) NOT NULL;