/*
 * @Author: cedric.jia
 * @Date: 2021-08-31 11:01:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-05 15:10:42
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
}
