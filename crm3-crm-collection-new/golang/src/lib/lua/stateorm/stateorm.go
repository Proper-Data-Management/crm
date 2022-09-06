package stateorm

import (
	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/statedata"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/Shopify/go-lua"
)

const orm_key = "__ORM"

func GetOrm(l *lua.State) orm.Ormer {
	return statedata.GetData(l, orm_key).(orm.Ormer)
}
func SetOrm(l *lua.State, o orm.Ormer) {
	statedata.SetData(l, orm_key, o)
}

func Open(l *lua.State, o orm.Ormer) {
	SetOrm(l, o)
}
