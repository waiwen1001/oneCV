package main

import (
	"log"
	"oneCV/config"
	"oneCV/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Init Database failed: %v", err)
	}

	defer db.Close()

	r := gin.Default()
	routes.InitRoutes(r, db)

	r.Run(":8080")
}
