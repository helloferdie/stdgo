package libquery

import (
	"fmt"
	"reflect"
	"strconv"
)

// Config -
type Config struct {
	Condition   string
	Param       string
	Column      string
	ColumnValue string
}

// QueryCondition -
func QueryCondition(cfg Config, params map[string]interface{}, condition string, values map[string]interface{}) (string, map[string]interface{}, error) {
	paramVal, paramExist := params[cfg.Param]
	if paramExist {
		// Set default if not define
		if cfg.Column == "" {
			cfg.Column = "`" + cfg.Param + "`"
		}
		if cfg.ColumnValue == "" {
			cfg.ColumnValue = cfg.Param
		}

		//dt := reflect.TypeOf(paramVal).String()
		dk := reflect.TypeOf(paramVal).Kind()
		if cfg.Condition == "equal" {
			// "AND col = val "
			if dk != reflect.String || (dk == reflect.String && paramVal.(string) != "") {
				condition += fmt.Sprintf("AND %s = :%s ", cfg.Column, cfg.ColumnValue)
				values[cfg.ColumnValue] = paramVal
			}
		} else if cfg.Condition == "not_equal" {
			// "AND col != val "
			if dk != reflect.String || (dk == reflect.String && paramVal.(string) != "") {
				condition += fmt.Sprintf("AND %s != :%s ", cfg.Column, cfg.ColumnValue)
				values[cfg.ColumnValue] = paramVal
			}
		} else if cfg.Condition == "like" {
			// "AND col LIKE %val% "
			if dk != reflect.String || (dk == reflect.String && paramVal.(string) != "") {
				condition += fmt.Sprintf("AND %s LIKE :%s ", cfg.Column, cfg.ColumnValue)
				values[cfg.ColumnValue] = fmt.Sprintf("%%%v%%", paramVal)
			}
		} else if cfg.Condition == "like_match" {
			// "AND col LIKE val "
			if dk != reflect.String || (dk == reflect.String && paramVal.(string) != "") {
				condition += fmt.Sprintf("AND %s LIKE :%s ", cfg.Column, cfg.ColumnValue)
				values[cfg.ColumnValue] = paramVal
			}
		} else if cfg.Condition == "in" {
			// "AND col IN (:val1, :val2, ...)"
			if dk == reflect.Slice {
				s := reflect.ValueOf(paramVal)
				if s.Len() > 0 {
					syntax := ""
					for i := 0; i < s.Len(); i++ {
						if i != 0 {
							syntax += ", "
						}
						colVal := cfg.ColumnValue + "_in_" + strconv.Itoa(i)
						syntax += ":" + colVal
						values[colVal] = s.Index(i).Interface()
					}
					condition += fmt.Sprintf("AND %s IN (%s) ", cfg.Column, syntax)
				}
			}
		} else if cfg.Condition == "not in" {
			// "AND col NOT IN (:val1, :val2, ...)"
			if dk == reflect.Slice {
				s := reflect.ValueOf(paramVal)
				if s.Len() > 0 {
					syntax := ""
					for i := 0; i < s.Len(); i++ {
						if i != 0 {
							syntax += ", "
						}
						colVal := cfg.ColumnValue + "_in_" + strconv.Itoa(i)
						syntax += ":" + colVal
						values[colVal] = s.Index(i).Interface()
					}
					condition += fmt.Sprintf("AND %s NOT IN (%s) ", cfg.Column, syntax)
				}
			}
		}
	}
	return condition, values, nil
}
