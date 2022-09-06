package ast

import (
	lua2 "git.dar.kz/crediton-3/crm-mfo/src/lib/go-lua-parser"
	util "git.dar.kz/crediton-3/crm-mfo/src/lib/lua/util"
	lua "github.com/Shopify/go-lua"
)

func luaParse(l *lua.State) int {
	code := lua.CheckString(l, 1)
	l2 := lua2.NewState()
	err := lua2.LoadString(l2, code)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	n := l2.GetParser().AstTop()
	util.DeepPush(l, n)
	l.PushNil()
	l.PushInteger(0)
	return 3
}

var lib = []lua.RegistryFunction{
	{"parse", luaParse},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/ast", libOpen, false)
	l.Pop(1)
}
