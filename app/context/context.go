/*
 * @Author: cedric.jia
 * @Date: 2021-04-23 22:38:50
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:05:02
 */

package ctx

import (
	gocontext "context"
)

type Context struct {
	gocontext.Context
	methodName string
	requestKey string

	ResBody *ResponseBody
	Params  map[string]interface{}
}
