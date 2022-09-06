package statedata

import (
	"github.com/Shopify/go-lua"
)

func GetData(l *lua.State, key string) interface{} {
	l.Global(key)
	data := l.ToUserData(-1)
	l.Pop(1)
	return data
}
func SetData(l *lua.State, key string, data interface{}) {
	l.PushUserData(data)
	l.SetGlobal(key)
}
