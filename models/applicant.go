package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"oneCV/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Applicant struct {
	Id               uuid.UUID         `json:"id"`
	Name             string            `json:"name" binding:"required"`
	EmploymentStatus string            `json:"employment_status" binding:"required"`
	Sex              string            `json:"sex" binding:"required"`
	DateOfBirth      string            `json:"date_of_birth" binding:"required"`
	MaritalStatus    string            `json:"marital_status"`
	HouseholdMembers []HouseholdMember `json:"household"`
}

type HouseholdMember struct {
	Id               uuid.UUID `json:"id"`
	Name             *string   `json:"name"`
	EmploymentStatus *string   `json:"employment_status"`
	Sex              *string   `json:"sex" `
	Relation         *string   `json:"relation"`
	DateOfBirth      *string   `json:"date_of_birth"`
}

func (s *Applicant) GetAllApplicants(ctx context.Context, db *sql.DB) (data []Applicant, err error) {
	return s.FetchApplicant(ctx, db, "")
}

func (s *Applicant) FetchApplicant(ctx context.Context, db *sql.DB, whereClause string, args ...interface{}) (data []Applicant, err error) {
	query := `SELECT a.id, a.name, a.employment_status, a.sex, TO_CHAR(a.date_of_birth, 'YYYY-MM-DD') as date_of_birth, a.marital_status, hm.id as h_id, hm.name as h_name, hm.relation, TO_CHAR(hm.date_of_birth, 'YYYY-MM-DD') as hm_date_of_birth, hm.employment_status as h_employment_status, hm.sex as h_sex FROM applicants a LEFT JOIN household_members hm ON a.id = hm.applicant_id WHERE a.deleted = false ` + whereClause

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Println("Error querying applicants:", err)
		return nil, err
	}
	defer rows.Close()

	var applicantsMap = make(map[uuid.UUID]*Applicant)
	for rows.Next() {
		var applicant Applicant
		var householdMember HouseholdMember

		err := rows.Scan(&applicant.Id, &applicant.Name, &applicant.EmploymentStatus, &applicant.Sex, &applicant.DateOfBirth, &applicant.MaritalStatus, &householdMember.Id, &householdMember.Name, &householdMember.Relation, &householdMember.DateOfBirth, &householdMember.EmploymentStatus, &householdMember.Sex)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		if len(applicant.HouseholdMembers) == 0 {
			applicant.HouseholdMembers = []HouseholdMember{}
		}

		if existingApplicant, exists := applicantsMap[applicant.Id]; exists {
			if householdMember.Id != uuid.Nil {
				*householdMember.DateOfBirth = utils.DateFormat(*householdMember.DateOfBirth, "2006-01-02")
				existingApplicant.HouseholdMembers = append(existingApplicant.HouseholdMembers, householdMember)
			}
		} else {
			if householdMember.Id != uuid.Nil {
				*householdMember.DateOfBirth = utils.DateFormat(*householdMember.DateOfBirth, "2006-01-02")
				applicant.HouseholdMembers = append(applicant.HouseholdMembers, householdMember)
			}

			applicant.DateOfBirth = utils.DateFormat(applicant.DateOfBirth, "2006-01-02")
			applicantsMap[applicant.Id] = &applicant
		}
	}

	for _, applicant := range applicantsMap {
		data = append(data, *applicant)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil, err
	}

	return data, nil
}

func (s *Applicant) CreateApplicant(ctx context.Context, db *sql.DB) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO applicants (name, employment_status, sex, date_of_birth, marital_status) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var applicantId uuid.UUID
	err = tx.QueryRowContext(ctx, query, s.Name, strings.ToLower(s.EmploymentStatus), strings.ToLower(s.Sex), strings.ToLower(s.DateOfBirth), strings.ToLower(s.MaritalStatus)).Scan(&applicantId)
	if err != nil {
		log.Println("Error inserting applicant:", err)
		return err
	}

	// insert household member
	s.Id = applicantId
	err = s.CreateHouseholdMembers(ctx, tx)
	if err != nil {
		log.Println("Error create household member", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}
	return nil
}

