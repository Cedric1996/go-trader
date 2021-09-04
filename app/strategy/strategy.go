/*
 * @Author: cedric.jia
 * @Date: 2021-08-31 11:01:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-04 14:35:13
 */

package strategy

type Strategy interface {
	Run() error
	Store()
}

type TradeSignal struct {
	Code      string
	StartUnix int64
}
