package config

import (
	"fmt"
	"go-api-kbt/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error

	// Muat variabel lingkungan dari file .env
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Membaca variabel lingkungan
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	// Memeriksa jika variabel lingkungan yang penting tidak diatur
	if host == "" || user == "" || password == "" || dbname == "" || port == "" {
		fmt.Println("Error: One or more required environment variables are not set")
		return
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto Migrate models
	err = DB.AutoMigrate(&models.Event{}, &models.User{}, &models.Location{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// // Check if tables exist
	// checkTables(DB)

	if DB == nil {
		log.Fatal("DB connection is nil")
	}

	log.Println("Database connected")
}

// func checkTables(db *gorm.DB) {
// 	var tableNames []string
// 	rows, err := db.Raw("SELECT event FROM information_schema.tables WHERE table_schema = 'public'").Rows()
// 	if err != nil {
// 		log.Fatal("Failed to query database schema:", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var tableName string
// 		if err := rows.Scan(&tableName); err != nil {
// 			log.Fatal("Failed to scan table name:", err)
// 		}
// 		tableNames = append(tableNames, tableName)
// 	}

// 	// Print or log table names
// 	fmt.Println("Tables in the database:")
// 	for _, name := range tableNames {
// 		fmt.Println(name)
// 	}
// }
