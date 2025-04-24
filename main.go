package main

import (
	"angelotero/commonBackend/dbhelper"
	"angelotero/commonBackend/dbhelper/drivers"

	_ "github.com/lib/pq"
)

func main() {
	var DBPlug dbhelper.DBPlug
	DBPlug.Host = "hh-pgsql-public.ebi.ac.uk"
	DBPlug.Port = 5432
	DBPlug.User = "reader"
	DBPlug.Password = "NWDMCE5xdipIjRrp"
	DBPlug.Database = "pfmegrnargs"
	DBPlug.Driver = "postgres"
	var psqlDriver drivers.PostgresDriver
	psqlDriver.SetDBPlug(&DBPlug)
	psqlDriver.Connect()

	defer psqlDriver.Close()

	type Rna struct {
		Id  int `db:"id"`
		Len int `db:"len"`
	}

	var rnas []Rna = make([]Rna, 0)

	err := psqlDriver.Select(&rnas, "SELECT * FROM rna LIMIT 10")
	if err != nil {
		panic(err)
	}

	for _, rna := range rnas {
		println(rna.Id, rna.Len)
	}
}
