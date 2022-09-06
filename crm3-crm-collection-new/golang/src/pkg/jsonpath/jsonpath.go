package jsonpath

import (
	lua "github.com/Shopify/go-lua"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/util"
	jplib "github.com/yalp/jsonpath"
)

func luaRead(l *lua.State) int {
	json, err := util.PullTable(l, 1)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	path := lua.CheckString(l, 2)
	data, err := jplib.Read(json, path)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	_,ok:=data.([]interface{})
	if ok{
		util.DeepPush(l, data)
	}else{
		newSlice:=make([]interface{},1,1)
		newSlice[0]=data
		util.DeepPush(l, newSlice)
	}
	
	l.PushNil()
	l.PushInteger(0)
	return 3
}

var lib = []lua.RegistryFunction{
	{"Read", luaRead},
	//{"Prepare", luaPrepare},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/jsonpath", libOpen, false)
	l.Pop(1)
}
