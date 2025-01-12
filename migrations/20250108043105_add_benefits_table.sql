-- +goose Up
-- +goose StatementBegin
CREATE TABLE benefits (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  scheme_id UUID NOT NULL,
  criteria_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  amount DECIMAl(16, 2) NOT NULL,
  deleted BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours'),
  updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours')
);

ALTER TABLE benefits ADD CONSTRAINT fk_scheme_id FOREIGN KEY (scheme_id) REFERENCES schemes(id);
ALTER TABLE benefits ADD CONSTRAINT fk_criteria_id FOREIGN KEY (criteria_id) REFERENCES criteria(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE benefits DROP CONSTRAINT fk_scheme_id;
ALTER TABLE benefits DROP CONSTRAINT fk_criteria_id;
DROP TABLE IF EXISTS benefits;
-- +goose StatementEnd
