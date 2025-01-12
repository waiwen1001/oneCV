package models

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Scheme struct {
	Id          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"-"`
	Criteria    map[string]interface{} `json:"criteria"`
	Benefits    []Benefit              `json:"benefits"`
}

type Criteria struct {
	Id uuid.UUID
	CriteriaData
}

type CriteriaData struct {
	CriteriaKey   string
	CriteriaValue string
}

type Benefit struct {
	Id         uuid.UUID `json:"id"`
	CriteriaId uuid.UUID `json:"-"`
	Name       *string   `json:"name"`
	Amount     *float64  `json:"amount"`
}

type SchemeRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Criteria    []CriteriaRequest `json:"criteria"`
}

type CriteriaRequest struct {
	Conditions map[string]interface{} `json:"conditions"`
	Benefits   []BenefitRequest       `json:"benefits"`
}

type BenefitRequest struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

func (s *Scheme) GetAllSchemes(ctx context.Context, db *sql.DB) ([]Scheme, error) {
	return s.FetchSchemes(ctx, db, "")
}

func (s *Scheme) FetchSchemes(ctx context.Context, db *sql.DB, whereClause string, args ...interface{}) ([]Scheme, error) {
	query := `SELECT s.id, s.name, s.description, c.criteria_key, c.criteria_value, b.id AS b_id, b.name AS b_name, b.amount FROM schemes s LEFT JOIN criteria c ON s.id = c.scheme_id LEFT JOIN benefits b ON c.id = b.criteria_id WHERE s.deleted = false AND c.deleted = false AND b.deleted = false ` + whereClause + ` ORDER BY s.created_at DESC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Println("Error querying schemes:", err)
		return nil, err
	}
	defer rows.Close()

	schemeMap := make(map[uuid.UUID]*Scheme)
	for rows.Next() {
		var scheme Scheme
		var criteria Criteria
		var benefit Benefit

		err = rows.Scan(&scheme.Id, &scheme.Name, &scheme.Description, &criteria.CriteriaKey, &criteria.CriteriaValue, &benefit.Id, &benefit.Name, &benefit.Amount)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		if _, exists := schemeMap[scheme.Id]; !exists {
			schemeMap[scheme.Id] = &Scheme{
				Id:          scheme.Id,
				Name:        scheme.Name,
				Description: scheme.Description,
				Criteria:    make(map[string]interface{}),
				Benefits:    []Benefit{},
			}
		}

		if criteria.CriteriaKey != "" && criteria.CriteriaValue != "" {
			var criteriaValue interface{}
			err := json.Unmarshal([]byte(criteria.CriteriaValue), &criteriaValue)
			if err != nil {
				criteriaValue = criteria.CriteriaValue
			}

			// check criteria exist
			if _, exists := schemeMap[scheme.Id].Criteria[criteria.CriteriaKey]; !exists {
				schemeMap[scheme.Id].Criteria[criteria.CriteriaKey] = criteriaValue
			}
		}

		if benefit.Name != nil && benefit.Amount != nil {
			if schemeMap[scheme.Id].Benefits == nil {
				schemeMap[scheme.Id].Benefits = make([]Benefit, 0)
			}

			// check benefit exist
			exists := false
			for _, b := range schemeMap[scheme.Id].Benefits {
				if b.Id == benefit.Id {
					exists = true
					break
				}
			}

			if !exists {
				schemeMap[scheme.Id].Benefits = append(schemeMap[scheme.Id].Benefits, Benefit{
					Id:     benefit.Id,
					Name:   benefit.Name,
					Amount: benefit.Amount,
				})
			}
		}
	}

	var schemes []Scheme
	for _, scheme := range schemeMap {
		schemes = append(schemes, *scheme)
	}

	return schemes, nil
}

