package utils

import (
	"fmt"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func TranslateTo(Code, Lang string) string {
	o := orm.NewOrm()
	o.Using("default")
	cnt := 0
	//Check For SQL Injection
	err := o.Raw(DbBindReplace("select 1 from langs where code = ?"), Lang).QueryRow(&cnt)
	if err != nil {
		fmt.Println("TranslateTo error 1 ", Code, " "+Lang, " ", err)
		return Code
	}
	val := ""

	delim := GetDbStringDelimiter()
	err = o.Raw(DbBindReplace("select "+delim+Lang+delim+" from translates where code=? limit 1"), Code).QueryRow(&val)
	if err != nil {
		fmt.Println("TranslateTo error 2 ", Code, " "+Lang, " ", err)
		return Code
	} else {
		return val
	}

}
