package config

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var Mastertxn map[string]bool

func InitDB() {
	var err error
	// Ganti dengan DSN milikmu
	DB, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/honor_dokter")
	if err != nil {
		log.Fatal(err)
	}
	DB.SetMaxOpenConns(30)
	DB.SetMaxIdleConns(30)
	DB.SetConnMaxLifetime(time.Minute * 5)

	if err := DB.Ping(); err != nil {
		log.Fatal("DB Ping error:", err)
	}

	loadMastertxn()
}

func loadMastertxn() {
	Mastertxn = make(map[string]bool)
	rows, err := DB.Query("SELECT txn_code FROM master_txn")
	if err != nil {
		log.Fatal("load master txn error:", err)
	}
	defer rows.Close()

	var txn_code string
	for rows.Next() {
		if err := rows.Scan(&txn_code); err == nil {
			Mastertxn[txn_code] = true
		}
	}
	log.Printf("Loaded %d master txn\n", len(Mastertxn))
}
