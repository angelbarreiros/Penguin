package dbhelper

import (
	"angelotero/commonBackend/dbhelper/drivers"
	"database/sql"
)

type Driver interface {
	Connect(opts ...drivers.ConnOption) error
	Execute(query string, args any) (sql.Result, error)
	Query(query string, args any) (*sql.Rows, error)
	Prepare(query string) (*sql.Stmt, error)
	SetDBPlug(plug *drivers.DBPlug)
	GetConexion() *sql.DB
	ScanRow(dest any, rows *sql.Rows) error
	ScanRows(dest any, rows *sql.Rows) error
	Close() error
}

func NewConnection(driverType string, plug *drivers.DBPlug) Driver {
	var driver Driver

	switch driverType {
	case "postgres":
		driver = &drivers.PostgresDriver{}

	default:
		driver = &drivers.PostgresDriver{}
	}
	plug.SetDriver(driverType)
	driver.SetDBPlug(plug)
	return driver
}
