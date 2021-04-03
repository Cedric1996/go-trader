/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:29:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-03 16:59:58
 */
package handler

import "strings"

func ParseFundamentals(input string) map[string]string {
	res := make(map[string]string)
	arr := strings.Split(input, "\n")
	keys := strings.Split(arr[0], ",")
	vals := strings.Split(arr[1], ",")
	for i, key := range keys {
		if len(vals[i]) > 0 {
			res[key] = vals[i]
		}
	}
	return res
}
