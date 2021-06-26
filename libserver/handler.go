package libserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/helloferdie/stdgo/libresponse"

	"github.com/labstack/echo/v4"
)

// Ping -
func Ping(c echo.Context) (err error) {
	res := libresponse.GetDefault()
	res.Success = true
	res.Code = 200
	res.Message = "general.success"
	return Response(c, res)
}

// ErrorHandler -
func ErrorHandler(err error, c echo.Context) {
	res := libresponse.GetDefault()
	report, ok := err.(*echo.HTTPError)
	if !ok {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res.Code = int64(report.Code)
	switch res.Code {
	case 400:
		res.Message = "general.error_request"
		reportMsg := report.Message.(string)
		if reportMsg == "missing or malformed jwt" {
			res.Error = "general.error_jwt_required"
		} else {
			if e1, ok := report.Internal.(*json.UnmarshalTypeError); ok {
				fi := fmt.Sprintf("!%s", e1.Field)
				ex := fmt.Sprintf("!%s", e1.Type)
				gt := fmt.Sprintf("!%s", e1.Value)
				res.Error = "general.error_unmarshal_request_var"
				res.ErrorVar = []interface{}{fi, ex, gt}
			}
		}

		if res.Error == "" {
			res.Error = "general.error_bad_request"
		}
	case 401:
		res.Message = "general.error_request"
		reportMsg := report.Message.(string)
		if reportMsg == "invalid or expired jwt" {
			reportInternal := report.Internal.Error()
			if reportInternal == "Token is expired" {
				res.Error = "general.error_jwt_expired"
			} else {
				res.Error = "general.error_jwt_invalid"
			}
		}

		if res.Error == "" {
			res.Error = "general.error_unauthorized"
		}
	case 403:
		res.Message = "general.error_request"
		res.Error = "general.error_forbidden"
	case 404:
		res.Message = "general.error_request"
		if res.Error == "" {
			res.Error = "general.error_not_found"
		}
	case 405:
		res.Message = "general.error_request"
		res.Error = "general.error_method_not_allowed"
	case 413:
		res.Message = "general.error_request"
		res.Error = "general.error_request_too_large"
	case 415:
		res.Message = "general.error_request"
		res.Error = "general.error_unsupported_media_type"
	case 422:
		res.Message = "general.error_validation"
	case 500:
		res.Message = "general.error_internal"
		if res.Error == "" {
			res.Error = "general.error_internal"
		}
	case 502:
		res.Message = "general.error_gateway"
		res.Error = "general.error_service_unreachable"
	}
	Response(c, res)
}
