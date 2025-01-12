package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"oneCV/config"
	"oneCV/utils"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Application struct {
	Id          uuid.UUID `json:"id"`
	ApplicantID uuid.UUID `json:"applicant_id"`
	SchemeID    uuid.UUID `json:"scheme_id"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApplicationRequest struct {
	ApplicantID uuid.UUID `json:"applicant_id"`
	SchemeID    uuid.UUID `json:"scheme_id"`
}

type ApplicationUpdateRequest struct {
	Status string `json:"status"`
}

type ApplicationDetail struct {
	ApplicationId uuid.UUID `json:"application_id"`
	CriteriaId    uuid.UUID `json:"criteria_id"`
	CriteriaName  string    `json:"criteria_name"`
	CriteriaKey   string    `json:"criteria_key"`
	CriteriaValue string    `json:"criteria_value"`
	BenefitId     uuid.UUID `json:"benefit_id"`
	BenefitName   string    `json:"benefit_name"`
	BenefitAmount float64   `json:"benefit_amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ApplicationResult struct {
	Id          uuid.UUID            `json:"application_id"`
	Applicant   ApplicationApplicant `json:"applicant"`
	Scheme      ApplicationScheme    `json:"scheme"`
	Status      string               `json:"status"`
	SubmittedAt string               `json:"submitted_at"`
}

type ApplicationApplicant struct {
	Id               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	EmploymentStatus string    `json:"employment_status"`
}

type ApplicationScheme struct {
	Id               uuid.UUID             `json:"id"`
	Name             string                `json:"name"`
	EligibleCriteria []ApplicationEligible `json:"eligible"`
}

type ApplicationEligible struct {
	Criteria      map[string]interface{} `json:"criteria"`
	CriteriaKey   string                 `json:"-"`
	CriteriaValue string                 `json:"-"`
	Benefit       []Benefit              `json:"benefits"`
}

func (ac *Application) CreateApplication(ctx context.Context, db *sql.DB, req ApplicationRequest) error {
	applicant := Applicant{Id: req.ApplicantID}
	if err := applicant.GetApplicantById(ctx, db); err != nil {
		return err
	}

	scheme := Scheme{Id: req.SchemeID}
	criteria, err := scheme.GetSchemeCriteria(ctx, db)

	if err != nil {
		return err
	}

	eligibleCriteria, err := ac.CheckEligibility(applicant, criteria)
	if err != nil {
		return err
	}

	if len(eligibleCriteria) == 0 {
		return fmt.Errorf("this scheme is not eligible for this applicant")
	}

	ac.ApplicantID = applicant.Id
	ac.SchemeID = scheme.Id
	ac.Status = config.StatusPending
	ac.SubmittedAt = time.Now()

	err = ac.SaveApplication(ctx, db, scheme.Id, eligibleCriteria)
	if err != nil {
		return err
	}

	return nil
}

func (ac *Application) CheckEligibility(applicant Applicant, criteria []Criteria) ([]Criteria, error) {
	eligibleCriteria := []Criteria{}
	for _, v := range criteria {
		if len(v.CriteriaValue) > 0 {
			isJson := utils.IsJson(v.CriteriaValue)
			if !isJson {
				v.CriteriaValue = strings.ReplaceAll(v.CriteriaValue, `"`, "")
			}

			if isJson {
				var criteriaValue map[string]interface{}
				err := json.Unmarshal([]byte(v.CriteriaValue), &criteriaValue)
				if err != nil {
					log.Printf("Error unmarshaling criteria value for criteria %d: %v", v.Id, err)
					return eligibleCriteria, err
				}

				// check has_children
				if criteriaValue["school_level"] != nil {
					for _, member := range applicant.HouseholdMembers {
						age, err := utils.CalculateAge(*member.DateOfBirth)
						if err != nil {
							log.Printf("Parse applicant household member age error %+v", err)
							return eligibleCriteria, err
						}

						switch criteriaValue["school_level"] {
						case "== primary":
							if age >= 6 && age <= 12 {
								eligibleCriteria = append(eligibleCriteria, v)
							}
						case "== secondary":
							if age >= 13 && age <= 18 {
								eligibleCriteria = append(eligibleCriteria, v)
							}
						default:
							log.Printf("Applicant child not eligible %+v %+v", criteriaValue["school_level"], age)
							return eligibleCriteria, err
						}
					}
				}
			} else {
				switch v.CriteriaKey {
				case "employment_status":
					if v.CriteriaValue == applicant.EmploymentStatus {
						eligibleCriteria = append(eligibleCriteria, v)
					}
				case "marital_status":
					if v.CriteriaValue == applicant.MaritalStatus {
						eligibleCriteria = append(eligibleCriteria, v)
					}
				case "sex":
					if v.CriteriaValue == applicant.Sex {
						eligibleCriteria = append(eligibleCriteria, v)
					}
				}
			}
		}
	}

	return eligibleCriteria, nil
}

