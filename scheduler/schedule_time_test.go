package scheduler

import (
	"testing"
	"time"
)

func TestNextRunKeepsWallClockAcrossDST(t *testing.T) {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Skip("tzdata not available")
	}
	// 2026-03-29 is the spring-forward day in Madrid.
	from := time.Date(2026, 3, 29, 0, 30, 0, 0, madrid)
	next := nextRunEveryXDaysAt(from, 1, 9*time.Hour, madrid)
	if next.Hour() != 9 || next.Day() != 29 {
		t.Fatalf("next run = %v, want 09:00 on the 29th", next)
	}
	after := nextRunAfterXDays(next, 1, 9*time.Hour, madrid)
	if after.Hour() != 9 || after.Day() != 30 {
		t.Fatalf("following run = %v, want 09:00 on the 30th", after)
	}
}

func TestNextRunUsesLocationOfAt(t *testing.T) {
	tokyo := time.FixedZone("UTC+9", 9*60*60)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // 09:00 in Tokyo
	next := nextRunEveryXDaysAt(from, 1, 10*time.Hour, tokyo)
	if want := time.Date(2026, 1, 1, 10, 0, 0, 0, tokyo); !next.Equal(want) {
		t.Fatalf("next run = %v, want %v", next, want)
	}
}
