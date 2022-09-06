// +build linux

package xml

import (
	"fmt"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/util"

	lua "github.com/Shopify/go-lua"
	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
)

func luaParseSchema(l *lua.State) int {
	schemaText := lua.CheckString(l, 1)
	schema, err := xsd.Parse([]byte(schemaText))
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}

	l.NewTable()
	for name, goFn := range xmlSchemaFunc {
		// -1: tbl
		l.PushGoFunction(goFn(schema))
		// -1: fn, -2:tbl
		l.SetField(-2, name)
	}
	l.PushString("")
	l.PushInteger(0)
	return 3
}

var xmlSchemaFunc = map[string]func(*xsd.Schema) lua.Function{
	"Free":     xmlSchemaFuncFree,
	"Validate": xmlSchemaFuncValidate,
}

func xmlSchemaFuncFree(schema *xsd.Schema) lua.Function {
	return func(l *lua.State) int {
		schema.Free()
		return 0
	}
}
func xmlSchemaFuncValidate(schema *xsd.Schema) lua.Function {
	return func(l *lua.State) int {
		xmlText := lua.CheckString(l, 1)
		doc, err := libxml2.ParseString(xmlText)
		ers := make([]string, 0)
		if err != nil {
			ers = append(ers, fmt.Sprintf("failed to parse XML: %s", err))
			util.DeepPush(l, ers)
			l.PushInteger(1)
			return 2
		}
		if err := schema.Validate(doc); err != nil {
			for _, e := range err.(xsd.SchemaValidationError).Errors() {
				ers = append(ers, e.Error())
			}
			util.DeepPush(l, ers)
			l.PushInteger(1)
			return 2
		}
		l.PushNil()
		l.PushInteger(0)
		return 2

	}
}

var lib = []lua.RegistryFunction{
	{"ParseSchema", luaParseSchema},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/xml", libOpen, false)
	l.Pop(1)
}
