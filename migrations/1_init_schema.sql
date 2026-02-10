-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email     VARCHAR NOT NULL,
    pass_hash VARCHAR NOT NULL,
    user_role VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS pvz
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_date TIMESTAMP NOT NULL DEFAULT NOW(),
    city              VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS reception
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    pvz_id    UUID,
    status    VARCHAR NOT NULL,

    CONSTRAINT fk_pvz_id FOREIGN KEY (pvz_id) REFERENCES pvz(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS products
(
    id UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time    TIMESTAMP NOT NULL DEFAULT NOW(),
    product_type VARCHAR NOT NULL,
    reception_id UUID NOT NULL,

    CONSTRAINT fk_reception_id FOREIGN KEY (reception_id) REFERENCES reception(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS pvz;
DROP TABLE IF EXISTS reception;
DROP TABLE IF EXISTS products;