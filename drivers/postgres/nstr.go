package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type NullString struct {
	String string
	Valid  bool // Valid is true if String is not NULL
}

func (f NullString) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.String)
}

func (f *NullString) UnmarshalJSON(b []byte) error {
	if f == nil {
		return errors.New("NullString: UnmarshalJSON on nil pointer")
	}

	if err := json.Unmarshal(b, &f.String); err != nil {
		return err
	}
	f.Valid = true
	return nil
}

type NullStringArray []NullString

func (f *NullStringArray) Scan(value interface{}) error {
	if value == nil {
		*f = make([]NullString, 0)
		return nil
	}

	switch val := value.(type) {
	case []byte:
		return json.Unmarshal(val, f)
	}

	return nil
}

func (f NullStringArray) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f NullStringArray) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("[]"), nil
	}

	var v []NullString = f

	return json.Marshal(v)
}

func (f *NullStringArray) UnmarshalJSON(b []byte) error {
	if f == nil {
		return errors.New("NullStringArray: UnmarshalJSON on nil pointer")
	}

	var v []NullString
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	*f = v

	return nil
}
