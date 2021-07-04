package service

import (
	"math"
	"strings"

	"github.com/helloferdie/stdgo/audittrail"
	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/libencryption"
	"github.com/helloferdie/stdgo/libresponse"
	"github.com/helloferdie/stdgo/libslice"
	"github.com/helloferdie/stdgo/libtime"
	"github.com/helloferdie/stdgo/libvalidator"
)

// FormatOutput -
func FormatOutput(obj *audittrail.AuditTrail, format map[string]interface{}) map[string]interface{} {
	m := libresponse.MapOutput(obj, false, format)
	m["created_at"] = libtime.NullFormat(m["created_at"], format["tz"].(string))
	return m
}

// ListRequest -
type ListRequest struct {
	Page             int64  `json:"page" loc:"general" validate:"required,numeric,min=1"`
	ItemsPerPage     int64  `json:"items_per_page" loc:"general" validate:"required,numeric,min=1,max=500"`
	OrderByField     string `json:"order_by_field" loc:"general"`
	OrderByDir       string `json:"order_by_direction" loc:"general"`
	ShowRelationship bool   `json:"show_relationship" loc:"general"`
	ID               string `json:"id" loc:"general"`
	Operation        string `json:"operation" loc:"audit"`
	ModuleName       string `json:"module_name" loc:"audit"`
	TableName        string `json:"table_name" loc:"audit"`
	TablePK          string `json:"table_pk" loc:"audit"`
	Change           string `json:"change" loc:"audit"`
	Remark           string `json:"remark" loc:"audit"`
	CreatedBy        string `json:"created_by" loc:"audit"`
}

// List -
func List(r *ListRequest, format map[string]interface{}) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	d, _ := db.Open("")
	defer d.Close()

	params := map[string]interface{}{
		"id":          r.ID,
		"operation":   r.Operation,
		"module_name": r.ModuleName,
		"table_name":  r.TableName,
		"table_pk":    r.TablePK,
		"change":      r.Change,
		"remark":      r.Remark,
		"created_by":  r.CreatedBy,
	}
	orderParams := map[string]interface{}{
		"field":     r.OrderByField,
		"direction": r.OrderByDir,
		"start":     ((r.Page - 1) * r.ItemsPerPage),
		"limit":     r.ItemsPerPage,
	}

	list, totalItems, err := audittrail.List(d, params, orderParams)
	if err != nil {
		res.Code = 500
		res.Message = "general.error_internal"
		res.Error = "general.error_list"
	} else {
		tmp := make([]interface{}, len(list))
		format["show_relationship"] = r.ShowRelationship
		for k, obj := range list {
			tmp[k] = FormatOutput(&obj, format)
		}
		res.Success = true
		res.Code = 200
		res.Message = "general.success_list"
		totalPages := math.Ceil(float64(totalItems) / float64(r.ItemsPerPage))
		res.Data = map[string]interface{}{
			"items":       tmp,
			"total_items": totalItems,
			"total_pages": totalPages,
		}
	}
	return res
}

// CreateRequest -
type CreateRequest struct {
	Operation  string `json:"operation" loc:"audit"`
	ModuleName string `json:"module_name" loc:"audit"`
	TableName  string `json:"table_name" loc:"audit"`
	TablePK    string `json:"table_pk" loc:"audit"`
	Change     string `json:"change" loc:"audit"`
	Remark     string `json:"remark" loc:"audit"`
	CreatedBy  int64  `json:"created_by" loc:"audit"`
}

// Create -
func Create(r *CreateRequest, format map[string]interface{}) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	d, _ := db.Open("")
	defer d.Close()

	allowOperation := []string{"add", "edit", "delete", "view"}
	_, ok := libslice.Contains(r.Operation, allowOperation)
	if !ok {
		allowStr := "!" + strings.Join(allowOperation[:], ", ")
		res.Code = 422
		res.Message = "general.error_validation"
		res.Error = "general.error_validation_option_var"
		res.ErrorVar = []interface{}{"audit.var_operation", allowStr}
		res.Data = map[string]interface{}{
			"operation": map[string]interface{}{
				"error":     "general.error_validation_option",
				"error_var": []interface{}{allowStr},
			},
		}
		return res
	}

	at := new(audittrail.AuditTrail)
	at.Operation = r.Operation
	at.ModuleName = r.ModuleName
	at.TableName = r.TableName
	at.TablePK = r.TablePK
	at.Change = r.Change
	at.Remark = r.Remark
	at.CreatedBy = r.CreatedBy
	id, err := at.Create(d)
	if err != nil {
		res.Code = 500
		res.Message = "general.error_internal"
		res.Error = "general.error_create"
	} else {
		res.Success = true
		res.Code = 200
		res.Message = "general.success_create"
		res.Data = map[string]interface{}{
			"id": id,
		}
	}
	return res
}

// EncryptRequest -
type EncryptRequest struct {
	Data string `json:"data"`
}

// Encrypt -
func Encrypt(r *EncryptRequest) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	v, err := libencryption.Encrypt(r.Data)
	if err != nil {
		res.Code = 500
		res.Message = "general.error_internal"
		res.Error = "general.error_list"
		res.Data = err.Error()
	} else {
		res.Success = true
		res.Code = 200
		res.Message = "general.success_list"
		res.Data = v
	}
	return res
}

// DecryptRequest -
type DecryptRequest struct {
	Data string `json:"data"`
}

// Decrypt -
func Decrypt(r *DecryptRequest) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	v, err := libencryption.Decrypt(r.Data)
	if err != nil {
		res.Code = 500
		res.Message = "general.error_internal"
		res.Error = "general.error_list"
		res.Data = err.Error()
	} else {
		res.Success = true
		res.Code = 200
		res.Message = "general.success_list"
		res.Data = v
	}
	return res
}