func (s *Scheme) CreateScheme(ctx context.Context, db *sql.DB, req SchemeRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO schemes (name, description) VALUES ($1, $2) RETURNING id`
	var schemeID uuid.UUID
	err = tx.QueryRowContext(ctx, query, req.Name, req.Description).Scan(&schemeID)
	if err != nil {
		return fmt.Errorf("could not insert scheme: %v", err)
	}

	err = s.CreateCriteriaAndBenefit(ctx, tx, req)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (s *Scheme) GetSchemeCriteria(ctx context.Context, db *sql.DB) ([]Criteria, error) {
	query := `SELECT c.id, c.criteria_key, c.criteria_value FROM schemes s LEFT JOIN criteria c ON s.id = c.scheme_id WHERE s.id = $1 AND c.deleted = false AND s.deleted = false`

	rows, err := db.QueryContext(ctx, query, s.Id)
	if err != nil {
		log.Println("Error querying schemes:", err)
		return nil, err
	}
	defer rows.Close()

	var criteria []Criteria

	for rows.Next() {
		var crit Criteria
		err := rows.Scan(&crit.Id, &crit.CriteriaKey, &crit.CriteriaValue)
		if err != nil {
			log.Println("Error scanning criteria row:", err)
			return nil, err
		}

		criteria = append(criteria, crit)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error with rows:", err)
		return nil, err
	}

	return criteria, nil
}

func (s *Scheme) GetEligibleSchemes(ctx context.Context, db *sql.DB, applicant Applicant) ([]Scheme, error) {
	schemes := []Scheme{}
	criteriaData, err := applicant.GetApplicantCriteriaData(ctx, db)
	if err != nil {
		return schemes, err
	}

	var query strings.Builder
	var args []interface{}

	query.WriteString(" AND ( ")

	for i, v := range criteriaData {
		if i > 0 {
			query.WriteString(" OR ")
		}
		query.WriteString("(c.criteria_key = $")
		query.WriteString(fmt.Sprintf("%d", i*2+1))
		query.WriteString(" AND c.criteria_value = $")
		query.WriteString(fmt.Sprintf("%d", i*2+2))
		query.WriteString(")")

		args = append(args, v.CriteriaKey, v.CriteriaValue)
	}

	query.WriteString(" )")

	schemes, err = s.FetchSchemes(ctx, db, query.String(), args...)
	if err != nil {
		return schemes, err
	}

	return schemes, nil
}

func (s *Scheme) GetBenefitsByCriteriaIds(ctx context.Context, db *sql.DB, ids []uuid.UUID) ([]Benefit, error) {
	query := `SELECT id, criteria_id, name, amount FROM benefits WHERE deleted = false AND criteria_id = ANY($1)`

	rows, err := db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		log.Println("Error querying benefits:", err)
		return nil, err
	}
	defer rows.Close()

	var benefits []Benefit

	for rows.Next() {
		var benefit Benefit
		if err := rows.Scan(&benefit.Id, &benefit.CriteriaId, &benefit.Name, &benefit.Amount); err != nil {
			log.Println("Error scanning benefit row:", err)
			return nil, err
		}
		benefits = append(benefits, benefit)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil, err
	}

	return benefits, nil
}

func (s *Scheme) CheckSchemeExist(ctx context.Context, db *sql.DB) error {
	query := `SELECT EXISTS(SELECT 1 from schemes WHERE id = $1 AND deleted = false)`
	var exists bool
	err := db.QueryRowContext(ctx, query, s.Id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking scheme existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("scheme %s does not exist", s.Id)
	}

	return nil
}

func (s *Scheme) DeleteScheme(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `UPDATE schemes SET deleted = $1, updated_at = $2 WHERE id = $3 AND deleted = false`
	result, err := tx.ExecContext(ctx, query, true, time.Now(), s.Id)
	if err != nil {
		log.Println("Error updating scheme:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no scheme found with ID %s or scheme already deleted", s.Id)
	}

	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}

func (s *Scheme) UpdateScheme(ctx context.Context, db *sql.DB, req SchemeRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `UPDATE schemes SET name = $1, description = $2, updated_at = $3 WHERE id = $4 AND deleted = false`
	_, err = tx.ExecContext(ctx, query, req.Name, req.Description, time.Now(), s.Id)
	if err != nil {
		log.Println("Error updating applicant:", err)
		return err
	}

	// Remove existing criteria and benefits for the scheme
	deleteBenefitQuery := `UPDATE benefits SET deleted = $1, updated_at = $2 WHERE scheme_id = $3`
	_, err = tx.ExecContext(ctx, deleteBenefitQuery, true, time.Now(), s.Id)
	if err != nil {
		log.Println("Error deleting old benefits:", err)
		return err
	}

	deleteCriteriaQuery := `UPDATE criteria SET deleted = $1, updated_at = $2 WHERE scheme_id = $3`
	_, err = tx.ExecContext(ctx, deleteCriteriaQuery, true, time.Now(), s.Id)
	if err != nil {
		log.Println("Error deleting old criteria:", err)
		return err
	}

	err = s.CreateCriteriaAndBenefit(ctx, tx, req)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (s *Scheme) CreateCriteriaAndBenefit(ctx context.Context, tx *sql.Tx, req SchemeRequest) error {
	for _, criteria := range req.Criteria {
		for i, condition := range criteria.Conditions {
			criteriaValue, err := json.Marshal(condition)
			if err != nil {
				return fmt.Errorf("marshal criteria value failed: %v", err)
			}
			insertCriteria := `INSERT INTO criteria (scheme_id, criteria_key, criteria_value) VALUES ($1, $2, $3) RETURNING id`
			var criteriaID uuid.UUID
			err = tx.QueryRowContext(ctx, insertCriteria, s.Id, i, bytes.ToLower(criteriaValue)).Scan(&criteriaID)
			if err != nil {
				return fmt.Errorf("could not insert criteria: %v", err)
			}

			// insert benefits
			for _, benefit := range criteria.Benefits {
				insertBenefit := `INSERT INTO benefits (scheme_id, criteria_id, name, amount) VALUES ($1, $2, $3, $4)`
				_, err := tx.ExecContext(ctx, insertBenefit, s.Id, criteriaID, benefit.Name, benefit.Amount)
				if err != nil {
					return fmt.Errorf("could not insert benefit: %v", err)
				}
			}
		}
	}

	return nil
}
