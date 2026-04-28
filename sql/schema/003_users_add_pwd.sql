-- +goose Up
ALTER TABLE users
ADD password TEXT NOT NULL DEFAULT 'totallysafedefaultthatsnotevenahash';

-- +goose Down
ALTER TABLE users
DROP COLUMN password;