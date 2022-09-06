package luas

import (
	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/stateorm"
	//"git.dar.kz/crediton-3/src-mfo/src/restapi/ext/cached"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	lua "github.com/Shopify/go-lua"
)

func GetScript(l *lua.State, code string) (string, error) {
	var script string
	//o:=cached.O()
	o := stateorm.GetOrm(l) //todo:заменить на кэш, после добавления триггера на luas
	err := o.Raw(utils.DbBindReplace("select script from luas where code=?"), code).QueryRow(&script)
	return script, err
}
