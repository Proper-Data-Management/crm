package utils

import (
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"

	"log"

	"errors"
	"regexp"
	"strings"
)

//var openshift_db = os.Getenv("OPENSHIFT_APP_NAME")

func CheckListIDSRegexpBool(list string) bool {
	var validID = regexp.MustCompile(`^[0-9\,\)\(]+$`)
	res := validID.MatchString(list)
	validID = nil
	return res

}

func CheckListFilterSRegexpBool(list string) bool {
	list = SQLInjectTruncate(list)
	var validID = regexp.MustCompile(`^[0-9\,\)\(-]+$`)
	//return validID.MatchString(list)
	res := validID.MatchString(list)
	validID = nil
	return res

}

func SQLInjectTruncate(sql string) string {
	sql = strings.Replace(sql, "\"", "", -1)
	sql = strings.Replace(sql, "'", "", -1)
	sql = strings.Replace(sql, "/*", "", -1)
	sql = strings.Replace(sql, "--", "", -1)
	sql = strings.Replace(sql, "#", "", -1)
	return sql

}

func CheckTableRegexpBool(entityCode string) bool {
	var validID = regexp.MustCompile(`^[a-z|0-9|\_]+$`)
	//return validID.MatchString(entityCode)
	res := validID.MatchString(entityCode)
	validID = nil
	return res

}

func CheckTableRegexp(fieldName string) error {
	var validID = regexp.MustCompile(`^[a-z|0-9|\_]+$`)

	var err error = nil
	if !validID.MatchString(fieldName) {
		validID = nil
		err = errors.New("invalid field name " + fieldName)
	} else {
		validID = nil
		err = nil
	}
	return err

}

func CheckFieldRegexp(fieldName string) error {
	var validID = regexp.MustCompile(`^[a-z|0-9|\_|\$]+$`)

	var err error = nil
	if !validID.MatchString(fieldName) {
		err = errors.New("invalid field name {" + fieldName + "}")
	} else {
		err = nil
	}
	validID = nil
	return err

}

func CheckEntityFullAudit(o orm.Ormer, entityCode interface{}) bool {

	ok := 0
	err := o.Raw(DbBindReplace("select coalesce(is_full_audit,0) from entities where code=?"), entityCode).QueryRow(&ok)
	if err != nil {
		log.Println("error CheckEntityFullAudit" + err.Error())
	}
	return err == nil && ok == 1
}

func GetEntityId(entityCode interface{}) int64 {

	o := orm.NewOrm()
	o.Using("default")
	res := int64(0)
	err := o.Raw(DbBindReplace("select id from entities where code=?"), entityCode).QueryRow(&res)
	if err != nil {
		return 0
	}
	return res
}

func CheckEntity(entityCode interface{}) bool {

	o := orm.NewOrm()
	o.Using("default")
	ok := 0
	err := o.Raw(DbBindReplace("select 1 from entities where code=?"), entityCode).QueryRow(&ok)
	return err == nil && ok == 1

	/*	func main() {
		var validID = regexp.MustCompile(`^[a-z|0-9]+$`)
		fmt.Println(validID.MatchString("1sal/em"))
	}*/
}

func tableExists(entityCode string) bool {

	o := orm.NewOrm()
	o.Using("default")
	i := 0
	err := o.Raw(DbBindReplace("SELECT 1 FROM information_schema.tables	WHERE table_schema = ? AND table_name = ? limit 1"), openshift_db, entityCode).QueryRow(&i)
	return err == nil
}
