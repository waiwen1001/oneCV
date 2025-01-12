-- +goose Up
-- +goose StatementBegin
CREATE TABLE application_details (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  application_id UUID NOT NULL,
  criteria_id UUID,
  criteria_name VARCHAR(255),
  criteria_key VARCHAR(255),
  criteria_value TEXT,
  benefit_id UUID,
  benefit_name VARCHAR(255),
  benefit_amount DECIMAl(16, 2),
  created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours'),
  updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours')
);

ALTER TABLE application_details ADD CONSTRAINT fk_application_id FOREIGN KEY (application_id) REFERENCES applications(id);
ALTER TABLE application_details ADD CONSTRAINT fk_criteria_id FOREIGN KEY (criteria_id) REFERENCES criteria(id);
ALTER TABLE application_details ADD CONSTRAINT fk_benefit_id FOREIGN KEY (benefit_id) REFERENCES benefits(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE application_details DROP CONSTRAINT fk_application_id;
ALTER TABLE application_details DROP CONSTRAINT fk_criteria_id;
ALTER TABLE application_details DROP CONSTRAINT fk_benefit_id;
DROP TABLE IF EXISTS application_details;
-- +goose StatementEnd
