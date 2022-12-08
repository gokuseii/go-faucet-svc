-- +migrate Up
CREATE TABLE balances (
    user_id       character(64) NOT NULL,
    chain_id      character(64) NOT NULL,
    chain_type    character(32) NOT NULL,
    token_address character(64) NOT NULL,
    amount        float NOT NULL
);

CREATE UNIQUE INDEX uniq_idx_balances
    ON balances(user_id, chain_id, chain_type, token_address);

-- +migrate Down
DROP TABLE balances cascade;