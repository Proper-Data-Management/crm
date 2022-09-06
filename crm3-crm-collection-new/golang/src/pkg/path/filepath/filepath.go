package filepath

import (
	fp "path/filepath"

	lua "github.com/Shopify/go-lua"
)

func Split(l *lua.State) int {

	path, ok := l.ToString(1)
	if !ok {
		l.PushString("")
		l.PushString("")
		return 2
	}

	dir, file := fp.Split(path)

	l.PushString(dir)
	l.PushString(file)
	return 2
}

var lib = []lua.RegistryFunction{
	{"Split", Split},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/path/filepath", libOpen, false)
	l.Pop(1)
}
