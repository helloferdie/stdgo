package service

import (
	"github.com/helloferdie/stdgo/client"
	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/libresponse"
	"github.com/helloferdie/stdgo/libvalidator"
)

// GetRequest -
type GetRequest struct {
	ClientUUID   string `json:"client_uuid" loc:"auth" validate:"required"`
	ClientSecret string `json:"client_secret" loc:"auth" validate:"required"`
}

// Get -
func Get(r *GetRequest, format map[string]interface{}) *libresponse.Default {
	res, err := libvalidator.Validate(r)
	if err != nil {
		return res
	}

	d, _ := db.Open("")
	defer d.Close()

	cl := new(client.Client)
	exist, _ := cl.GetByUUID(d, r.ClientUUID)
	if exist {
		res.Success = true
		res.Code = 200
		res.Message = "general.success_data_found"
		res.Data = map[string]interface{}{
			"id":          cl.ID,
			"client_name": cl.ClientName,
		}
	} else {
		res.Code = 401
		res.Message = "general.error_request"
		res.Error = "general.error_unauthorized_client"
	}
	return res
}
