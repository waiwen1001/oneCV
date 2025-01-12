-- +goose Up
-- +goose StatementBegin
CREATE TABLE applications (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  applicant_id UUID NOT NULL,
  scheme_id UUID NOT NULL,
  status VARCHAR(255) NOT NULL,
  submitted_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours'),
  updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours')
);

ALTER TABLE applications ADD CONSTRAINT fk_applicant_id FOREIGN KEY (applicant_id) REFERENCES applicants(id);
ALTER TABLE applications ADD CONSTRAINT fk_scheme_id FOREIGN KEY (scheme_id) REFERENCES schemes(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE applications DROP CONSTRAINT fk_scheme_id;
ALTER TABLE applications DROP CONSTRAINT fk_applicant_id;
DROP TABLE IF EXISTS applications;
-- +goose StatementEnd
