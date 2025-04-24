package dbhelper

import (
	"database/sql"
)

type DBPlug struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Driver   string `json:"driver"`
	Timeout  int    `json:"timeout"`
}
type Driver interface {
	Connect() error
	Disconnect() error
	Execute(query string, args any) (sql.Result, error)
	Query(query string, args any) (*sql.Rows, error)
	Prepare(query string) (*sql.Stmt, error)
	SetDBPlug(plug *DBPlug)
	GetConexion() *sql.DB
	Close() error
}
