package helpers

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetNullNaiveTimeQueryParam(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantValid bool
		wantTime  time.Time
		wantErr   bool
	}{
		{
			name:      "missing",
			target:    "/test",
			wantValid: false,
		},
		{
			name:      "valid",
			target:    "/test?createdAt=2024-01-02T03:04:05",
			wantValid: true,
			wantTime:  time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC),
		},
		{
			name:      "valid with milliseconds",
			target:    "/test?createdAt=2024-01-02T03:04:05.123",
			wantValid: true,
			wantTime:  time.Date(2024, time.January, 2, 3, 4, 5, 123000000, time.UTC),
		},
		{
			name:    "rejects zoned time",
			target:  "/test?createdAt=2024-01-02T03:04:05Z",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.target, nil)

			got, err := GetNullNaiveTimeQueryParam("createdAt", req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetNullNaiveTimeQueryParam() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.Valid != tt.wantValid {
				t.Fatalf("GetNullNaiveTimeQueryParam() valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Valid && !got.Time.Equal(tt.wantTime) {
				t.Fatalf("GetNullNaiveTimeQueryParam() time = %v, want %v", got.Time, tt.wantTime)
			}
		})
	}
}

func TestGetNullNaiveTimePathValue(t *testing.T) {
	tests := []struct {
		name      string
		pathValue string
		wantValid bool
		wantTime  time.Time
		wantErr   bool
	}{
		{
			name:      "missing",
			wantValid: false,
		},
		{
			name:      "valid",
			pathValue: "2024-01-02T03:04:05",
			wantValid: true,
			wantTime:  time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC),
		},
		{
			name:      "valid with milliseconds",
			pathValue: "2024-01-02T03:04:05.123",
			wantValid: true,
			wantTime:  time.Date(2024, time.January, 2, 3, 4, 5, 123000000, time.UTC),
		},
		{
			name:      "rejects zoned time",
			pathValue: "2024-01-02T03:04:05Z",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.SetPathValue("createdAt", tt.pathValue)

			got, err := GetNullNaiveTimePathValue("createdAt", req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetNullNaiveTimePathValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.Valid != tt.wantValid {
				t.Fatalf("GetNullNaiveTimePathValue() valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Valid && !got.Time.Equal(tt.wantTime) {
				t.Fatalf("GetNullNaiveTimePathValue() time = %v, want %v", got.Time, tt.wantTime)
			}
		})
	}
}
