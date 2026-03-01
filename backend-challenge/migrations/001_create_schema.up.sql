-- Order Food Online Schema
-- Products with stock tracking and currency
CREATE TABLE IF NOT EXISTS products (
    id             TEXT PRIMARY KEY,
    name           TEXT NOT NULL,
    price          NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    currency       TEXT NOT NULL DEFAULT 'USD',
    category       TEXT NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    image_thumb    TEXT,
    image_mobile   TEXT,
    image_tablet   TEXT,
    image_desktop  TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Promo codes loaded from data files with optional discount percentage
CREATE TABLE IF NOT EXISTS promo_codes (
    code                TEXT NOT NULL,
    file_index          SMALLINT NOT NULL CHECK (file_index BETWEEN 1 AND 3),
    discount_percentage NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    PRIMARY KEY (code, file_index)
);

-- Orders with currency
CREATE TABLE IF NOT EXISTS orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_code TEXT,
    currency    TEXT NOT NULL DEFAULT 'USD',
    total       NUMERIC(10,2) NOT NULL CHECK (total >= 0),
    discounts   NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (discounts >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Order line items
CREATE TABLE IF NOT EXISTS order_items (
    id         SERIAL PRIMARY KEY,
    order_id   UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL REFERENCES products(id),
    quantity   INT NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10,2) NOT NULL CHECK (unit_price >= 0)
);

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_promo_codes_code ON promo_codes(code);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
