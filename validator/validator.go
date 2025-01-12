package validator

import (
	"log"
	"oneCV/config"
	"oneCV/models"
	"strings"

	"github.com/google/uuid"
)

func ValidateApplicantForm(applicant models.Applicant) bool {
	if len(applicant.HouseholdMembers) > 0 {
		if validate := ValidateHouseholdMembers(applicant.HouseholdMembers); !validate {
			return false
		}
	}
	return applicant.DateOfBirth != "" && ValidateEmploymentStatus(applicant.EmploymentStatus) && ValidateMaritalStatus(applicant.MaritalStatus) && ValidateSex(applicant.Sex)
}

func ValidateEmploymentStatus(status string) bool {
	validStatuses := []string{"employed", "unemployed"}
	return Validator(status, validStatuses)
}

func ValidateMaritalStatus(status string) bool {
	if status == "" {
		// marital status can be empty
		return true
	}
	validStatuses := []string{"single", "married", "widowed", "divorced"}
	return Validator(status, validStatuses)
}

func ValidateSex(sex string) bool {
	validSexes := []string{"male", "female"}
	return Validator(sex, validSexes)
}

func ValidateSchool(school string) bool {
	validSchool := []string{"== primary", "== secondary"}
	return Validator(school, validSchool)
}

func ValidateApplicationStatus(status string) bool {
	validStatus := []string{config.StatusPending, config.StatusRejected, config.StatusApproved, config.StatusInProgress, config.StatusCompleted, config.StatusOnHold, config.StatusCancelled}
	return Validator(status, validStatus)
}

func ValidateHouseholdMembers(members []models.HouseholdMember) bool {
	for _, v := range members {
		if v.DateOfBirth == nil || v.Name == nil || v.Relation == nil {
			log.Printf("Household member form validate failed : %+v", v)
			return false
		}
	}

	return true
}

func Validator(value string, values []string) bool {
	for _, valid := range values {
		if strings.ToLower(value) == valid {
			return true
		}
	}

	log.Printf("Input validate failed => value : %+v matching : %+v", value, values)
	return false
}

func ValidateSchemeForm(scheme models.SchemeRequest) bool {
	if scheme.Name == "" {
		log.Printf("Scheme name is required")
		return false
	}

	if len(scheme.Criteria) == 0 {
		log.Printf("At least one criteria is required")
		return false
	}

	validCriteriaKeys := map[string]bool{
		"employment_status": true,
		"has_children":      true,
		"marital_status":    true,
		"sex":               true,
	}

	for _, v := range scheme.Criteria {
		for key := range v.Conditions {
			if !validCriteriaKeys[key] {
				log.Printf("Invalid criteria key: %s", key)
				return false
			}

			switch key {
			case "employment_status":
				if employmentStatus, ok := v.Conditions[key].(string); ok {
					if !ValidateEmploymentStatus(employmentStatus) {
						log.Printf("Invalid employment status")
						return false
					}
				}
			case "has_children":
				if hasChildren, ok := v.Conditions[key].(map[string]interface{}); ok {
					// only recognize school_level for now
					if schoolLevel, ok := hasChildren["school_level"].(string); ok {
						if !ValidateSchool(schoolLevel) {
							log.Printf("Invalid school")
							return false
						}
					} else {
						log.Printf("Invalid type for school_level: %+v", hasChildren)
						return false
					}
				}

			case "marital_status":
				if maritalStatus, ok := v.Conditions[key].(string); ok {
					if !ValidateMaritalStatus(maritalStatus) {
						log.Printf("Invalid marital status")
						return false
					}
				}
			case "sex":
				if sex, ok := v.Conditions[key].(string); ok {
					if !ValidateSex(sex) {
						log.Printf("Invalid sex")
						return false
					}
				}
			}
		}

		for _, benefit := range v.Benefits {
			if benefit.Name == "" || benefit.Amount <= 0 {
				log.Printf("Invalid benefit: %+v", benefit)
				return false
			}
		}
	}

	return true
}

func ValidateApplicationForm(application models.ApplicationRequest) bool {
	if application.ApplicantID == uuid.Nil || application.SchemeID == uuid.Nil {
		return false
	}
	return true
}
