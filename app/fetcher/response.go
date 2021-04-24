/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:30:21
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:11:00
 */

package fetcher

import (
	"strings"

	"github.cedric1996.com/go-trader/app/context"
)

func ParseResponse(ctx *context.Ctx, input []byte) error {
	resBody := &context.ResponseBody{}
	arr := strings.Split(string(input), "\n")
	resBody.SetKeys(strings.Split(arr[0], ",")...)
	for _, val := range arr[1:] {
		resBody.SetVals(strings.Split(val, ","))
	}
	ctx.ResBody = resBody
	return nil
}
