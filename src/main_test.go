package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func TestDatabaseConnection(t *testing.T) {
	dsn := "user=admin password=root dbname=mydatabase sslmode=disable host=localhost port=5433"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database connection: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}
