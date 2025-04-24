package dbhelper

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func (db *DBPlug) Connect() (*sql.DB, error) {
	var sb strings.Builder
	sb.WriteString("host=")
	sb.WriteString(db.Host)
	sb.WriteString(" port=")
	sb.WriteString(strconv.Itoa(db.Port))
	sb.WriteString(" user=")
	sb.WriteString(db.User)
	sb.WriteString(" password=")
	sb.WriteString(db.Password)
	sb.WriteString(" dbname=")
	sb.WriteString(db.Database)
	sb.WriteString(" sslmode=disable")

	connStr := sb.String()

	var conn *sql.DB
	var err error

	for i := range 5 {
		conn, err = sql.Open(db.Driver, connStr)
		if err == nil {
			err = conn.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Connection string: %s, Error: %v", connStr, err)
		waitTime := time.Duration(1<<i) * time.Second
		log.Printf("Retrying connection to database in %v seconds...", waitTime.Seconds())
		time.Sleep(waitTime)
	}

	if err != nil {
		return nil, fmt.Errorf("error connecting to database after retries: %v", err)
	}

	return conn, nil
}
