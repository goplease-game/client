package game

import (
	"time"
)

// DisplayDateFromRFC parses an RFC3339 timestamp and formats it
// as a human-readable date (e.g. "July 2, 2026").
func DisplayDateFromRFC(v string) (string, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return "", err
	}

	return t.Format("January 2, 2006"), nil
}
