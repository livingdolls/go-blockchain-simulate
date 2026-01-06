CREATE TABLE user_balances (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) UNIQUE NOT NULL,
    usd_balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- Saldo USD (2 decimal)
    locked_balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- Untuk pending orders
    available_balance DECIMAL(20, 2) GENERATED ALWAYS AS (usd_balance - locked_balance) STORED,
    total_deposited DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- Total USD yang pernah di-deposit
    total_withdrawn DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- Total USD yang pernah di-withdraw
    total_traded DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- Total nilai trading
    last_transaction_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_last_transaction (last_transaction_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE,

    -- Constraint: locked tidak boleh melebihi balance
    CHECK (locked_balance <= usd_balance),
    CHECK (usd_balance >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE balance_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) NOT NULL,
    order_id BIGINT NULL, -- Untuk trading orders
    change_type ENUM(
        'DEPOSIT',       -- Setor USD
        'WITHDRAWAL',    -- Tarik USD
        'BUY_ORDER',     -- Beli YTE (debit USD)
        'SELL_ORDER',    -- Jual YTE (credit USD)
        'CANCEL_ORDER',  -- Batalin order (refund USD)
        'FEE',           -- Biaya trading
        'LOCK',          -- Lock untuk pending order
        'UNLOCK'         -- Unlock dari order
    ) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL, -- Positive = tambah, negative = kurang
    balance_before DECIMAL(20, 2) NOT NULL,
    balance_after DECIMAL(20, 2) NOT NULL,
    locked_before DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    locked_after DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    reference_id VARCHAR(255) NULL, -- order_id, withdrawal_id, etc
    description VARCHAR(512) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_order_id (order_id),
    INDEX idx_change_type (change_type),
    INDEX idx_reference_id (reference_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE balance_locks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    lock_type ENUM('BUY_ORDER', 'SELL_ORDER', 'OTHER') NOT NULL,
    reference_id VARCHAR(255) NOT NULL, -- order_id
    status ENUM('ACTIVE', 'RELEASED', 'EXECUTED') NOT NULL DEFAULT 'ACTIVE',
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP NULL,

    INDEX idx_user_address (user_address),
    INDEX idx_reference_id (reference_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE,

    UNIQUE KEY unique_lock (reference_id, lock_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_wallets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) UNIQUE NOT NULL,
    yte_balance DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000, -- Saldo YTE (8 decimal)
    locked_balance DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000, -- Untuk pending sell orders
    available_balance DECIMAL(20, 8) GENERATED ALWAYS AS (yte_balance - locked_balance) STORED,
    total_received DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000, -- Total YTE diterima
    total_sent DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000, -- Total YTE dikirim
    last_transaction_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_last_transaction (last_transaction_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE,

    CHECK (locked_balance <= yte_balance),
    CHECK (yte_balance >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE wallet_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_address VARCHAR(255) NOT NULL,
    tx_id BIGINT NULL, -- Blockchain transaction
    order_id BIGINT NULL, -- Trading order
    change_type ENUM(
        'RECEIVE',      -- Terima dari transfer
        'SEND',         -- Kirim ke orang lain
        'BUY_ORDER',    -- Beli dari market
        'SELL_ORDER',   -- Jual ke market
        'MINING',       -- Reward mining
        'FEE_PAID',     -- Bayar fee blockchain
        'LOCK',         -- Lock untuk pending sell
        'UNLOCK'        -- Unlock dari pending sell
    ) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    balance_before DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    locked_before DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    locked_after DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000,
    reference_id VARCHAR(255) NULL,
    description VARCHAR(512) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_tx_id (tx_id),
    INDEX idx_order_id (order_id),
    INDEX idx_change_type (change_type),
    INDEX idx_reference_id (reference_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_address) REFERENCES users(address) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;