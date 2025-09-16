package types

import (
	"encoding/json"
	"strings"
	"time"
)

type PostgreSQLDate struct {
	Year  int
	Month int
	Day   int
}

func (d PostgreSQLDate) ToString() string {
	t := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	return t.Format("2006-01-02")
}
func (d PostgreSQLDate) Marshal() ([]byte, error) {
	return []byte(d.ToString()), nil
}

func (d *PostgreSQLDate) Unmarshal(data []byte) error {
	t, err := time.Parse("2006-01-02", string(data))
	if err != nil {
		return err
	}
	d.Year = t.Year()
	d.Month = int(t.Month())
	d.Day = t.Day()
	return nil
}

type PostgreSQLTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

func (pst PostgreSQLTime) ToString() string {
	t := time.Date(pst.Year, time.Month(pst.Month), pst.Day, pst.Hour, pst.Minute, pst.Second, 0, time.UTC)
	return t.Format("2006-01-02T15:04:05")
}

func (pst PostgreSQLTime) Marshal() ([]byte, error) {
	return []byte(pst.ToString()), nil
}

func (pst *PostgreSQLTime) Unmarshal(data []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05", string(data))
	if err != nil {
		return err
	}
	pst.Year = t.Year()
	pst.Month = int(t.Month())
	pst.Day = t.Day()
	pst.Hour = t.Hour()
	pst.Minute = t.Minute()
	pst.Second = t.Second()
	return nil
}

func (pst *PostgreSQLTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// Remove quotes
	str := strings.Trim(string(data), `"`)

	// Parse PostgreSQL timestamp format
	t, err := time.Parse("2006-01-02T15:04:05", str)
	if err != nil {
		return err
	}

	pst.Year = t.Year()
	pst.Month = int(t.Month())
	pst.Day = t.Day()
	pst.Hour = t.Hour()
	pst.Minute = t.Minute()
	pst.Second = t.Second()
	return nil
}

func (pst PostgreSQLTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(pst.ToString())
}

type PostgreSQLTimeTimeStamp struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Microsecond int
	Timezone    string
}

func (t PostgreSQLTimeTimeStamp) ToString() string {
	tm := time.Date(t.Year, time.Month(t.Month), t.Day, t.Hour, t.Minute, t.Second, t.Microsecond*1000, time.UTC)
	return tm.Format("2006-01-02 15:04:05.000000-07")
}
func (t PostgreSQLTimeTimeStamp) Marshal() ([]byte, error) {
	return []byte(t.ToString()), nil
}

func (t *PostgreSQLTimeTimeStamp) Unmarshal(data []byte) error {
	tm, err := time.Parse("2006-01-02 15:04:05.000000-07", string(data))
	if err != nil {
		return err
	}
	t.Year = tm.Year()
	t.Month = int(tm.Month())
	t.Day = tm.Day()
	t.Hour = tm.Hour()
	t.Minute = tm.Minute()
	t.Second = tm.Second()
	t.Microsecond = tm.Nanosecond() / 1000
	t.Timezone = tm.Format("-07")
	return nil
}
