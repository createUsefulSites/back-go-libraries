package main

import (
	"log"

	"library-app/database"
	"library-app/models"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}

	database.InitDB()
}

func main() {
	err := models.MigrateSchema(database.DB)
	if err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}

	log.Println("Database schema migrated successfully!")
}
