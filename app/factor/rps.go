/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-05 14:10:50
 */

package factor

type rpsFactor struct {
	name    string
	period  int
	percent int
	calDate string
}

func NewRpcFactor(name string, period int, percent int, calDate string) *rpsFactor {
	return &rpsFactor{name: name, period: period, percent: percent, calDate: calDate}
}
func (f *rpsFactor) Get() error {
	// datas, err := models.GetPriceList()
	return nil
}

func (f *rpsFactor) Run() error {
	return nil
}
