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

type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange creates a new DateRange with validation
func NewDateRange(start, end time.Time) (*DateRange, error) {
	if start.After(end) {
		return nil, json.Unmarshal([]byte(`"start date must be before or equal to end date"`), new(string))
	}
	return &DateRange{Start: start, End: end}, nil
}

// Duration returns the duration between start and end
func (dr DateRange) Duration() time.Duration {
	return dr.End.Sub(dr.Start)
}

// Days returns the number of days in the range
func (dr DateRange) Days() int {
	return int(dr.Duration().Hours() / 24)
}

// Contains checks if a date is within the range (inclusive)
func (dr DateRange) Contains(t time.Time) bool {
	return (t.Equal(dr.Start) || t.After(dr.Start)) && (t.Equal(dr.End) || t.Before(dr.End))
}

// Overlaps checks if two date ranges overlap
func (dr DateRange) Overlaps(other DateRange) bool {
	return dr.Start.Before(other.End) && dr.End.After(other.Start)
}

// IsValid checks if the date range is valid
func (dr DateRange) IsValid() bool {
	return !dr.Start.After(dr.End)
}

// String returns a string representation of the date range
func (dr DateRange) String() string {
	return dr.Start.Format("2006-01-02") + " to " + dr.End.Format("2006-01-02")
}

// MarshalJSON implements json.Marshaler
func (dr DateRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"start": dr.Start.Format("2006-01-02"),
		"end":   dr.End.Format("2006-01-02"),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (dr *DateRange) UnmarshalJSON(data []byte) error {
	var aux struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	start, err := time.Parse("2006-01-02", aux.Start)
	if err != nil {
		return err
	}

	end, err := time.Parse("2006-01-02", aux.End)
	if err != nil {
		return err
	}

	dr.Start = start
	dr.End = end
	return nil
}

// Split splits the date range into chunks of specified days
func (dr DateRange) Split(days int) []DateRange {
	var ranges []DateRange
	current := dr.Start

	for current.Before(dr.End) {
		next := current.AddDate(0, 0, days)
		if next.After(dr.End) {
			next = dr.End
		}
		ranges = append(ranges, DateRange{Start: current, End: next})
		current = next
	}

	return ranges
}

// Extend extends the date range by adding days to the end
func (dr *DateRange) Extend(days int) {
	dr.End = dr.End.AddDate(0, 0, days)
}

// Shift shifts both start and end dates by the specified days
func (dr *DateRange) Shift(days int) {
	dr.Start = dr.Start.AddDate(0, 0, days)
	dr.End = dr.End.AddDate(0, 0, days)
}
