package utils

import (
	"fmt"
	"log"
	"runtime"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func stack() string {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	return fmt.Sprintf("%s", buf)
}
func ErrorWrite(errorCode, details string, err error) {

	if GetParamValue("error_write_to_db") == "1" {
		o := orm.NewOrm()
		o.Using("default")

		o.Raw(DbBindReplace(`
	insert into error_logs
	(
	error_code_id,
	created_at,
	error_text,
	trace,
	details
	)
	values
	(
	(select id from error_codes where code=?),
	now(),
	?,
	?,
	?
	)
	`), errorCode, err.Error(), stack(), details).Exec()
	}
	if err != nil {
		log.Println(errorCode + " " + err.Error())
		//debug.PrintStack()
	}

}

func ErrorWriteUser(errorCode, details string, userId int64, err error) {

	if userId == 0 {
		ErrorWrite(errorCode, details, err)
		return
	}
	o := orm.NewOrm()
	o.Using("default")

	o.Raw(`
	insert into error_logs
	(
	error_code_id,
	created_at,
	error_text,
	trace,
	details,
	user_id
	)
	values
	(
	(select id from error_codes where code=?),
	now(),
	?,
	?,
	?,
	?
	)
	`, errorCode, err.Error(), stack(), details, userId).Exec()
	if err != nil {
		log.Println(errorCode + " " + err.Error())
		//debug.PrintStack()
	}

}
