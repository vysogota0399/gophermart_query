-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions(
  uuid UUID UNIQUE NOT NULL,
  order_number VARCHAR NOT NULL,
  account_id NUMERIC NOT NULL,
  amount NUMERIC NOT NULL,
  operation VARCHAR NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  processed_at TIMESTAMP WITH TIME ZONE NOT NULL,

  PRIMARY KEY(uuid)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd
