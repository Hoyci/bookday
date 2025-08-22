ALTER TABLE delivery_routes 
DROP CONSTRAINT IF EXISTS fk_delivery_routes_driver;

DROP INDEX IF EXISTS idx_delivery_routes_driver_id;
