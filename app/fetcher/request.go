/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:18:52
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-23 22:59:31
 */

package fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/database"
)

var (
	JQDATA_URL = "https://dataapi.joinquant.com/apis"
	client     = &http.Client{}
)

// Request create a http request
func Request(c *ctx.Context) ([]byte, error) {
	isRequested, err := checkRequest(c.Params)
	if err != nil {
		return nil, err
	} else if isRequested {
		return nil, nil
	}

	bodyStr, err := json.Marshal(c.Params)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return res, nil
}

func paramsEncoder(params map[string]interface{}) string {
	res := ""
	if val, ok := params["method"].(string); ok {
		res += fmt.Sprintf("method=%s;", val)
	}
	if val, ok := params["code"].(string); ok {
		res += fmt.Sprintf("code=%s;", val)
	}
	if val, ok := params["date"].(string); ok {
		res += fmt.Sprintf("date=%s;", val)
	}
	if val, ok := params["end_date"].(string); ok {
		res += fmt.Sprintf("end_date=%s;", val)
	}
	if val, ok := params["table"].(FinTable); ok {
		res += fmt.Sprintf("table=%s;", val)
	}
	if val, ok := params["unit"].(TimeScope); ok {
		res += fmt.Sprintf("unit=%s;", val)
	}
	if val, ok := params["count"].(int64); ok {
		res += fmt.Sprintf("count=%s;", strconv.FormatInt(val, 10))
	}
	return res
}

func checkRequest(body map[string]interface{}) (bool, error) {
	requestKey := paramsEncoder(body)
	success, err := database.IsFetchSuccess(requestKey)
	if err != nil {
		return false, err
	}
	return success, nil
}
