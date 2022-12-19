-- +migrate Up
CREATE TABLE balances (
    id            serial primary key,
    user_id       character(64) NOT NULL,
    chain_id      character(64) NOT NULL,
    chain_type    character(32) NOT NULL,
    token_address character(64),
    amount        decimal NOT NULL
);

-- +migrate Down
DROP TABLE balances cascade;