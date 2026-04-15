CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE order_status AS ENUM (
    'NEW',
    'PROCESSING',
    'INVALID',
    'PROCESSED'
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    number TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status order_status NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(19,2) NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    next_poll_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_uploaded_at
    ON orders (user_id, uploaded_at DESC);

CREATE INDEX idx_orders_polling
    ON orders (status, next_poll_at);

CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number TEXT NOT NULL UNIQUE,
    amount NUMERIC(19,2) NOT NULL CHECK (amount > 0),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_withdrawals_user_processed_at
    ON withdrawals (user_id, processed_at DESC);

-- docker exec -i gophermart-postgres psql -U gophermart -d gophermartdb < migrations/001_init.up.sql
