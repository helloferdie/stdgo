package libhttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/helloferdie/stdgo/libresponse"
	"github.com/helloferdie/stdgo/logger"
)

// Request - request HTTP and expect response in JSON map[string]interface{}
func Request(addr string, method string, payloadData map[string]interface{}, headerData map[string]string) (map[string]interface{}, int, error) {
	payloadBytes, _ := json.Marshal(payloadData)
	payload := strings.NewReader(string(payloadBytes))
	req, err := http.NewRequest(method, addr, payload)
	req.Header.Add("Content-Type", "application/json")
	for k, v := range headerData {
		req.Header.Add(k, v)
	}
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return nil, 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return nil, 0, err
	}
	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return nil, res.StatusCode, err
	}

	var resJSON map[string]interface{}
	err = json.Unmarshal(resBody, &resJSON)
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return nil, res.StatusCode, err
	}
	return resJSON, res.StatusCode, nil
}

// RequestRaw - request HTTP and expect response in raw string
func RequestRaw(addr string, method string, payloadData map[string]interface{}, headerData map[string]string) (string, int, error) {
	payloadBytes, _ := json.Marshal(payloadData)
	payload := strings.NewReader(string(payloadBytes))
	req, err := http.NewRequest(method, addr, payload)
	req.Header.Add("Content-Type", "application/json")
	for k, v := range headerData {
		req.Header.Add(k, v)
	}
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return "", 0, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return "", 0, err
	}
	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.MakeLogEntry(nil, false).Error(err)
		return "", res.StatusCode, err
	}

	return string(resBody), res.StatusCode, err
}

// RequestMicroservice -
func RequestMicroservice(f map[string]interface{}, addr string, method string, payloadData map[string]interface{}) (*libresponse.Default, int, error) {
	headerData := map[string]string{
		"X-Secret": os.Getenv("microservice_secret"),
		"X-Locale": "0",
	}
	if f["header"] != nil {
		if castedObject, ok := f["header"].(*libresponse.Header); ok {
			headerData["Accept-Language"] = castedObject.AcceptLanguage
			headerData["Accept-TZ"] = castedObject.AcceptTimezone
			if castedObject.Authorization != "" {
				headerData["Authorization"] = castedObject.Authorization
			}
		}
	}

	auditTrailURL := addr
	if os.Getenv("audit_trail_gateway") == "1" {
		auditTrailURL = os.Getenv("gateway_url") + addr
	}
	res, resCode, err := Request(auditTrailURL, method, payloadData, headerData)

	def := new(libresponse.Default)
	jsonString, _ := json.Marshal(res)
	err = json.Unmarshal(jsonString, def)
	if err != nil {
		def.Code = 500
		def.Message = "general.error_internal"
		def.Error = "general.error_response_marshal_microservice"
	}
	return def, resCode, err
}

// RequestAuditTrails -
func RequestAuditTrails(payloadData map[string]interface{}) (*libresponse.Default, int, error) {
	res, resCode, err := RequestMicroservice(nil, os.Getenv("audit_trail_url_create"), "POST", payloadData)
	return res, resCode, err
}
