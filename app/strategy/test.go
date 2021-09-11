/*
 * @Author: cedric.jia
 * @Date: 2021-09-05 22:14:02
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-05 22:16:46
 */

package strategy

import (
	"time"

	"github.cedric1996.com/go-trader/app/util"
)

type Test struct {
	Name   string
	Start  time.Time
	End    time.Time
	Net    float64
	Period int
}

func NewTest(name, start, end string) *Test {
	return &Test{Name: name, Start: util.ParseDate(start), End: util.ParseDate(end)}
}

func (t *Test) Run() error {
	return nil
}
