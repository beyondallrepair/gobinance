package gobinance

import (
	"encoding/json"
	"time"
)

// millisTimestamp represents a unix timestamp given in milliseconds
type millisTimestamp time.Time

// UnmarshalJSON converts an integer represented as json into a millisTimestamp
func (t *millisTimestamp) UnmarshalJSON(bs []byte) error {
	var millis int64
	if err := json.Unmarshal(bs, &millis); err != nil {
		return err
	}
	*t = millisTimestamp(millisToTime(millis))
	return nil
}

// millisToTime converts an integer number of milliseconds since the unix epoch into
// a time.Time
func millisToTime(millis int64) time.Time {
	return time.Unix(0,millis*int64(time.Millisecond)).UTC()
}
