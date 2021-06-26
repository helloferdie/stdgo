package libvalidator

import (
	"reflect"
	"strings"

	"github.com/helloferdie/stdgo/libresponse"

	"github.com/go-playground/validator"
)

var v10 = validator.New()

func init() {
	v10.RegisterTagNameFunc(func(fld reflect.StructField) string {
		loc := fld.Tag.Get("loc")
		if loc == "" {
			loc = "general"
		}

		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return loc + "." + name
	})
}

// VarValidationError -
type VarValidationError struct {
	Error    string        `json:"error"`
	ErrorVar []interface{} `json:"error_var"`
}

// Validate -
func Validate(i interface{}) (*libresponse.Default, error) {
	res := libresponse.GetDefault()

	err := v10.Struct(i)
	if castedObject, ok := err.(validator.ValidationErrors); ok {
		errData := map[string]*VarValidationError{}
		errVar := []interface{}{}
		errMsg := ""
		errLabel := ""
		if ok {
			res.Code = 422
			for _, err := range castedObject {
				rawField := strings.Split(err.Field(), ".")
				f := rawField[1]
				s := rawField[0]
				v := new(VarValidationError)
				switch err.Tag() {
				case "required", "required_with", "required_without":
					v.Error = "general.error_validation_required"
					v.ErrorVar = []interface{}{}
					if errMsg == "" {
						errMsg = "general.error_validation_required_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
					}
				case "numeric":
					v.Error = "general.error_validation_numeric"
					v.ErrorVar = []interface{}{}
					if errMsg == "" {
						errMsg = "general.error_validation_numeric_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
					}
				case "email":
					v.Error = "general.error_validation_email"
					v.ErrorVar = []interface{}{}
					if errMsg == "" {
						errMsg = "general.error_validation_email_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
					}
				case "e164":
					v.Error = "general.error_validation_phone"
					v.ErrorVar = []interface{}{}
					if errMsg == "" {
						errMsg = "general.error_validation_phone_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
					}
				case "min":
					tmpParam := "!" + err.Param()
					v.Error = "general.error_validation_min"
					v.ErrorVar = []interface{}{tmpParam}
					if errMsg == "" {
						errMsg = "general.error_validation_min_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
						errVar = append(errVar, tmpParam)
					}
				case "max":
					tmpParam := "!" + err.Param()
					v.Error = "general.error_validation_max"
					v.ErrorVar = []interface{}{tmpParam}
					if errMsg == "" {
						errMsg = "general.error_validation_max_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
						errVar = append(errVar, tmpParam)
					}
				case "len":
					tmpParam := "!" + err.Param()
					v.Error = "general.error_validation_len"
					v.ErrorVar = []interface{}{tmpParam}
					if errMsg == "" {
						errMsg = "general.error_validation_len_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
						errVar = append(errVar, tmpParam)
					}
				case "gte":
					tmpParam := "!" + err.Param()
					v.Error = "general.error_validation_gte"
					v.ErrorVar = []interface{}{tmpParam}
					if errMsg == "" {
						errMsg = "general.error_validation_gte_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
						errVar = append(errVar, tmpParam)
					}
				case "lte":
					tmpParam := "!" + err.Param()
					v.Error = "general.error_validation_lte"
					v.ErrorVar = []interface{}{tmpParam}
					if errMsg == "" {
						errMsg = "general.error_validation_lte_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
						errVar = append(errVar, tmpParam)
					}
				case "eqfield":
					v.Error = "general.error_validation_eqfield"
					v.ErrorVar = []interface{}{}
					if errMsg == "" {
						errMsg = "general.error_validation_eqfield_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
					}
				default:
					v.Error = "general.error_validation_default"
					v.ErrorVar = []interface{}{}
					if errMsg == "" {
						errMsg = "general.error_validation_default_var"
						errLabel = s + ".var_" + f
						errVar = append(errVar, errLabel)
					}
				}
				errData[f] = v
			}
		} else {
			res.Code = 500
			errMsg = "general.error_missing_var_validation"
		}
		res.Data = errData
		res.Error = errMsg
		res.ErrorVar = errVar
	}
	return res, err
}
