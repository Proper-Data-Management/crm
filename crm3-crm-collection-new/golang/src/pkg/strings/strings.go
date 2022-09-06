package strings

import (
	lua "github.com/Shopify/go-lua"
)

func luaLength(l *lua.State) int {
	input := lua.CheckString(l, 1)
	l.PushInteger(len([]rune(input)))
	return 1
}
func luaSubString(l *lua.State) int {

	input := lua.CheckString(l, 1)
	start := lua.CheckInteger(l, 2)
	if start < 0 {
		start = 0
	}

	length := lua.CheckInteger(l, 3)

	asRunes := []rune(input)

	if start >= len(asRunes) {
		l.PushString("")
		return 1
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	l.PushString(string(asRunes[start : start+length]))
	return 1

}

var lib = []lua.RegistryFunction{
	{"SubString", luaSubString},
	{"Length", luaLength},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/strings", libOpen, false)
	l.Pop(1)
}
