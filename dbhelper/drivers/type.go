package drivers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type ConnOption func(*sql.DB)
type DBPlug struct {
	host     string
	port     int
	user     string
	password string
	database string
	driver   string
	timeout  int
}

func NewDBPlug(host string, port int, user string, password string, database string, timeout int) *DBPlug {
	return &DBPlug{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
		timeout:  timeout,
	}
}
func (db *DBPlug) SetDriver(driver string) {
	db.driver = driver
}
func (db *DBPlug) Connect(opts ...ConnOption) (*sql.DB, error) {
	var sb strings.Builder
	sb.WriteString("host=")
	sb.WriteString(db.host)
	sb.WriteString(" port=")
	sb.WriteString(strconv.Itoa(db.port))
	sb.WriteString(" user=")
	sb.WriteString(db.user)
	sb.WriteString(" password=")
	sb.WriteString(db.password)
	sb.WriteString(" dbname=")
	sb.WriteString(db.database)
	sb.WriteString(" sslmode=disable")

	var connStr string = sb.String()

	var conn *sql.DB
	var err error

	for i := range 5 {
		conn, err = sql.Open(db.driver, connStr)
		if err == nil {
			for _, opt := range opts {
				opt(conn)
			}

			err = conn.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Connection string: %s, Error: %v", connStr, err)
		var waitTime time.Duration = time.Duration(1<<i) * time.Second
		log.Printf("Retrying connection to database in %v seconds...", waitTime.Seconds())
		time.Sleep(waitTime)
	}

	if err != nil {
		return nil, fmt.Errorf("error connecting to database after retries: %v", err)
	}

	return conn, nil
}

func WithMaxOpenConns(n int) ConnOption {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}

func WithMaxIdleConns(n int) ConnOption {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

func WithConnMaxLifetime(d time.Duration) ConnOption {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

func WithConnMaxIdleTime(d time.Duration) ConnOption {
	return func(db *sql.DB) {
		db.SetConnMaxIdleTime(d)
	}
}
