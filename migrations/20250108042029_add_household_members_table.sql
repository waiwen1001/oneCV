-- +goose Up
-- +goose StatementBegin
CREATE TABLE household_members (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  applicant_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  employment_status VARCHAR(255) NOT NULL,
  sex VARCHAR(255),
  date_of_birth DATE NOT NULL,
  relation VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours'),
  updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours')
);

ALTER TABLE household_members ADD CONSTRAINT fk_applicant_id FOREIGN KEY (applicant_id) REFERENCES applicants(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE household_members DROP CONSTRAINT fk_applicant_id;
DROP TABLE IF EXISTS household_members;
-- +goose StatementEnd
