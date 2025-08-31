-- CrossPay Database Schema

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    blockchain_id BIGINT NOT NULL,
    sender_address VARCHAR(42) NOT NULL,
    recipient_address VARCHAR(42) NOT NULL,
    sender_ens VARCHAR(255),
    recipient_ens VARCHAR(255),
    token_address VARCHAR(42) NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    fee DECIMAL(36, 18) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    tx_hash VARCHAR(66),
    chain_id INTEGER NOT NULL,
    metadata_uri TEXT,
    receipt_cid VARCHAR(100),
    oracle_price VARCHAR(50),
    random_seed VARCHAR(66),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    INDEX idx_sender (sender_address),
    INDEX idx_recipient (recipient_address),
    INDEX idx_status (status),
    INDEX idx_chain (chain_id),
    INDEX idx_tx_hash (tx_hash)
);

-- Receipts table
CREATE TABLE IF NOT EXISTS receipts (
    id SERIAL PRIMARY KEY,
    payment_id BIGINT REFERENCES payments(blockchain_id),
    receipt_cid VARCHAR(100) NOT NULL UNIQUE,
    metadata_cid VARCHAR(100),
    format VARCHAR(10) NOT NULL DEFAULT 'json',
    language VARCHAR(5) NOT NULL DEFAULT 'en',
    content_hash VARCHAR(66),
    signature TEXT,
    verified BOOLEAN DEFAULT FALSE,
    compliance_fields TEXT,
    creator_address VARCHAR(42) NOT NULL,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP,
    INDEX idx_payment (payment_id),
    INDEX idx_creator (creator_address),
    INDEX idx_cid (receipt_cid),
    INDEX idx_verified (verified)
);

-- Oracle requests table
CREATE TABLE IF NOT EXISTS oracle_requests (
    id SERIAL PRIMARY KEY,
    request_id VARCHAR(100) NOT NULL UNIQUE,
    request_type VARCHAR(20) NOT NULL, -- 'price', 'random', 'proof'
    symbol VARCHAR(20),
    requester_address VARCHAR(42),
    payment_id BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    fulfilled_at TIMESTAMP,
    INDEX idx_request_id (request_id),
    INDEX idx_type (request_type),
    INDEX idx_status (status),
    INDEX idx_requester (requester_address)
);

-- ENS cache table
CREATE TABLE IF NOT EXISTS ens_cache (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    address VARCHAR(42) NOT NULL,
    avatar_url TEXT,
    text_records JSONB,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ttl INTEGER DEFAULT 3600,
    INDEX idx_name (name),
    INDEX idx_address (address),
    INDEX idx_cached_at (cached_at)
);

-- Storage operations table  
CREATE TABLE IF NOT EXISTS storage_operations (
    id SERIAL PRIMARY KEY,
    operation_id VARCHAR(100) NOT NULL UNIQUE,
    operation_type VARCHAR(20) NOT NULL, -- 'upload', 'retrieve'
    file_cid VARCHAR(100),
    filename VARCHAR(255),
    file_size BIGINT,
    content_type VARCHAR(100),
    storage_cost DECIMAL(18, 8),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_id BIGINT,
    receipt_id BIGINT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    INDEX idx_operation_id (operation_id),
    INDEX idx_cid (file_cid),
    INDEX idx_type (operation_type),
    INDEX idx_status (status)
);

-- Analytics aggregations table
CREATE TABLE IF NOT EXISTS analytics_daily (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    chain_id INTEGER NOT NULL,
    payment_count BIGINT DEFAULT 0,
    payment_volume DECIMAL(36, 18) DEFAULT 0,
    completed_payments BIGINT DEFAULT 0,
    receipts_generated BIGINT DEFAULT 0,
    receipts_verified BIGINT DEFAULT 0,
    oracle_requests BIGINT DEFAULT 0,
    ens_resolutions BIGINT DEFAULT 0,
    storage_operations BIGINT DEFAULT 0,
    unique_senders BIGINT DEFAULT 0,
    unique_recipients BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, chain_id),
    INDEX idx_date (date),
    INDEX idx_chain_date (chain_id, date)
);

-- Service health table
CREATE TABLE IF NOT EXISTS service_health (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'healthy', 'unhealthy', 'degraded'
    last_check TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    response_time_ms INTEGER,
    error_count INTEGER DEFAULT 0,
    metadata JSONB,
    INDEX idx_service (service_name),
    INDEX idx_status (status),
    INDEX idx_last_check (last_check)
);

-- Insert initial data
INSERT INTO service_health (service_name, status) VALUES
    ('payment-processor', 'healthy'),
    ('storage-worker', 'healthy'), 
    ('oracle-service', 'healthy'),
    ('ens-resolver', 'healthy')
ON CONFLICT (service_name) DO NOTHING;

-- Create views for common queries
CREATE OR REPLACE VIEW payment_summary AS
SELECT 
    p.id,
    p.blockchain_id,
    p.sender_address,
    p.recipient_address,
    p.sender_ens,
    p.recipient_ens,
    p.amount,
    p.status,
    p.chain_id,
    p.created_at,
    r.receipt_cid,
    r.verified as receipt_verified
FROM payments p
LEFT JOIN receipts r ON p.blockchain_id = r.payment_id;

CREATE OR REPLACE VIEW daily_stats AS
SELECT 
    date,
    SUM(payment_count) as total_payments,
    SUM(payment_volume) as total_volume,
    SUM(completed_payments) as total_completed,
    SUM(receipts_generated) as total_receipts,
    SUM(oracle_requests) as total_oracle_requests
FROM analytics_daily
GROUP BY date
ORDER BY date DESC;