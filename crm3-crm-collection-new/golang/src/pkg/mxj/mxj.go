package mxj

import (
	"encoding/json"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/util"
	lua "github.com/Shopify/go-lua"
	"github.com/clbanning/mxj"
)

func luaJsonToXml(l *lua.State) int {
	inObj, err := util.PullTable(l, 1)
	if err != nil {
		l.PushNil()
		l.PushString("JsonToXml PullTable error:" + err.Error())
		l.PushInteger(1)
		return 3
	}
	bytes, err := json.Marshal(inObj)
	if err != nil {
		l.PushNil()
		l.PushString("JsonToXml Marshal error:" + err.Error())
		l.PushInteger(1)
		return 3
	}
	m, err := mxj.NewMapJson(bytes)
	if err != nil {
		l.PushNil()
		l.PushString("JsonToXml NewMapJson error:" + err.Error())
		l.PushInteger(1)
		return 3
	}
	data, err := m.Xml()
	if err != nil {
		l.PushNil()
		l.PushString("JsonToXml mXml error:" + err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString(string(data))
	l.PushNil()
	l.PushInteger(0)
	return 3
}

func luaXmlToJson(l *lua.State) int {
	xmlValue := lua.CheckString(l, 1)
	mapValue, err := mxj.NewMapXml([]byte(xmlValue))

	// data, err := json.Marshal(mapValue)
	if err != nil {
		l.PushNil()
		l.PushString("XmlToJson NewMapXml error:" + err.Error())
		l.PushInteger(1)
		return 3
	}
	util.DeepPush(l, mapValue)
	l.PushNil()
	l.PushInteger(0)
	return 3
}

var lib = []lua.RegistryFunction{
	{"JsonToXml", luaJsonToXml},
	{"XmlToJson", luaXmlToJson},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/mxj", libOpen, false)
	l.Pop(1)
}
