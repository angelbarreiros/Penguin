package drivers

import (
	"database/sql"

	"github.com/blockloop/scan/v2"
)

type PostgresDriver struct {
	dBConexion *sql.DB
	dBPlug     *DBPlug
}

func (p *PostgresDriver) Connect(opts ...ConnOption) error {
	db, err := p.dBPlug.Connect(opts...)
	if err != nil {
		return err
	}
	p.dBConexion = db

	return nil
}

func (p *PostgresDriver) Execute(query string, args any) (sql.Result, error) {
	return p.dBConexion.Exec(query, args)
}

func (p *PostgresDriver) Query(query string, args any) (*sql.Rows, error) {
	return p.dBConexion.Query(query, args)
}

func (p *PostgresDriver) Prepare(query string) (*sql.Stmt, error) {
	return p.dBConexion.Prepare(query)
}

func (p *PostgresDriver) SetDBPlug(plug *DBPlug) {
	p.dBPlug = plug
}

func (p *PostgresDriver) GetConexion() *sql.DB {
	return p.dBConexion
}

func (p *PostgresDriver) ScanRow(dest any, rows *sql.Rows) error {
	return scan.Row(dest, rows)
}

func (p *PostgresDriver) ScanRows(dest any, rows *sql.Rows) error {
	return scan.Rows(dest, rows)
}

func (p *PostgresDriver) Close() error {
	if p.dBConexion != nil {
		return p.dBConexion.Close()
	}
	return nil
}
