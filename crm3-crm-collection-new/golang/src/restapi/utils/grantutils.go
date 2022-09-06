package utils

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func CheckGrantOfEntity(o orm.Ormer, userId int64, entityCode string, grant string, pk int64) error {

	if !(grant == "is_list_view" || grant == "is_detail_view" || grant == "is_update" || grant == "is_delete" || grant == "is_insert") { //SQL Inject and Validate Grant type
		return errors.New(fmt.Sprintf("CheckGrantOfEntity. Entity %s. Unknown grant type %s", entityCode, grant))
	}

	if GetRoleParamValue(o, userId, "is_admin") == "1" {
		return nil
	}

	var queries []string
	sql1 := "select subquery from entity_grants eg,entities e,entity_view_limits evl,user_roles ur where " + grant + "=1 and e.code=? and e.id=eg.entity_id and eg.view_limit_id=evl.id and ur.role_id=eg.role_id and ur.user_id=?"
	_, err := o.Raw(DbBindReplace(sql1), entityCode, userId).QueryRows(&queries)

	if err != nil {
		log.Println("CheckGrantOfEntity err sql=" + sql1)
		return errors.New(fmt.Sprintf("CheckGrantOfEntity. Access Denied. Entity %s. Grant type %s, Error %s", entityCode, grant, err.Error()))
	}

	if len(queries) == 0 {
		log.Println("CheckGrantOfEntity err sql=" + sql1)
		return errors.New(fmt.Sprintf("CheckGrantOfEntity. Access Denied. Entity %s. Grant type %s", entityCode, grant))
	}
	log.Println(queries)
	query := strings.Join(queries, " and ")
	query = strings.Replace(query, ":user_id", strconv.Itoa(int(userId)), -1)

	cnt := 0
	sql2 := "select 1 from " + entityCode + " main where (main.id = ?) and (" + query + ")"
	err = o.Raw(DbBindReplace(sql2), pk).QueryRow(&cnt)

	if err != nil {
		log.Println("CheckGrantOfEntity err sql1=" + sql1)
		log.Println("CheckGrantOfEntity err sql2=" + sql2)
		return errors.New(fmt.Sprintf("CheckGrantOfEntity. Access Denied. Entity %s. Grant type %s, Error %s", entityCode, grant, err.Error()))
	}

	return nil
}
