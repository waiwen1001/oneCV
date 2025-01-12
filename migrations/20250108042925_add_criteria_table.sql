-- +goose Up
-- +goose StatementBegin
CREATE TABLE criteria (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  scheme_id UUID NOT NULL,
  name VARCHAR(255),
  criteria_key VARCHAR(255) NOT NULL,
  criteria_value TEXT NOT NULL,
  deleted BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours'),
  updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours')
);

ALTER TABLE criteria ADD CONSTRAINT fk_scheme_id FOREIGN KEY (scheme_id) REFERENCES schemes(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE criteria DROP CONSTRAINT fk_scheme_id;
DROP TABLE IF EXISTS criteria;
-- +goose StatementEnd
