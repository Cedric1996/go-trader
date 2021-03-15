/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:04:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-14 21:54:06
 */

package fetcher

import "fmt"

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

	t, err := Request(body)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	token = t
	return t
}
