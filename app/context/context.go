/*
 * @Author: cedric.jia
 * @Date: 2021-04-23 22:38:50
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-23 22:59:15
 */

package context

import (
	gocontext "context"
)

type Ctx struct {
	gocontext.Context
	methodName string
	requestKey string
	Params     map[string]interface{}
}
