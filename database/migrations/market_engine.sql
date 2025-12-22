CREATE TABLE market_engine (
    id INT PRIMARY KEY CHECK (id = 1),
    price DECIMAL(20, 8) NOT NULL,
    liquidity DECIMAL(20, 8) NOT NULL,
    last_block BIGINT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE market_ticks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    block_id BIGINT NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    buy_volume DECIMAL(20, 8) NOT NULL,
    sell_volume DECIMAL(20, 8) NOT NULL,
    tx_count INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_tick_block (block_id),
    FOREIGN KEY (block_id) REFERENCES blocks(id)
);

CREATE TABLE candles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    interval_type ENUM('1m', '5m', '15m', '30m', '1h', '4h', '1d') NOT NULL,
    start_time TIMESTAMP NOT NULL,

    open_price DECIMAL(20, 8) NOT NULL,
    high_price DECIMAL(20, 8) NOT NULL,
    low_price DECIMAL(20, 8) NOT NULL,
    close_price DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,

    UNIQUE KEY uniq_candle (interval_type, start_time)
);