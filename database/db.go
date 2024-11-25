package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var Db *sql.DB

func InitDB() error {
	// ①-1
	//DB接続のための準備

	/*
		mysqlUser := os.Getenv("MYSQL_USER")
		mysqlUserPwd := os.Getenv("MYSQL_PWD")
		mysqlHost := os.Getenv("MYSQL_HOST")
		mysqlDatabase := os.Getenv("MYSQL_DATABASE")
		log.Printf("mysqlUser: %s, mysqlUserPwd: %s, mysqlHost: %s, mysqlDatabase: %s", mysqlUser, mysqlUserPwd, mysqlHost, mysqlDatabase)

		connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlUserPwd, mysqlHost, mysqlDatabase)
		_db, err := sql.Open("mysql", connStr)
	*/

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlUserPwd := os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	if mysqlUser == "" || mysqlUserPwd == "" || mysqlDatabase == "" {
		log.Fatal("Environment variables MYSQL_USER, MYSQL_PASSWORD, and MYSQL_DATABASE must be set")
	}

	// ①-2
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(localhost:3306)/%s", mysqlUser, mysqlUserPwd, mysqlDatabase))

	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	// ①-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	Db = _db
	return nil
}