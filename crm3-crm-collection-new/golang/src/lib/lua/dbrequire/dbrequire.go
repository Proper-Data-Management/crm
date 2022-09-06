package dbrequire

import (
	"git.dar.kz/crediton-3/crm-mfo/src/lib/db/luas"
	lua "github.com/Shopify/go-lua"
)

func doFunction(l *lua.State, script string) {
	if err := lua.LoadString(l, script); err != nil {
		lua.Errorf(l, "lua.LoadString error:%s", err.Error())
		panic(err)
	}
	l.Call(0, 1)
}

func doFunctionNamed(l *lua.State, script string, name string) {
	if err := lua.LoadBuffer(l, script, name, ""); err != nil {
		lua.Errorf(l, "lua.LoadString error:%s", err.Error())
		panic(err)
	}
	l.Call(0, 1)
}

func luaDbRequire(l *lua.State) int {
	code := lua.CheckString(l, 1)
	script, err := luas.GetScript(l, code)
	if err != nil {
		lua.Errorf(l, "luas.GetCachedScript(%s) error:%s", code, err.Error())
		panic(err)
	}
	doFunctionNamed(l, script, code)
	return 1
}

func Open(l *lua.State) {
	//init cached dbrequire
	lua.DoString(l, `
	__pkg={}
	function dbrequire(code)
		local r=__pkg[code]
		if r then return r end
		r=__dbrequire(code)
		__pkg[code]=r
		return r
	end`)
	l.Register("__dbrequire", luaDbRequire)
}
