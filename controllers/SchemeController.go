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

type SchemeController struct {
	DB *sql.DB
}

// get all schemes
func (sc *SchemeController) GetAllSchemes(c *gin.Context) {
	ctx := c.Request.Context()
	scheme := models.Scheme{}
	data, err := scheme.GetAllSchemes(ctx, sc.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get schemes : " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schemes": data})
}

// create scheme
func (sc *SchemeController) CreateScheme(c *gin.Context) {
	ctx := c.Request.Context()
	var schemeReq models.SchemeRequest

	if err := c.ShouldBindJSON(&schemeReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create scheme : " + err.Error()})
		return
	}

	if formValidate := validator.ValidateSchemeForm(schemeReq); !formValidate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create scheme : " + config.REQUEST_FAILED})
		return
	}

	scheme := models.Scheme{}
	if err := scheme.CreateScheme(ctx, sc.DB, schemeReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create scheme : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.SCHEME_SUBMIT_SUCCESS})
}

// get eligible schemes
func (sc *SchemeController) GetEligibleSchemes(c *gin.Context) {
	ctx := c.Request.Context()
	aid := c.DefaultQuery("applicant", "")

	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get eligible scheme : " + config.APPLICANT_ID_EMPTY})
		return
	}

	applicantId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get eligible scheme : " + config.INVALID_APPLICANT_ID})
		return
	}

	applicant := models.Applicant{Id: applicantId}
	if err := applicant.GetApplicantById(ctx, sc.DB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get eligible scheme : " + config.APPLICANT_NOT_FOUND})
		return
	}

	scheme := models.Scheme{}
	schemes, err := scheme.GetEligibleSchemes(ctx, sc.DB, applicant)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get eligible scheme : " + err.Error()})
		return
	}

	if len(schemes) == 0 {
		c.JSON(http.StatusOK, gin.H{"scheme": schemes, "message": config.NO_ELIGIBLE_SCHEME})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scheme": schemes})
}

// update scheme
func (sc *SchemeController) UpdateScheme(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update scheme : " + config.SCHEME_ID_EMPTY})
		return
	}

	ctx := c.Request.Context()
	schemeId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update scheme : " + config.INVALID_SCHEME_ID})
		return
	}
	var schemeReq models.SchemeRequest
	scheme := models.Scheme{Id: schemeId}
	if err := c.ShouldBind(&schemeReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update scheme : " + err.Error()})
		return
	}

	if formValidate := validator.ValidateSchemeForm(schemeReq); !formValidate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update scheme : " + config.REQUEST_FAILED})
		return
	}

	// check does scheme Id exist
	err = scheme.CheckSchemeExist(ctx, sc.DB)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to update scheme : " + err.Error()})
		return
	}

	if err := scheme.UpdateScheme(ctx, sc.DB, schemeReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheme : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.SCHEME_UPDATE_SUCCESS})
}

// delete scheme
func (sc *SchemeController) DeleteScheme(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete scheme : " + config.SCHEME_ID_EMPTY})
		return
	}

	schemeId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete scheme : " + config.INVALID_SCHEME_ID})
		return
	}

	ctx := c.Request.Context()

	scheme := models.Scheme{Id: schemeId}
	if err := scheme.CheckSchemeExist(ctx, sc.DB); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete scheme : " + err.Error()})
		return
	}

	if err := scheme.DeleteScheme(ctx, sc.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheme: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheme deleted successfully"})
}
