package api

import (
	"fmt"
	"strings"
	"time"
)

type CustomTime time.Time

const ctLayout = time.RFC3339

// UnmarshalJSON Parses the json string in the custom format
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		return
	}
	nt, err := time.Parse(ctLayout, s)
	*ct = CustomTime(nt)
	return
}

// MarshalJSON writes a quoted string in the custom format
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(ct.String()), nil
}

// String returns the time in the custom format
func (ct *CustomTime) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(ctLayout))
}

func StartOfDate(date time.Time, localeStr string) (out time.Time, err error) {
	var loc *time.Location
	if loc, err = time.LoadLocation(localeStr); err != nil {
		return
	}

	yr, mth, day := date.Date()
	out = time.Date(yr, mth, day, 0, 0, 0, 0, loc)
	return
}
