CREATE INDEX IF NOT EXISTS idx_delivery_routes_driver_id ON delivery_routes(driver_id);

ALTER TABLE delivery_routes 
ADD CONSTRAINT fk_delivery_routes_driver 
FOREIGN KEY (driver_id) 
REFERENCES users(id) 
ON DELETE SET NULL;
