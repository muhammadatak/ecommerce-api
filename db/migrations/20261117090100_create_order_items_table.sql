-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS order_items (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      UUID         NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id    UUID         NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    product_name  VARCHAR(255) NOT NULL,
    price         INTEGER      NOT NULL CHECK (price >= 0),
    quantity      INTEGER      NOT NULL CHECK (quantity > 0),
    subtotal      INTEGER      NOT NULL CHECK (subtotal >= 0),
    created_at    TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items (product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_items;
-- +goose StatementEnd
