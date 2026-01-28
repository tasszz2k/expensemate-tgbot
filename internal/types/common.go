package types

import (
	"encoding/json"
	"time"
)

// ID represents a unique identifier (signed int64)
type ID int64

// Unsigned represents an unsigned integer value
type Unsigned uint64

func (u Unsigned) Uint32() uint32 {
	return uint32(u)
}

func (u Unsigned) Int() int {
	return int(u)
}

// Floating represents a float64 value
type Floating float64

func (f Floating) Float64() float64 {
	return float64(f)
}

func (f Floating) Float32() float32 {
	return float32(f)
}

// Time wraps time.Time with custom JSON marshaling
type Time struct {
	time.Time
}

// NewTime creates a new Time from time.Time
func NewTime(t time.Time) Time {
	return Time{Time: t}
}

// MarshalJSON implements json.Marshaler
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format("02/01/2006 15:04:05"))
}

// UnmarshalJSON implements json.Unmarshaler
func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse("02/01/2006 15:04:05", s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}
