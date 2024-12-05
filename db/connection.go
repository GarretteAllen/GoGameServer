package db

import (
	"database/sql"
	"log"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

// initialize database connection
func Init() {
	once.Do(func() {
		var err error
		db, err = sql.Open("sqlite", "./game.db")
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// test the connection
		if err := db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		log.Println("Database connection initialized")
	})
}

// return initialized database connection
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("Database is not initialized. Call db.Init() first.")
	}
	return db
}
