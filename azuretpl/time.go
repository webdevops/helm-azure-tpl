package azuretpl

import (
	"fmt"
	"strconv"
	"time"
)

// fromUnixtime converts unixtime to time.Time object
func fromUnixtime(val interface{}) (time.Time, error) {
	var timestamp int64
	switch v := val.(type) {
	// int
	case *int:
		if v == nil {
			return time.Unix(0, 0), fmt.Errorf(`cannot convert nil to time`)
		}
		timestamp = int64(*v)
	case int:
		timestamp = int64(v)

	// int64
	case *int64:
		if v == nil {
			return time.Unix(0, 0), fmt.Errorf(`cannot convert nil to time`)
		}
		timestamp = *v
	case int64:
		timestamp = int64(v)

	// float32
	case *float32:
		if v == nil {
			return time.Unix(0, 0), fmt.Errorf(`cannot convert nil to time`)
		}
		timestamp = int64(*v)
	case float32:
		timestamp = int64(v)

	// float64
	case *float64:
		if v == nil {
			return time.Unix(0, 0), fmt.Errorf(`cannot convert nil to time`)
		}
		timestamp = int64(*v)
	case float64:
		timestamp = int64(v)

	// string
	case *string:
		if v == nil {
			return time.Unix(0, 0), fmt.Errorf(`cannot convert nil to time`)
		}
		var err error
		timestamp, err = strconv.ParseInt(*v, 10, 64)
		if err != nil {
			return time.Unix(0, 0), err
		}
	case string:
		var err error
		timestamp, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return time.Unix(0, 0), err
		}

	// time.Time
	case *time.Time:
		if v != nil {
			return *v, nil
		}
		return time.Unix(0, 0), fmt.Errorf(`cannot convert nil to time`)
	case time.Time:
		return v, nil

	// default
	default:
		return time.Unix(0, 0), fmt.Errorf(`invalid unixtimestamp "%v" defined, must be int or float`, val)
	}

	return time.Unix(timestamp, 0), nil
}

// toRFC3339 converts time.Time object to RFC 3339 string
func toRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}
