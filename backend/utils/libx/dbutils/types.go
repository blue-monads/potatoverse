package dbutils

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

const layout = "2006-01-02 15:04:05.999999999-07:00"

type Time struct {
	Inner time.Time
}

func (mt Time) Value() (driver.Value, error) {
	return mt.Inner, nil
}

func (mt *Time) Scan(val any) error {

	switch vt := val.(type) {
	case string:
		t, err := time.Parse(layout, vt)
		if err != nil {
			return err
		}
		mt.Inner = t
	case time.Time:
		mt.Inner = vt
	}

	return nil
}

func (mt *Time) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &mt.Inner)
}

func (mt Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(&mt.Inner)
}
