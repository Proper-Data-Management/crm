package utils

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func GetCustomTemplatePageByCode(pageCode string, userId int64) (string, error) {
	o := orm.NewOrm()
	o.Using("default")
	pageId := ""
	o.Raw(DbBindReplace("select id from pages where code=?"), pageCode).QueryRow(&pageId)

	return GetCustomTemplatePageById(pageId, userId)
}
func GetCustomTemplatePageById(pageId string, userId int64) (string, error) {

	o := orm.NewOrm()
	o.Using("default")
	type tStr struct {
		SubQuery string `json:"sub_query"`
		Template string `json:"string"`
	}

	var arr []tStr
	_, err := o.Raw(DbBindReplace(`
	select cp.subquery sub_query,cp.template from cus_pages cp,pages p
	where cp.page_id=p.id and p.id=?
	`), pageId).QueryRows(&arr)
	if err != nil {
		log.Println("GetTemplateByCode Error " + err.Error())
		return "", err
	} else if len(arr) > 0 {
		for _, v := range arr {
			i := 0

			v.SubQuery = strings.Replace(v.SubQuery, ":user_id", strconv.Itoa(int(userId)), -1)
			err := o.Raw(DbBindReplace("select 1 from dual where " + v.SubQuery)).QueryRow(&i)
			if err == nil {
				return v.Template, nil
			} else {
				log.Println("suk eees " + err.Error())
			}
		}
	}
	return "", errors.New("No Custom Template Found")
}