func (ac *Application) SaveApplication(ctx context.Context, db *sql.DB, schemUUID uuid.UUID, criteria []Criteria) error {
	criteriaIds := []uuid.UUID{}
	for _, v := range criteria {
		criteriaIds = append(criteriaIds, v.Id)
	}

	scheme := Scheme{}
	benefits, err := scheme.GetBenefitsByCriteriaIds(ctx, db, criteriaIds)
	if err != nil {
		return err
	}

	// create application
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO applications (applicant_id, scheme_id, status, submitted_at) VALUES ($1, $2, $3, $4) RETURNING (id)`

	var applicationId uuid.UUID
	err = tx.QueryRowContext(ctx, query, ac.ApplicantID, ac.SchemeID, ac.Status, ac.SubmittedAt).Scan(&applicationId)
	if err != nil {
		log.Println("Error inserting application:", err)
		return err
	}

	for _, b := range benefits {
		ad := ApplicationDetail{ApplicationId: applicationId, BenefitId: b.Id, BenefitName: *b.Name, BenefitAmount: *b.Amount}

		for _, c := range criteria {
			if c.Id == b.CriteriaId {
				ad.CriteriaId = c.Id
				ad.CriteriaKey = c.CriteriaKey
				ad.CriteriaValue = c.CriteriaValue

				break
			}
		}

		query := `INSERT INTO application_details (application_id, criteria_id, criteria_name, criteria_key, criteria_value, benefit_id, benefit_name, benefit_amount) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = tx.ExecContext(ctx, query, ad.ApplicationId, ad.CriteriaId, ad.CriteriaName, ad.CriteriaKey, ad.CriteriaValue, ad.BenefitId, ad.BenefitName, ad.BenefitAmount)
		if err != nil {
			log.Println("Error inserting application detail:", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (ac *Application) GetAllApplications(ctx context.Context, db *sql.DB) ([]ApplicationResult, error) {
	query := `SELECT a.id AS a_id, app.id AS app_id, app.name AS app_name, app.employment_status, s.id AS s_id, s.name AS s_name, ad.criteria_key, ad.criteria_value, ad.benefit_id, b.name AS benefit_name, b.amount, a.status, TO_CHAR(a.submitted_at, 'YYYY-MM-DD HH24:MI:SS') as submitted_at FROM applications a INNER JOIN applicants app ON a.applicant_id = app.id INNER JOIN schemes s ON a.scheme_id = s.id LEFT JOIN application_details ad ON ad.application_id = a.id LEFT JOIN benefits b ON ad.benefit_id = b.id`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	applicationMap := make(map[uuid.UUID]*ApplicationResult)

	for rows.Next() {
		var id uuid.UUID
		var applicant ApplicationApplicant
		var scheme ApplicationScheme
		var status string
		var submittedAt string
		var criteriaKey, criteriaValue string
		var benefit Benefit

		if err := rows.Scan(&id, &applicant.Id, &applicant.Name, &applicant.EmploymentStatus, &scheme.Id, &scheme.Name, &criteriaKey, &criteriaValue, &benefit.Id, &benefit.Name, &benefit.Amount, &status, &submittedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		application, exists := applicationMap[id]
		if !exists {
			application = &ApplicationResult{
				Id: id,
				Applicant: ApplicationApplicant{
					Id:               applicant.Id,
					Name:             applicant.Name,
					EmploymentStatus: applicant.EmploymentStatus,
				},
				Scheme: ApplicationScheme{
					Id:               scheme.Id,
					Name:             scheme.Name,
					EligibleCriteria: []ApplicationEligible{},
				},
				Status:      status,
				SubmittedAt: submittedAt,
			}
			applicationMap[id] = application
		}

		if criteriaKey != "" && criteriaValue != "" {
			found := false
			var parsedCriteria interface{}
			if err := json.Unmarshal([]byte(criteriaValue), &parsedCriteria); err != nil {
				parsedCriteria = criteriaValue
			}

			for i, eligible := range application.Scheme.EligibleCriteria {
				if val, ok := eligible.Criteria[criteriaKey]; ok && reflect.DeepEqual(val, parsedCriteria) {
					application.Scheme.EligibleCriteria[i].Benefit = append(application.Scheme.EligibleCriteria[i].Benefit, benefit)
					found = true
					break
				}
			}
			if !found {
				application.Scheme.EligibleCriteria = append(application.Scheme.EligibleCriteria, ApplicationEligible{
					Criteria: map[string]interface{}{
						criteriaKey: parsedCriteria,
					},
					Benefit: []Benefit{benefit},
				})
			}
		}
	}

	var results []ApplicationResult
	for _, application := range applicationMap {
		results = append(results, *application)
	}

	return results, nil
}

func (ac *Application) CheckApplicationExist(ctx context.Context, db *sql.DB) error {
	query := `SELECT EXISTS(SELECT 1 from applications WHERE id = $1)`
	var exists bool
	err := db.QueryRowContext(ctx, query, ac.Id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking application existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("application %s does not exist", ac.Id)
	}

	return nil
}

func (ac *Application) DeleteApplication(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	adQuery := `DELETE FROM application_details WHERE application_id = $1`
	_, err = tx.ExecContext(ctx, adQuery, ac.Id)
	if err != nil {
		log.Println("Error delete application details:", err)
		return err
	}

	aQuery := `DELETE FROM applications WHERE id = $1`
	result, err := tx.ExecContext(ctx, aQuery, ac.Id)
	if err != nil {
		log.Println("Error delete application:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no application found with ID %s or application already deleted", ac.Id)
	}

	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}

func (ac *Application) UpdateApplication(ctx context.Context, db *sql.DB, req ApplicationUpdateRequest) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	query := `UPDATE applications SET status = $1, updated_at = $2 WHERE id = $3`
	_, err = tx.ExecContext(ctx, query, req.Status, time.Now(), ac.Id)
	if err != nil {
		log.Println("Error updating applicant:", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}
