package libresponse

import (
	"encoding/json"

	"github.com/helloferdie/stdgo/libtime"

	"github.com/golang-jwt/jwt"
)

// Default - Default response
type Default struct {
	Success       bool          `json:"success"`
	Code          int64         `json:"code"`
	Message       string        `json:"message"`
	MessageLocale string        `json:"message_locale"`
	MessageVar    []interface{} `json:"message_var"`
	Error         string        `json:"error"`
	ErrorLocale   string        `json:"error_locale"`
	ErrorVar      []interface{} `json:"error_var"`
	Data          interface{}   `json:"data"`
}

// Pagination - Default pagination response
type Pagination struct {
	Items      []interface{} `json:"items"`
	TotalItems int64         `json:"total_items"`
	TotalPages int64         `json:"total_pages"`
}

// GetDefault -
func GetDefault() *Default {
	res := new(Default)
	res.Success = false
	res.Code = 500
	res.Message = "general.error_general"
	return res
}

// MapOutput -
func MapOutput(obj interface{}, stdTimestamp bool, format map[string]interface{}) map[string]interface{} {
	tz, ok := format["tz"].(string)
	if !ok {
		tz = "UTC"
	}
	databytes, _ := json.Marshal(obj)
	m := map[string]interface{}{}
	json.Unmarshal(databytes, &m)
	if stdTimestamp {
		m["created_at"] = libtime.NullFormat(m["created_at"], tz)
		m["updated_at"] = libtime.NullFormat(m["updated_at"], tz)
		m["deleted_at"] = libtime.NullFormat(m["deleted_at"], tz)
	}
	return m
}

// Header -
type Header struct {
	AcceptLanguage string
	AcceptTimezone string
	Authorization  string
	Access         string
	RealIP         string
	RequestURL     string
	Claims         jwt.MapClaims
	UserID         int64
	IsLogin        bool
	IsVerified     bool
}

// FormatOutputDefault -
func FormatOutputDefault() map[string]interface{} {
	tmp := new(Header)
	return map[string]interface{}{
		"tz":     "UTC",
		"header": tmp,
	}
}
