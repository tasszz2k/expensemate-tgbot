package types

import (
	"encoding/json"
	"time"

	"expensemate-tgbot/pkg/utils/timeutils"
)

type Id Signed

type Floating float64

func (f Floating) Float64() float64 {
	return float64(f)
}

func (f Floating) Float32() float32 {
	return float32(f)
}

type Signed int64

type Unsigned uint64

type Port Unsigned

func (u Unsigned) Uint32() uint32 {
	return uint32(u)
}

func (u Unsigned) Int() int {
	return int(u)
}

type KeyValue struct {
	Key   string
	Value string
}

type Time struct {
	time.Time
}

func NewTime(t time.Time) Time {
	return Time{Time: t}
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(timeutils.FormatAPI(t.Time))
}

func (t Time) UnmarshalJSON(data []byte) (err error) {
	t.Time, err = timeutils.ParseAPILocal(string(data))
	return
}
