package metodosHttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func DecodeHTTPBody[T any](r *http.Request, data T) error {
	return json.NewDecoder(r.Body).Decode(data)
}

func GetHTTP[T any](ip string, port int, endpoint string) (*T, error) {
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, endpoint)

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data T
	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
	}
	return &data, nil
}

func PutHTTPwithBody[T any, R any](ip string, port int, endpoint string, data T) (*R, error) {
	var RespData *R
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, endpoint)

	body, err := json.Marshal(data)
	if err != nil {

		return RespData, err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		//fmt.Print("A")
		return RespData, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		//fmt.Print("B")
		return RespData, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		//fmt.Print("C")
		return nil, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&RespData)
	if err != nil {
		//fmt.Print("D")
		return RespData, err
	}

	return RespData, nil
}

func DeleteHTTPwithBody[T any, R any](ip string, port int, endpoint string, data T) (*R, error) {
	var RespData *R
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, endpoint)

	body, err := json.Marshal(data)
	if err != nil {
		return RespData, err
	}

	request, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(body))
	if err != nil {
		return RespData, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return RespData, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&RespData)
	if err != nil {
		return RespData, err
	}

	return RespData, nil
}

func DeleteHTTPwithQueryPath[T any, R any](ip string, port int, endpoint string, data T) (*R, error) {
	var RespData *R
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, endpoint)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return RespData, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return RespData, err
	}

	if resp.StatusCode != http.StatusOK {
		return RespData, nil
	}

	return RespData, nil
}
