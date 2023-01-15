package main

import (
	"database/sql"
)

func tradeExists(db *sql.DB, id string) bool {
	sqlStmt := `SELECT count(*) WHERE id=$1`
	row := db.QueryRow(sqlStmt, id)
	var count int
	row.Scan(&count)
	return count == 1
}
