package types

import "time"

type Date struct {
	Year  int
	Month int
	Day   int
}

func (d Date) ToString() string {
	t := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	return t.Format("2006-01-02")
}
func (d Date) Marshal() ([]byte, error) {
	return []byte(d.ToString()), nil
}

func (d *Date) Unmarshal(data []byte) error {
	t, err := time.Parse("2006-01-02", string(data))
	if err != nil {
		return err
	}
	d.Year = t.Year()
	d.Month = int(t.Month())
	d.Day = t.Day()
	return nil
}

type TimeStamp struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Microsecond int
	Timezone    string
}

func (t TimeStamp) ToString() string {
	tm := time.Date(t.Year, time.Month(t.Month), t.Day, t.Hour, t.Minute, t.Second, t.Microsecond*1000, time.UTC)
	return tm.Format("2006-01-02 15:04:05.000000-07")
}
func (t TimeStamp) Marshal() ([]byte, error) {
	return []byte(t.ToString()), nil
}

func (t *TimeStamp) Unmarshal(data []byte) error {
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

type Time struct {
	Hour       int
	Minute     int
	Second     int
	Milisecond int
}

func (t Time) ToString() string {
	tm := time.Date(0, 1, 1, t.Hour, t.Minute, t.Second, t.Milisecond*1000000, time.UTC)
	return tm.Format("15:04:05.000")
}
func (t Time) Marshal() ([]byte, error) {
	return []byte(t.ToString()), nil
}

func (t *Time) Unmarshal(data []byte) error {
	tm, err := time.Parse("15:04:05.000", string(data))
	if err != nil {
		return err
	}
	t.Hour = tm.Hour()
	t.Minute = tm.Minute()
	t.Second = tm.Second()
	t.Milisecond = tm.Nanosecond() / 1000000
	return nil
}
