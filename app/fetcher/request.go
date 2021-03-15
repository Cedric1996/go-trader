/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:18:52
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-14 21:54:10
 */

package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	JQDATA_URL = "https://dataapi.joinquant.com/apis"
	client     = &http.Client{}
)

// Request represents a Http request to plastic web server
// type Request struct {
// 	ctx         context.Context
// 	method      string
// 	url         string
// 	headers     http.Header
// 	queryParam  url.Values
// 	data        interface{}
// 	request     *http.Request
// 	contentType string
// 	client      *http.Client
// 	buffer      *bytes.Buffer
// }

// Request create a http request
func Request(body map[string]interface{}) (string, error) {
	bodyStr, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", JQDATA_URL, strings.NewReader(string(bodyStr)))
	resp, err := client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
