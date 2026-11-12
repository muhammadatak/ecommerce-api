-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cart_items (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id     UUID         NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    product_id  UUID         NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    quantity    INTEGER      NOT NULL CHECK (quantity > 0),
    updated_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_cart_items_product_id ON cart_items (product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart_items;
-- +goose StatementEnd
