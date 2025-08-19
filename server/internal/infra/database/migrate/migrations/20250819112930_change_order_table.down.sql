ALTER TABLE orders ADD COLUMN customer_name VARCHAR(255);

DROP INDEX IF EXISTS idx_orders_customer_id;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_customer;

ALTER TABLE orders DROP COLUMN customer_id;

