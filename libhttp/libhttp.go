package libhttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

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
