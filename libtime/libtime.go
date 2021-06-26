package libtime

import "time"

// NullFormat - Return string or nil for sql.Nulltime
func NullFormat(input interface{}, tz string) interface{} {
	v, ok := input.(map[string]interface{})
	if ok {
		b, ok := v["Valid"].(bool)
		if ok && b {
			if s, ok := v["Time"].(string); ok {
				t, err := time.Parse(time.RFC3339, s)
				if err == nil {
					loc, _ := time.LoadLocation(tz)
					return t.In(loc).Format("2006-01-02T15:04:05-0700")
				}
			}
		}
	}
	return nil
}

// DateFormat - Return string or nil for sql.Nulltime
func DateFormat(input interface{}, tz string) interface{} {
	v, ok := input.(map[string]interface{})
	if ok {
		b, ok := v["Valid"].(bool)
		if ok && b {
			if s, ok := v["Time"].(string); ok {
				t, err := time.Parse(time.RFC3339, s)
				if err == nil {
					return t.Format("2006-01-02")
				}
			}
		}
	}
	return nil
}
