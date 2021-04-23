/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:04:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-23 22:55:57
 */

package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

var token string

func Token() string {
	if len(token) > 0 {
		return token
	}

	body := map[string]interface{}{
		"method": "get_token",
		"mob":    "18851280888",
		"pwd":    "ZJjc961031",
	}

	bodyStr, err := json.Marshal(body)
	if err != nil {
		return ""
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
		return ""
	}
	token = string(res)
	return token
}
