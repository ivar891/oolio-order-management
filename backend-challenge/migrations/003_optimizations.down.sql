-- Remove check constraint from products table.
ALTER TABLE products DROP CONSTRAINT IF EXISTS check_stock_non_negative;

-- Remove indexes from order_items table.
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
