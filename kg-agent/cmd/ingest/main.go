package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Unable to load env variables")
	}

	ctx := context.Background()

	config := database.Config{
		Host:     os.Getenv("KG_AGENT_VECTOR_DB_HOST"),
		Port:     os.Getenv("KG_AGENT_VECTOR_DB_PORT"),
		User:     os.Getenv("KG_AGENT_VECTOR_DB_USER"),
		Password: os.Getenv("KG_AGENT_VECTOR_DB_PASSWORD"),
		Database: os.Getenv("KG_AGENT_VECTOR_DB_DATABASE"),
		SSLMode:  os.Getenv("KG_AGENT_VECTOR_DB_SSLMode"),
	}

	db, err := database.New(ctx, config)
	if err != nil {
		log.Fatal("Failed to connect to database. Error: %w", err)
	}

	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Database failed: %s", err)
	}

	fmt.Println("Connected successfully")

}
