package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool

func Connect() {

	errr := godotenv.Load()
	if errr != nil {
		log.Fatal("ERROR: .env file is not found")
	}

	url := os.Getenv("DATABASE_URL")

	if url == "" {
		url = "postgres://postgres:12345@localhost:5432/order_app"
	}

	var err error
	Pool, err = pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal("ERROR: Unable to connect to the database", err)
	}

	fmt.Println("SUCCESS: Database connection established.")
}
