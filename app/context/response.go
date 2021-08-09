/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 12:05:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:10:58
 */

package ctx

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

func (res *ResponseBody) GetNoKeyVals() []string {
	datas := res.keys
	for _, val := range res.vals {
		datas = append(datas, val[0])
	}
	return datas
}
