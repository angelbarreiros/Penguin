package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidDateRange = errors.New("start date must be before or equal to end date")

const (
	postgreSQLDateLayout      = "2006-01-02"
	postgreSQLTimeLayout      = "2006-01-02T15:04:05"
	postgreSQLTimestampFormat = "2006-01-02 15:04:05.000000-07:00"
	// postgreSQLTimestampParse accepts 0-9 fractional digits and offsets
	// written as -07, -07:00 or -0700.
	postgreSQLTimestampParse = "2006-01-02 15:04:05.999999999"
)

func unmarshalJSONString(data []byte) (string, bool, error) {
	if string(data) == "null" {
		return "", true, nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return "", false, err
	}
	return s, false, nil
}

type PostgreSQLDate struct {
	Year  int
	Month int
	Day   int
}

func (d PostgreSQLDate) ToString() string {
	t := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	return t.Format(postgreSQLDateLayout)
}
func (d PostgreSQLDate) Marshal() ([]byte, error) {
	return []byte(d.ToString()), nil
}

func (d *PostgreSQLDate) Unmarshal(data []byte) error {
	t, err := time.Parse(postgreSQLDateLayout, string(data))
	if err != nil {
		return err
	}
	d.Year = t.Year()
	d.Month = int(t.Month())
	d.Day = t.Day()
	return nil
}

func (d PostgreSQLDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.ToString())
}

func (d *PostgreSQLDate) UnmarshalJSON(data []byte) error {
	s, isNull, err := unmarshalJSONString(data)
	if err != nil || isNull {
		return err
	}
	return d.Unmarshal([]byte(s))
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
	return t.Format(postgreSQLTimeLayout)
}

func (pst PostgreSQLTime) Marshal() ([]byte, error) {
	return []byte(pst.ToString()), nil
}

func (pst *PostgreSQLTime) Unmarshal(data []byte) error {
	t, err := time.Parse(postgreSQLTimeLayout, string(data))
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
	s, isNull, err := unmarshalJSONString(data)
	if err != nil || isNull {
		return err
	}
	return pst.Unmarshal([]byte(s))
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
	// Timezone is a UTC offset such as "+02", "-05:30" or "+00". Empty means UTC.
	Timezone string
}

func parseOffset(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}
	for _, layout := range []string{"-07", "-07:00", "-0700", "Z07:00"} {
		if t, err := time.Parse(layout, tz); err == nil {
			_, offset := t.Zone()
			return time.FixedZone("", offset), nil
		}
	}
	return nil, fmt.Errorf("invalid timezone offset %q", tz)
}

func (t PostgreSQLTimeTimeStamp) ToString() string {
	loc, err := parseOffset(t.Timezone)
	if err != nil {
		loc = time.UTC
	}
	tm := time.Date(t.Year, time.Month(t.Month), t.Day, t.Hour, t.Minute, t.Second, t.Microsecond*1000, loc)
	return tm.Format(postgreSQLTimestampFormat)
}
func (t PostgreSQLTimeTimeStamp) Marshal() ([]byte, error) {
	if _, err := parseOffset(t.Timezone); err != nil {
		return nil, err
	}
	return []byte(t.ToString()), nil
}

func (t *PostgreSQLTimeTimeStamp) Unmarshal(data []byte) error {
	value := string(data)
	var tm time.Time
	var err error
	for _, offsetLayout := range []string{"-07:00", "-0700", "-07"} {
		tm, err = time.Parse(postgreSQLTimestampParse+offsetLayout, value)
		if err == nil {
			break
		}
	}
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
	t.Timezone = tm.Format("-07:00")
	return nil
}

func (t PostgreSQLTimeTimeStamp) MarshalJSON() ([]byte, error) {
	if _, err := parseOffset(t.Timezone); err != nil {
		return nil, err
	}
	return json.Marshal(t.ToString())
}

func (t *PostgreSQLTimeTimeStamp) UnmarshalJSON(data []byte) error {
	s, isNull, err := unmarshalJSONString(data)
	if err != nil || isNull {
		return err
	}
	return t.Unmarshal([]byte(s))
}

// DateRange is inclusive on both ends for Contains; Overlaps treats ranges
// as half-open, so ranges that only touch at an endpoint do not overlap.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange creates a new DateRange with validation
func NewDateRange(start, end time.Time) (*DateRange, error) {
	if start.After(end) {
		return nil, ErrInvalidDateRange
	}
	return &DateRange{Start: start, End: end}, nil
}

// Duration returns the duration between start and end
func (dr DateRange) Duration() time.Duration {
	return dr.End.Sub(dr.Start)
}

// Days returns the number of calendar days between Start and End, evaluated
// in Start's location, so DST changes do not shorten the count.
func (dr DateRange) Days() int {
	loc := dr.Start.Location()
	sy, sm, sd := dr.Start.Date()
	ey, em, ed := dr.End.In(loc).Date()
	start := time.Date(sy, sm, sd, 0, 0, 0, 0, time.UTC)
	end := time.Date(ey, em, ed, 0, 0, 0, 0, time.UTC)
	return int(end.Sub(start).Hours() / 24)
}

// Contains checks if a date is within the range (inclusive)
func (dr DateRange) Contains(t time.Time) bool {
	return !t.Before(dr.Start) && !t.After(dr.End)
}

// Overlaps checks if two date ranges overlap (half-open: touching ranges do not overlap)
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

// Split splits the date range into chunks of specified days. It returns nil
// when days <= 0.
func (dr DateRange) Split(days int) []DateRange {
	if days <= 0 {
		return nil
	}
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
