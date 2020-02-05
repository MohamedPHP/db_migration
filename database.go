package main

import (
	"database/sql"
	"os"

	// Import Mysql DB Driver
	_ "github.com/go-sql-driver/mysql"

	// Autoload The .env File
	_ "github.com/joho/godotenv/autoload"
)

// ConnectFrom To Database
func ConnectFrom() (db *sql.DB) {
	db, err := sql.Open(os.Getenv("DB_CONNECTION"), os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@/"+os.Getenv("DB_DATABASE"))

	if err != nil {
		panic(err.Error())
	}

	return db
}

// ConnectTo To Database
func ConnectTo() (db *sql.DB) {
	db, err := sql.Open(os.Getenv("DB_CONNECTION"), os.Getenv("DB_USERNAME_2")+":"+os.Getenv("DB_PASSWORD_2")+"@/"+os.Getenv("DB_DATABASE_2"))

	if err != nil {
		panic(err.Error())
	}

	return db
}
