package controllers

import (
	"database/sql"
	"net/http"
	"oneCV/config"
	"oneCV/models"
	"oneCV/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApplicantController struct {
	DB *sql.DB
}

// get all applicants
func (ac *ApplicantController) GetAllApplicants(c *gin.Context) {
	ctx := c.Request.Context()
	applicant := models.Applicant{}
	data, err := applicant.GetAllApplicants(ctx, ac.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applicants : " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"applicants": data})
}

// create applicant
func (ac *ApplicantController) CreateApplicant(c *gin.Context) {
	ctx := c.Request.Context()
	applicant := models.Applicant{}
	if err := c.ShouldBind(&applicant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create applicant : " + err.Error()})
		return
	}

	if formValidate := validator.ValidateApplicantForm(applicant); !formValidate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create applicant : " + config.REQUEST_FAILED})
		return
	}

	if err := applicant.CreateApplicant(ctx, ac.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create applicant : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.APPLICANT_SUBMIT_SUCCESS})
}

// get applicant by ID
func (ac *ApplicantController) GetApplicantByID(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get applicant : " + config.APPLICANT_ID_EMPTY})
		return
	}

	ctx := c.Request.Context()
	applicantId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get applicant : " + config.INVALID_APPLICANT_ID})
		return
	}
	applicant := models.Applicant{Id: applicantId}
	err = applicant.GetApplicantById(ctx, ac.DB)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to create applicant : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applicant": applicant})
}

// update applicant by ID
func (ac *ApplicantController) UpdateApplicant(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update applicant : " + config.APPLICANT_ID_EMPTY})
		return
	}

	ctx := c.Request.Context()
	applicantId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update applicant : " + config.INVALID_APPLICANT_ID})
		return
	}
	applicant := models.Applicant{Id: applicantId}
	if err := c.ShouldBind(&applicant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update applicant : " + err.Error()})
		return
	}

	if formValidate := validator.ValidateApplicantForm(applicant); !formValidate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update applicant : " + config.REQUEST_FAILED})
		return
	}

	// check does applicant Id exist
	err = applicant.CheckApplicantExist(ctx, ac.DB)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to update applicant : " + err.Error()})
		return
	}

	if err := applicant.UpdateApplicant(ctx, ac.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update applicant : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.APPLICANT_UPDATE_SUCCESS})
}

// delete applicant by Id
func (ac *ApplicantController) DeleteApplicant(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete applicant : " + config.APPLICANT_ID_EMPTY})
		return
	}

	applicantId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete applicant : " + config.INVALID_APPLICANT_ID})
		return
	}

	ctx := c.Request.Context()

	applicant := models.Applicant{Id: applicantId}
	if err := applicant.CheckApplicantExist(ctx, ac.DB); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete applicant : " + err.Error()})
		return
	}

	if err := applicant.DeleteApplicant(ctx, ac.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete applicant: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.APPLICANT_DELETE_SUCCESS})
}
