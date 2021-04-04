/*
 * @Author: cedric.jia
 * @Date: 2021-04-04 22:13:07
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-04 22:32:59
 */
package redis

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestRedis(t *testing.T) {
	err := Client().SetString("test", "new test")
	assert.NoError(t, err)
	val, err := Client().GetString("test")
	assert.NoError(t, err)
	assert.Equal(t, "new test", val)
	del := Client().Delete("test")
	assert.Equal(t, del, true)
	assert.Equal(t, Client().Exist("test"), false)
}
