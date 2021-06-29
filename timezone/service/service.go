package service

import (
	"math"

	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/libresponse"
	"github.com/helloferdie/stdgo/libslice"
	"github.com/helloferdie/stdgo/libvalidator"
	"github.com/helloferdie/stdgo/timezone"
)

// FormatOutput -
func FormatOutput(obj *timezone.Timezone, format map[string]interface{}) map[string]interface{} {
	m := libresponse.MapOutput(obj, true, format)
	return m
}

// ListRequest -
type ListRequest struct {
	Page             int64  `json:"page" loc:"general" validate:"required,numeric,min=1"`
	ItemsPerPage     int64  `json:"items_per_page" loc:"general" validate:"required,numeric,min=1,max=500"`
	OrderByField     string `json:"order_by_field" loc:"general"`
	OrderByDir       string `json:"order_by_direction" loc:"general"`
	ShowRelationship bool   `json:"show_relationship" loc:"general"`
	ID               string `json:"id" loc:"general" validate:"omitempty,numeric"`
	Label            string `json:"label" loc:"timezone"`
	UTFOffset        string `json:"utc_offset" loc:"timezone"`
}

// List -
func List(r *ListRequest, format map[string]interface{}) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	params := map[string]interface{}{
		"id":         r.ID,
		"label":      r.Label,
		"utc_offset": r.UTFOffset,
	}
	orderParams := map[string]interface{}{
		"field":     r.OrderByField,
		"direction": r.OrderByDir,
		"start":     ((r.Page - 1) * r.ItemsPerPage),
		"limit":     r.ItemsPerPage,
	}

	d, _ := db.Open("")
	defer d.Close()

	list, totalItems, err := timezone.List(d, params, orderParams)
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

// ViewRequest -
type ViewRequest struct {
	ID int64 `json:"id" loc:"general" validate:"required,numeric"`
}

// View -
func View(r *ViewRequest, format map[string]interface{}) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	d, _ := db.Open("")
	defer d.Close()
	tz := new(timezone.Timezone)
	exist, err := tz.GetByID(d, r.ID)
	if err == nil && exist {
		res.Success = true
		res.Code = 200
		res.Message = "general.success_data_found"
		res.Data = FormatOutput(tz, format)
	} else {
		res.Code = 404
		res.Error = "general.error_data_not_found"
	}

	return res
}

// CheckRequest -
type CheckRequest struct {
	ID []int64 `json:"id" loc:"general" validate:"required,min=1"`
}

// Check -
func Check(r *CheckRequest, format map[string]interface{}) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	d, _ := db.Open("")
	defer d.Close()

	valid, err := timezone.MassCheckID(d, libslice.UniqueInt64(r.ID))
	if err == nil && valid {
		res.Success = true
		res.Code = 200
		res.Message = "general.success_check"
	} else {
		res.Success = true
		res.Code = 422
		res.Message = "general.error_validation"
		res.Error = "general.error_check"
	}
	return res
}
