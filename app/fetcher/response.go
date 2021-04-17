/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:30:21
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-17 17:32:36
 */

package fetcher

import "strings"

type dataUnit []string
type ResponseBody struct {
	keys []string
	vals []dataUnit
}

func ParseResponse(input string) *ResponseBody {
	resBody := &ResponseBody{}
	arr := strings.Split(input, "\n")
	resBody.keys = strings.Split(arr[0], ",")
	for _, val := range arr[1:] {
		resBody.vals = append(resBody.vals, strings.Split(val, ","))
	}
	return resBody
}

// func (res *ResponseBody) ToString() string {

// }
