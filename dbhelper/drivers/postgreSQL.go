package drivers

import (
	"angelotero/commonBackend/dbhelper"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresDriver struct {
	dBConexion *sqlx.DB
	dBPlug     *dbhelper.DBPlug
}

func (p *PostgresDriver) SetDBPlug(plug *dbhelper.DBPlug) {
	p.dBPlug = plug
}

func (p PostgresDriver) GetConexion() *sqlx.DB {
	return p.dBConexion
}

func (p *PostgresDriver) Connect() error {
	var err error
	rawDB, err := p.dBPlug.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	p.dBConexion = sqlx.NewDb(rawDB, "postgres")
	return nil
}

func (p *PostgresDriver) Close() error {
	if p.dBConexion != nil {
		return p.dBConexion.Close()
	}
	return fmt.Errorf("no connection to close")
}
func (p PostgresDriver) Select(dest interface{}, query string, args ...interface{}) error {
	if p.dBConexion == nil {
		return fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Select(dest, query, args...)
}

func (p PostgresDriver) Get(dest interface{}, query string, args ...interface{}) error {
	if p.dBConexion == nil {
		return fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Get(dest, query, args...)
}

func (p PostgresDriver) Insert(query string, args ...interface{}) (sql.Result, error) {
	if p.dBConexion == nil {
		return nil, fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Exec(query, args...)
}

func (p PostgresDriver) Update(query string, args ...interface{}) (sql.Result, error) {
	if p.dBConexion == nil {
		return nil, fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Exec(query, args...)
}

func (p PostgresDriver) Delete(query string, args ...interface{}) (sql.Result, error) {
	if p.dBConexion == nil {
		return nil, fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Exec(query, args...)
}
func (p PostgresDriver) Execute(query string, args ...interface{}) (sql.Result, error) {
	if p.dBConexion == nil {
		return nil, fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Exec(query, args...)
}

func (p PostgresDriver) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	if p.dBConexion == nil {
		return nil, fmt.Errorf("no connection to execute query")
	}
	return p.dBConexion.Queryx(query, args...)
}

func (p PostgresDriver) Prepare(query string) (*sqlx.Stmt, error) {
	if p.dBConexion == nil {
		return nil, fmt.Errorf("no connection to prepare statement")
	}
	return p.dBConexion.Preparex(query)
}
