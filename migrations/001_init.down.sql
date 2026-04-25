DROP INDEX IF EXISTS idx_withdrawals_user_processed_at;
DROP TABLE IF EXISTS withdrawals;

DROP INDEX IF EXISTS idx_orders_polling;
DROP INDEX IF EXISTS idx_orders_user_uploaded_at;
DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS order_status;
