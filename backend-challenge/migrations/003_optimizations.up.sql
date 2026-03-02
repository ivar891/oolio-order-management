-- Add check constraint to prevent negative stock at the database level.
ALTER TABLE products ADD CONSTRAINT check_stock_non_negative CHECK (stock_quantity >= 0);

-- Add indexes for better join and lookup performance in orders.
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
