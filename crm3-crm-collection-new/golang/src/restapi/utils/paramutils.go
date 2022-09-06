package utils

import (
	"log"
	"os"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func GetDomainParamValue(domain string, param string) string {
	o := orm.NewOrm()
	o.Using("default")
	val := ""
	//log.Println("domain="+domain)
	//log.Println("param="+param)
	sql := "select dpv.value from domain_param_values dpv,params p,domains d where d.id=dpv.domain_id and p.id=dpv.param_id and d.domain=? and p.code=?"
	if os.Getenv("CRM_DB_TYPE") == "oracle" {
		sql = "select dpv.value from domain_param_values dpv,params p,domains d where d.id=dpv.domain_id and p.id=dpv.param_id and d.domain=:1 and p.code=:2"
	}
	err := o.Raw(DbBindReplace(sql), domain, param).QueryRow(&val)
	if err != nil {
		if IsNoRowFound(err) {
			return GetParamValue(param)
		} else {
			log.Println("error on GetDomainParamValue " + err.Error())
		}
	}
	return val
}

func GetRoleParamValue(o orm.Ormer, userId int64, param string) string {
	val := ""
	err := o.Raw(DbBindReplace(`select rp.value from role_params rp,params p,user_roles ur
	WHERE
	  rp.role_id=ur.role_id and p.id=rp.param_id and ur.user_id=?
	  and  p.code=? limit 1`), userId, param).QueryRow(&val)
	if err != nil {
		return ""
	} else {
		return val
	}
}

func GetParamValue(param string) string {
	o := orm.NewOrm()
	o.Using("default")
	val := ""
	sql := "select p.value from params p where p.code=?"
	if os.Getenv("CRM_DB_TYPE") == "oracle" {
		sql = "select p.value from params p where p.code=:1"
	}
	err := o.Raw(DbBindReplace(sql), param).QueryRow(&val)
	if err != nil {
		log.Println("GetParamValue Error on get Param", param)
	}
	return val
}

func GetUserParamValue(o orm.Ormer, user int64, param string) string {

	val := GetRoleParamValue(o, user, param)
	log.Println("GetRoleParamValue=" + val)
	if val == "" {
		err := o.Raw(DbBindReplace("select up.value from user_params up,params p where p.id=up.param_id and up.user_id=? and p.code=?"), user, param).QueryRow(&val)
		if err != nil {
			log.Println("GetUserParamValue err " + err.Error())
			return GetParamValue(param)
		}
	}
	log.Println("userParamValue" + val + ",param=" + param)
	log.Println(user)
	return val
}
