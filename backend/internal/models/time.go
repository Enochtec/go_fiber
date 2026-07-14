package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Time struct {
	time.Time
}

func (t *Time) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		for _, layout := range []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
			time.RFC3339,
		} {
			if parsed, err := time.Parse(layout, v); err == nil {
				t.Time = parsed
				return nil
			}
		}
		return fmt.Errorf("cannot parse time string: %s", v)
	case int64:
		t.Time = time.Unix(v, 0)
		return nil
	case []byte:
		return t.Scan(string(v))
	default:
		return fmt.Errorf("unsupported time type: %T", value)
	}
}

func (t Time) Value() (driver.Value, error) {
	return t.Format("2006-01-02 15:04:05"), nil
}
