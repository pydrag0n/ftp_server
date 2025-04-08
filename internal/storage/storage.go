package storage

import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

func CreateTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Table created!")
}

func InsertUser(db *sql.DB, username, password string) {
	stmt, err := db.Prepare("INSERT INTO users(username, password) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, password)
	if err != nil {
		log.Fatal(err)
	}

	id, _ := res.LastInsertId()
	log.Printf("User inserted with ID: %d", id)
}

func GetUsers(db *sql.DB) {
	rows, err := db.Query("SELECT id, username, password FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("User: %d, %s, %s", id, name, email)
	}
}

func InitDB(dbname string) *sql.DB {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	CreateTable(db)
	return db
}
