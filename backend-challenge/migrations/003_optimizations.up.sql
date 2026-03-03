-- Optimization indexes (idempotent — safe to re-run on every startup).
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
