// SERVER

package handlers

import (
	"database/sql"
	"fmt"
	_ "github.com/joho/godotenv"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	CloudSQLDB *sql.DB
	QuestionDB *sql.DB
	TestDB     *sql.DB
	UserDataDB *sql.DB
)

// ConnectToDB establishes a connection to the specified Cloud SQL database
func ConnectToDB(dbName string) (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	instanceConnectionName := os.Getenv("INSTANCE_CONNECTION_NAME")

	// Cloud SQL connection using Unix socket for production
	dsn := fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?parseTime=true", dbUser, dbPassword, instanceConnectionName, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func InitDB() {
	var err error

	// Initialize Cloud SQL database
	CloudSQLDB, err = ConnectToDB(os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatalf("Failed to connect to Cloud SQL: %v", err)
	}
	CloudSQLDB.SetMaxOpenConns(20)
	CloudSQLDB.SetMaxIdleConns(10)

}

// local development

// package handlers

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"

// 	_ "github.com/go-sql-driver/mysql"
// 	"github.com/joho/godotenv"
// )

// var (
// 	CloudSQLDB *sql.DB
// 	QuestionDB *sql.DB
// 	TestDB     *sql.DB
// 	UserDataDB *sql.DB
// )

// // LoadEnv loads environment variables from the .env file
// func LoadEnv() {
// 	// Load the environment variables from .env file
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}
// }

// // ConnectToDB establishes a connection to the specified MySQL database
// func ConnectToDB(dbName string) (*sql.DB, error) {
// 	dbUser := os.Getenv("DB_USER")
// 	dbPassword := os.Getenv("DB_PASSWORD")
// 	dbHost := os.Getenv("DB_HOST")
// 	dbPort := os.Getenv("DB_PORT")

// 	// Connection using TCP (local development)
// 	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

// 	db, err := sql.Open("mysql", dsn)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open connection: %w", err)
// 	}

// 	// Check connection
// 	if err := db.Ping(); err != nil {
// 		return nil, fmt.Errorf("failed to ping database: %w", err)
// 	}

// 	return db, nil
// }

// // InitDB initializes all database connections for local development
// func InitDB() {
// 	// Load environment variables
// 	LoadEnv()

// 	var err error

// 	// Initialize Cloud SQL database
// 	CloudSQLDB, err = ConnectToDB(os.Getenv("DB_NAME"))
// 	if err != nil {
// 		log.Fatalf("Failed to connect to Cloud SQL: %v", err)
// 	}
// 	CloudSQLDB.SetMaxOpenConns(20)
// 	CloudSQLDB.SetMaxIdleConns(10)

// }
