package lock

import (
	lua "github.com/Shopify/go-lua"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lock"
)

var lck lock.Lock

func luaAcquire(l *lua.State) int {
	key := lua.CheckString(l, 1)
	id, ok := lck.Acquire(key)
	l.PushInteger(int(id))
	if ok {
		l.PushInteger(0)
	} else {
		l.PushInteger(1)
	}
	return 2
}
func luaRelease(l *lua.State) int {
	key := lua.CheckString(l, 1)
	id, _ := l.ToInteger(2)
	ok := lck.Release(key, uint32(id))
	if ok {
		l.PushInteger(0)
	} else {
		l.PushInteger(1)
	}
	return 1
}

var lib = []lua.RegistryFunction{
	{"Acquire", luaAcquire},
	{"Release", luaRelease},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/lock", libOpen, false)
	l.Pop(1)
}
