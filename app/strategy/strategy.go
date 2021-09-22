/*
 * @Author: cedric.jia
 * @Date: 2021-08-31 11:01:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-23 15:15:10
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

type TestResult struct {
	hold     float64
	winRate  float64
	netRatio float64
	netTotal   float64
	netCount float64
	drawdown float64
	period   float64
}