/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:30:21
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:11:00
 */

package fetcher

import (
	"strings"

	ctx "github.cedric1996.com/go-trader/app/context"
)

func ParseResponse(c *ctx.Context, input []byte) error {
	resBody := &ctx.ResponseBody{}
	arr := strings.Split(string(input), "\n")
	resBody.SetKeys(strings.Split(arr[0], ",")...)
	for _, val := range arr[1:] {
		resBody.SetVals(strings.Split(val, ","))
	}
	c.ResBody = resBody
	return nil
}
