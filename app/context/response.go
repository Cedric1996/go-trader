/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 12:05:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:10:58
 */

package context

type dataUnit []string
type ResponseBody struct {
	keys []string
	vals []dataUnit
}

func (res *ResponseBody) GetKeys() []string {
	return res.keys
}

func (res *ResponseBody) GetVals() []dataUnit {
	return res.vals
}

func (res *ResponseBody) SetKeys(keys ...string) {
	res.keys = append(res.keys, keys...)
}

func (res *ResponseBody) SetVals(vals dataUnit) {
	res.vals = append(res.vals, vals)
}
