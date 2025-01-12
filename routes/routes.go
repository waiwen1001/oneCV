package routes

import (
	"database/sql"
	"oneCV/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine, db *sql.DB) {

	applicantController := &controllers.ApplicantController{DB: db}
	applicantionController := &controllers.ApplicantionController{DB: db}
	schemeController := &controllers.SchemeController{DB: db}

	api := router.Group("/api")
	// Applicant routes
	api.GET("/applicants", applicantController.GetAllApplicants)
	api.POST("/applicants", applicantController.CreateApplicant)
	api.GET("/applicants/:id", applicantController.GetApplicantByID)
	api.PUT("/applicants/:id", applicantController.UpdateApplicant)
	api.DELETE("/applicants/:id", applicantController.DeleteApplicant)

	// Scheme routes
	api.GET("/schemes", schemeController.GetAllSchemes)
	api.POST("/schemes", schemeController.CreateScheme)
	api.GET("/schemes/eligible", schemeController.GetEligibleSchemes)
	api.PUT("/schemes/:id", schemeController.UpdateScheme)
	api.DELETE("/schemes/:id", schemeController.DeleteScheme)

	// Application routes
	api.GET("/applications", applicantionController.GetAllApplications)
	api.POST("/applications", applicantionController.CreateApplication)
	api.PUT("/applications/:id", applicantionController.UpdateApplication)
	api.DELETE("/applications/:id", applicantionController.DeleteApplication)
}
