package tests

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/angelbarreiros/Penguin/router/helpers"
)

func TestGetNullTimeQueryParam(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "zulu",
			value: "2024-01-02T03:04:05Z",
			want:  time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC),
		},
		{
			name:  "lowercase t and z",
			value: "2024-01-02t03:04:05z",
			want:  time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC),
		},
		{
			name:  "python iso offset with microseconds",
			value: "2024-01-02T03:04:05.123456+02:00",
			want:  time.Date(2024, time.January, 2, 3, 4, 5, 123456000, time.FixedZone("", 2*60*60)),
		},
		{
			name:  "space separator with offset",
			value: "2024-01-02 03:04:05.123456+02:00",
			want:  time.Date(2024, time.January, 2, 3, 4, 5, 123456000, time.FixedZone("", 2*60*60)),
		},
		{
			name:  "compact offset",
			value: "2024-01-02T03:04:05+0200",
			want:  time.Date(2024, time.January, 2, 3, 4, 5, 0, time.FixedZone("", 2*60*60)),
		},
		{
			name:    "rejects naive datetime",
			value:   "2024-01-02T03:04:05",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?createdAt="+url.QueryEscape(tt.value), nil)

			got, err := helpers.GetNullTimeQueryParam("createdAt", req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetNullTimeQueryParam() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !got.Valid {
				t.Fatal("GetNullTimeQueryParam() valid = false, want true")
			}
			if !got.Time.Equal(tt.want) {
				t.Fatalf("GetNullTimeQueryParam() time = %v, want %v", got.Time, tt.want)
			}
		})
	}
}

func TestGetNullNaiveDateTimeQueryParam(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    helpers.NaiveDateTime
		wantErr bool
	}{
		{
			name:  "full datetime",
			value: "2024-01-02T03:04:05",
			want:  naiveDateTime(2024, time.January, 2, 3, 4, 5, 0),
		},
		{
			name:  "microseconds",
			value: "2024-01-02T03:04:05.123456",
			want:  naiveDateTime(2024, time.January, 2, 3, 4, 5, 123456000),
		},
		{
			name:  "space separator",
			value: "2024-01-02 03:04:05",
			want:  naiveDateTime(2024, time.January, 2, 3, 4, 5, 0),
		},
		{
			name:  "minutes precision",
			value: "2024-01-02T03:04",
			want:  naiveDateTime(2024, time.January, 2, 3, 4, 0, 0),
		},
		{
			name:  "date only",
			value: "2024-01-02",
			want:  naiveDateTime(2024, time.January, 2, 0, 0, 0, 0),
		},
		{
			name:    "rejects zulu",
			value:   "2024-01-02T03:04:05Z",
			wantErr: true,
		},
		{
			name:    "rejects offset",
			value:   "2024-01-02T03:04:05+02:00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?createdAt="+url.QueryEscape(tt.value), nil)

			got, err := helpers.GetNullNaiveDateTimeQueryParam("createdAt", req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetNullNaiveDateTimeQueryParam() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !got.Valid {
				t.Fatal("GetNullNaiveDateTimeQueryParam() valid = false, want true")
			}
			if got.DateTime != tt.want {
				t.Fatalf("GetNullNaiveDateTimeQueryParam() datetime = %+v, want %+v", got.DateTime, tt.want)
			}
		})
	}
}

func TestGetNullNaiveDateTimePathValue(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.SetPathValue("createdAt", "2024-01-02 03:04:05.123456")

	got, err := helpers.GetNullNaiveDateTimePathValue("createdAt", req)
	if err != nil {
		t.Fatalf("GetNullNaiveDateTimePathValue() error = %v", err)
	}
	if !got.Valid {
		t.Fatal("GetNullNaiveDateTimePathValue() valid = false, want true")
	}

	want := naiveDateTime(2024, time.January, 2, 3, 4, 5, 123456000)
	if got.DateTime != want {
		t.Fatalf("GetNullNaiveDateTimePathValue() datetime = %+v, want %+v", got.DateTime, want)
	}
}

func naiveDateTime(year int, month time.Month, day int, hour int, minute int, second int, nanosecond int) helpers.NaiveDateTime {
	return helpers.NaiveDateTime{
		Year:       year,
		Month:      month,
		Day:        day,
		Hour:       hour,
		Minute:     minute,
		Second:     second,
		Nanosecond: nanosecond,
	}
}
