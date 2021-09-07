/*
 * @Author: cedric.jia
 * @Date: 2021-08-31 11:01:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 22:03:31
 */

package strategy

type Strategy interface {
	Run() error
	Store() error
	Output() error
	Kelly() error
}

type TradeSignal struct {
	Code      string
	StartUnix int64
	Data      interface{}
}
