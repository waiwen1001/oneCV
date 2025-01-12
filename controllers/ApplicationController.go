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

type ApplicantionController struct {
	DB *sql.DB
}

// get all applications
func (ac *ApplicantionController) GetAllApplications(c *gin.Context) {
	ctx := c.Request.Context()
	application := models.Application{}
	applications, err := application.GetAllApplications(ctx, ac.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applications : " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"applications": applications})
}

// create applications
func (ac *ApplicantionController) CreateApplication(c *gin.Context) {
	ctx := c.Request.Context()
	applicationReq := models.ApplicationRequest{}
	if err := c.ShouldBind(&applicationReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create application : " + err.Error()})
		return
	}

	if formValidate := validator.ValidateApplicationForm(applicationReq); !formValidate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create application : " + config.REQUEST_FAILED})
		return
	}

	// check scheme exist
	scheme := models.Scheme{Id: applicationReq.SchemeID}
	err := scheme.CheckSchemeExist(ctx, ac.DB)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to create application : " + err.Error()})
		return
	}

	application := models.Application{}
	if err := application.CreateApplication(ctx, ac.DB, applicationReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.APPLICATION_SUBMIT_SUCCESS})
}

// update applications
func (ac *ApplicantionController) UpdateApplication(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update application : " + config.APPLICATION_ID_EMPTY})
		return
	}

	ctx := c.Request.Context()
	applicationId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update application : " + config.INVALID_APPLICATION_ID})
		return
	}

	applicationReq := models.ApplicationUpdateRequest{}
	if err := c.ShouldBind(&applicationReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update application : " + err.Error()})
		return
	}

	if formValidate := validator.ValidateApplicationStatus(applicationReq.Status); !formValidate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update application : " + config.REQUEST_FAILED})
		return
	}

	// check application exist
	application := models.Application{Id: applicationId}
	err = application.CheckApplicationExist(ctx, ac.DB)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to update application : " + err.Error()})
		return
	}

	if err := application.UpdateApplication(ctx, ac.DB, applicationReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application : " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": config.APPLICATION_SUBMIT_SUCCESS})
}

// delete application
func (ac *ApplicantionController) DeleteApplication(c *gin.Context) {
	aid := c.Param("id")
	if aid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete application : " + config.APPLICATION_ID_EMPTY})
		return
	}

	applicationId, err := uuid.Parse(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete application : " + config.INVALID_APPLICATION_ID})
		return
	}

	ctx := c.Request.Context()

	application := models.Application{Id: applicationId}
	if err := application.CheckApplicationExist(ctx, ac.DB); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete application : " + err.Error()})
		return
	}

	if err := application.DeleteApplication(ctx, ac.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application deleted successfully"})
}
