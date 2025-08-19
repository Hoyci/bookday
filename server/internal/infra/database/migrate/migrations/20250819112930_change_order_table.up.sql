ALTER TABLE orders ADD COLUMN customer_id UUID;

ALTER TABLE orders 
ADD CONSTRAINT fk_orders_customer 
FOREIGN KEY (customer_id) 
REFERENCES users(id) 
ON DELETE SET NULL;

CREATE INDEX idx_orders_customer_id ON orders(customer_id);

ALTER TABLE orders DROP COLUMN customer_name;
