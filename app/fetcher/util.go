/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:04:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-26 20:34:20
 */

package fetcher

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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


func PostRefDate() string {
	t:= strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	return t
}
