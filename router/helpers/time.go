package helpers

import (
	"strings"
	"time"
)

var zonedTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999Z0700",
	"2006-01-02T15:04:05Z0700",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05.999999999Z0700",
	"2006-01-02 15:04:05Z0700",
}

var naiveDateTimeLayouts = []string{
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

func parseTime(value string) (time.Time, error) {
	value = normalizeTimeValue(value)
	for _, layout := range zonedTimeLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errInvalidTime
}

func parseNaiveTime(value string) (time.Time, error) {
	value = normalizeTimeValue(value)
	for _, layout := range naiveDateTimeLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errInvalidTime
}

func normalizeTimeValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= len("2006-01-02t15:04") && value[10] == 't' {
		value = value[:10] + "T" + value[11:]
	}
	if strings.HasSuffix(value, "z") {
		value = value[:len(value)-1] + "Z"
	}
	return value
}

type NaiveDateTime struct {
	Year       int
	Month      time.Month
	Day        int
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

type NullNaiveDateTime struct {
	DateTime NaiveDateTime
	Valid    bool
}

func parseNaiveDateTime(value string) (NaiveDateTime, error) {
	parsed, err := parseNaiveTime(value)
	if err != nil {
		return NaiveDateTime{}, err
	}
	return newNaiveDateTime(parsed), nil
}

func newNaiveDateTime(value time.Time) NaiveDateTime {
	year, month, day := value.Date()
	hour, minute, second := value.Clock()
	return NaiveDateTime{
		Year:       year,
		Month:      month,
		Day:        day,
		Hour:       hour,
		Minute:     minute,
		Second:     second,
		Nanosecond: value.Nanosecond(),
	}
}

func (value NaiveDateTime) TimeIn(location *time.Location) time.Time {
	if location == nil {
		location = time.Local
	}
	return time.Date(
		value.Year,
		value.Month,
		value.Day,
		value.Hour,
		value.Minute,
		value.Second,
		value.Nanosecond,
		location,
	)
}

func (value NaiveDateTime) String() string {
	return value.TimeIn(time.UTC).Format("2006-01-02T15:04:05.999999999")
}