func (s *Applicant) CreateHouseholdMembers(ctx context.Context, tx *sql.Tx) (err error) {
	if len(s.HouseholdMembers) > 0 {
		for _, member := range s.HouseholdMembers {
			memberQuery := `INSERT INTO household_members (applicant_id, name, date_of_birth, relation, employment_status, sex) VALUES ($1, $2, $3, $4, $5, $6)`

			log.Printf("query %+v s.Id %+v ", memberQuery, s.Id)
			_, err := tx.ExecContext(ctx, memberQuery, s.Id, member.Name, member.DateOfBirth, member.Relation, strings.ToLower(*member.EmploymentStatus), strings.ToLower(*member.Sex))
			if err != nil {
				log.Println("Error inserting household member:", err)
				return err
			}
		}
	}

	return nil
}

func (s *Applicant) UpdateApplicant(ctx context.Context, db *sql.DB) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	// Update the applicant record
	query := `UPDATE applicants SET name = $1, employment_status = $2, sex = $3, date_of_birth = $4, marital_status = $5, updated_at = $6 WHERE id = $7 AND deleted = false`
	_, err = tx.ExecContext(ctx, query, s.Name, strings.ToLower(s.EmploymentStatus), strings.ToLower(s.Sex), s.DateOfBirth, strings.ToLower(s.MaritalStatus), time.Now(), s.Id)
	if err != nil {
		log.Println("Error updating applicant:", err)
		return err
	}

	// Remove existing household members for the applicant
	deleteQuery := `DELETE FROM household_members WHERE applicant_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, s.Id)
	if err != nil {
		log.Println("Error deleting old household members:", err)
		return err
	}

	// Insert updated household members
	err = s.CreateHouseholdMembers(ctx, tx)
	if err != nil {
		log.Println("Error create household member", err)
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}
	return nil
}

func (s *Applicant) GetApplicantById(ctx context.Context, db *sql.DB) error {
	whereClause := ` AND a.id = $1`
	applicants, err := s.FetchApplicant(ctx, db, whereClause, s.Id)

	if err != nil {
		return err
	}

	if len(applicants) == 0 {
		return fmt.Errorf("applicant not found: %v", s.Id)
	}

	*s = applicants[0]
	return nil
}

func (s *Applicant) GetApplicantCriteriaData(ctx context.Context, db *sql.DB) ([]CriteriaData, error) {
	data := []CriteriaData{}

	// store applicant employment status
	employmentStatus := CriteriaData{}
	employmentStatus.CriteriaKey = "employment_status"
	employmentStatus.CriteriaValue = `"` + s.EmploymentStatus + `"`
	data = append(data, employmentStatus)

	// store applicant sex
	sex := CriteriaData{}
	sex.CriteriaKey = "sex"
	sex.CriteriaValue = `"` + s.Sex + `"`
	data = append(data, sex)

	// store applicant marital status
	maritalStatus := CriteriaData{}
	maritalStatus.CriteriaKey = "marital_status"
	maritalStatus.CriteriaValue = `"` + s.MaritalStatus + `"`
	data = append(data, maritalStatus)

	isPrimary := false
	isSecondary := false

	for _, v := range s.HouseholdMembers {
		age, err := utils.CalculateAge(*v.DateOfBirth)
		if err != nil {
			return []CriteriaData{}, err
		}

		if age >= 6 && age <= 12 {
			isPrimary = true
		} else if age >= 13 && age <= 18 {
			isSecondary = true
		}
	}

	if isPrimary {
		maritalStatus := CriteriaData{}
		maritalStatus.CriteriaKey = "has_children"
		maritalStatus.CriteriaValue = `{"school_level":"== primary"}`
		data = append(data, maritalStatus)
	}

	if isSecondary {
		maritalStatus := CriteriaData{}
		maritalStatus.CriteriaKey = "has_children"
		maritalStatus.CriteriaValue = `{"school_level":"== secondary"}`
		data = append(data, maritalStatus)
	}

	return data, nil
}

func (s *Applicant) CheckApplicantExist(ctx context.Context, db *sql.DB) error {
	query := `SELECT EXISTS(SELECT 1 from applicants WHERE id = $1 AND deleted = false)`
	var exists bool
	err := db.QueryRowContext(ctx, query, s.Id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking applicant existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("applicant %s does not exist", s.Id)
	}

	return nil
}

func (s *Applicant) DeleteApplicant(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `UPDATE applicants SET deleted = $1, updated_at = $2 WHERE id = $3 AND deleted = false`
	result, err := tx.ExecContext(ctx, query, true, time.Now(), s.Id)
	if err != nil {
		log.Println("Error updating applicant:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no applicant found with ID %s or applicant already deleted", s.Id)
	}

	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}
