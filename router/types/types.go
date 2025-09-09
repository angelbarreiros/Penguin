package types

import (
	"fmt"
	"time"
)

type BirthDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type NullBirthDate struct {
	Valid     bool      `json:"valid"`
	BirthDate BirthDate `json:"birth_date"`
}

func (b *BirthDate) Marshal() ([]byte, error) {
	return []byte(b.String()), nil

}
func (b *BirthDate) Unmarshal(data []byte) error {
	t, err := time.Parse("2006-01-02", string(data))
	if err != nil {
		return err
	}
	b.Year = t.Year()
	b.Month = int(t.Month())
	b.Day = t.Day()
	return nil
}

func (b *BirthDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", b.Year, b.Month, b.Day)
}
