-- +goose Up
-- +goose StatementBegin
CREATE TABLE applicants (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  employment_status VARCHAR(255) NOT NULL,
  sex VARCHAR(255) NOT NULL,
  date_of_birth DATE NOT NULL,
  marital_status VARCHAR(255) NOT NULL,
  deleted BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours'),
  updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC' + INTERVAL '8 hours')
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS applicants;
-- +goose StatementEnd
