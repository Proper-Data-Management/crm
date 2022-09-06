package luautils

import (
	"log"
	"runtime/debug"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/Shopify/go-lua"
)

func AutoLoads() error {

	type TRows struct {
		Title  string
		Script string
	}
	log.Println("loading1... AutoLoads")
	var rows []string
	o := orm.NewOrm()
	o.Using("default")
	_, err := o.Raw("select script from lua_autoloads where is_active = 1 order by nn").QueryRows(&rows)
	if err != nil {
		log.Println("error... AutoLoads", err)
		return err
	}
	log.Println("loading2... AutoLoads")
	l := lua.NewState()
	lua.OpenLibraries(l)

	o.Begin()
	defer o.Rollback()
	RegisterAPI(l, o)
	RegisterBPMLUaAPI(nil, l, o)

	Open(l)
	luaContext := LuaContext{}
	luaContext.o = o

	for _, v := range rows {

		if err := lua.DoString(l, v); err != nil {

			log.Println("AutoLoads error lua  " + err.Error())
			debug.PrintStack()
		}

	}
	o.Commit()
	log.Println("end 3... AutoLoads")
	return nil

}
