package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	dbconn := os.Getenv("DB_CONNECTION")
	db, err := sql.Open("postgres", dbconn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	log.Println("Database connected successfully")
	return db, nil
}
