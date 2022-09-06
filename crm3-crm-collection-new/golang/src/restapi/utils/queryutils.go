package utils

import (
	"database/sql"
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
)

func QueryByUrl(o orm.Ormer, urlStr string, host string, user_id int64, calcCount bool, lang string) (int64, []orm.Params, error) {

	url, err := url.Parse(urlStr)
	if err != nil {
		return 0, nil, err
	}
	form := url.Query()
	code := form.Get("code")
	allCount := int64(-1)

	//log.Println("urlStr=")
	//log.Println(urlStr)
	//log.Println("code=")
	//log.Println(code)

	sqlStr := ""
	entityId := int64(0)
	//o := orm.NewOrm()
	//o.Using("default")
	err = o.Raw(DbBindReplace("select sql_text,entity_id from queries where code=?"), code).QueryRow(&sqlStr, &entityId)
	if err != nil {
		return 0, nil, err
	}

	_, sqlStr, filterArray, formArray, err := QueryFilterBuild(o, entityId, sqlStr, form, user_id, lang)
	if err != nil {
		return 0, nil, err
	}

	arr := []orm.Params{}
	_, err = o.Raw(DbBindReplace(sqlStr), filterArray, formArray).Values(&arr)

	if calcCount {
		cntsqlStr := "SELECT count(1) FROM (" + sqlStr + ") alldata"
		allCount = int64(0)
		err = o.Raw(DbBindReplace(cntsqlStr), filterArray, formArray).QueryRow(&allCount)
		if err != nil {
			return 0, nil, err
		}
	}

	return allCount, arr, nil
}

func QueryGetViewQueryLimit(o orm.Ormer, entityId, userId int64) string {

	if GetRoleParamValue(o, userId, "is_admin") == "1" {
		return ""
	}

	var arr []string

	sqlStr := `
	select concat('(',evl.subquery,')') from entity_view_limits evl,entity_grants eg
	where eg.view_limit_id=evl.id
	and eg.is_list_view = 1
	and eg.role_id in (select ur.role_id from user_roles ur where user_id=?)
	and eg.entity_id=?`

	if GetDbDriverType() == orm.DROracle {
		sqlStr = `	select '('||evl.subquery||')' from entity_view_limits evl,entity_grants eg
		where eg.view_limit_id=evl.id
		and eg.is_list_view = 1
		and eg.role_id in (select ur.role_id from user_roles ur where user_id=?)
		and eg.entity_id=?`
	}
	_, err := o.Raw(DbBindReplace(sqlStr), userId, entityId).QueryRows(&arr)
	if err != nil {
		log.Println("QueryGetViewLimit " + err.Error())
		return ""
	} else if len(arr) > 0 {

		//log.Println("QueryGetViewLimit "+"and "+strings.Join(arr, " and "))
		return "and " + strings.Join(arr, " and ")
	} else {
		return " and 1 = 0"
	}
}

func UpdateLimitByEntityCode(o orm.Ormer, entityCode string, userId int64) string {

	if GetRoleParamValue(o, userId, "is_admin") == "1" {
		return ""
	}

	var arr []string

	sqlStr := `
	select concat('(',subquery,')') from entity_view_limits evl,entity_grants eg
	where eg.view_limit_id=evl.id
	and eg.is_update = 1
	and eg.role_id in (select ur.role_id from user_roles ur where user_id=?)
	and eg.entity_id=(select id from entities where code=?)`

	if GetDbDriverType() == orm.DROracle {
		sqlStr = `
		select '('||subquery||')' from entity_view_limits evl,entity_grants eg
		where eg.view_limit_id=evl.id
		and eg.is_update = 1
		and eg.role_id in (select ur.role_id from user_roles ur where user_id=?)
		and eg.entity_id=(select id from entities where code=?)`
	}
	_, err := o.Raw(DbBindReplace(sqlStr), userId, entityCode).QueryRows(&arr)
	if err != nil {
		log.Println("QueryGetViewLimit " + err.Error())
		return ""
	} else if len(arr) > 0 {

		//log.Println("QueryGetViewLimit "+"and "+strings.Join(arr, " and "))
		sqlStr := "and " + strings.Join(arr, " and ")
		sqlStr = strings.Replace(sqlStr, ":user_id", strconv.Itoa(int(userId)), -1)
		return sqlStr
	} else {
		return ""
	}
}

func QueryGetViewDetailLimitByEntityCode(entityCode string, userId int64) string {

	o := orm.NewOrm()
	o.Using("default")
	var arr []string

	sqlStr := `
	select concat('(',subquery,')') from entity_view_limits evl,entity_grants eg
	where eg.view_limit_id=evl.id
	and eg.is_detail_view = 1
	and eg.role_id in (select ur.role_id from user_roles ur where user_id=?)
	and eg.entity_id=(select id from entities where code=?)`

	if GetDbDriverType() == orm.DROracle {
		sqlStr = `
		select '('||subquery||')' from entity_view_limits evl,entity_grants eg
		where eg.view_limit_id=evl.id
		and eg.is_detail_view = 1
		and eg.role_id in (select ur.role_id from user_roles ur where user_id=?)
		and eg.entity_id=(select id from entities where code=?)`
	}

	_, err := o.Raw(DbBindReplace(sqlStr), userId, entityCode).QueryRows(&arr)
	if err != nil {
		log.Println("QueryGetViewLimit " + err.Error())
		return ""
	} else if len(arr) > 0 {

		//log.Println("QueryGetViewLimit " + "and " + strings.Join(arr, " and "))
		return "and " + strings.Join(arr, " and ")
	} else {
		return "and 1 = 0 "
	}
}

func DetailEntityGrantCheck(o orm.Ormer, entityCode string, id, userId int64) bool {

	if userId == 0 {
		return false
	}
	if GetRoleParamValue(o, userId, "is_admin") == "1" {
		return true
	}
	//log.Println("sqlStr="+sqlStr)

	vBuf := int64(0)
	sqlStr := "select 1 from " + entityCode + " main where id=? " + QueryGetViewDetailLimitByEntityCode(entityCode, userId)

	sqlStr = strings.Replace(sqlStr, ":user_id", strconv.Itoa(int(userId)), -1)

	err := o.Raw(DbBindReplace(sqlStr), id).QueryRow(&vBuf)
	if err != nil {
		log.Println("DetailGrantCheck err2 " + err.Error() + sqlStr)
	}
	return err == nil
}

func DetailGrantCheck(o orm.Ormer, detailCode string, id, userId int64) bool {

	if GetRoleParamValue(o, userId, "is_admin") == "1" || id == 0 {
		return true
	}

	entityCode := ""
	err := o.Raw(DbBindReplace("select e.code from details main,entities e where e.id = main.entity_id and main.code=?"), detailCode).QueryRow(&entityCode)
	if err != nil {
		log.Println("DetailGrantCheck err1 " + err.Error())
		return false
	}

	vBuf := int64(0)
	sqlStr := "select 1 from " + entityCode + " main where id=? " + QueryGetViewDetailLimitByEntityCode(entityCode, userId)

	sqlStr = strings.Replace(sqlStr, ":user_id", strconv.Itoa(int(userId)), -1)

	err = o.Raw(DbBindReplace(sqlStr), id).QueryRow(&vBuf)
	if err != nil {
		log.Println("DetailGrantCheck err2 " + err.Error() + sqlStr)
	}
	return err == nil
}

func QueryFilterBuild(o orm.Ormer, entityId int64, sqlStr string, form url.Values, userId int64, lang string) (bool, string, []interface{}, []interface{}, error) {

	//log.Println("sqlStr="+sqlStr)

	var filterArray []interface{}
	var formArray []interface{}
	lookupExpr := ""
	found_flt := false
	err := o.Raw(DbBindReplace("select coalesce(nullif(e.lookup_expr,''),'main.title') lookup_expr from entities e where id=?"), entityId).QueryRow(&lookupExpr)
	if err != nil {
		log.Println("Error on QueryFilterBuild", entityId, err)
		return false, "", filterArray, formArray, err
	}
	//	formArray :=	make([]interface{},0)

	sqlStr = strings.Replace(sqlStr, ":user_id", strconv.Itoa(int(userId)), -1)
	sqlStr = strings.Replace(sqlStr, ":lang", "'"+lang+"'", -1)

	if strings.Contains(sqlStr, "%filter%") {
		sqlStr_filter := " where 1 = 1 " + QueryGetViewQueryLimit(o, entityId, userId)

		//Begin #Issue #75
		if form.Get("ids") != "" && CheckListFilterSRegexpBool(form.Get("ids")) {
			sqlStr_filter = sqlStr_filter + " and main.id in (" + form.Get("ids") + ")"
		}
		//End #Issue #75

		orderBy := form.Get("orderBy")
		if orderBy == "null" || orderBy == "undefined" {
			orderBy = ""
		}
		orderAsc := form.Get("orderAsc")

		if orderAsc == "null" || orderAsc == "undefined" {
			orderAsc = ""
		}

		if orderBy != "" && CheckFieldRegexp(orderBy) != nil {
			log.Println("Error on QueryFilterBuild CheckTableRegexp orderBy")
			return false, "", filterArray, formArray, errors.New("orderBy Hacking")
		}

		if orderAsc != "" && CheckTableRegexp(orderAsc) != nil {
			log.Println("Error on QueryFilterBuild CheckTableRegexp Asc")
			return false, "", filterArray, formArray, errors.New("orderAsc Hacking")
		}

		default_order_by := ""
		default_order_asc := ""

		for formName, formValue := range form {
			if strings.HasPrefix(formName, "flt$") {
				found_flt = true
				//formArray = append(formArray,formValue[0])

				filterId := strings.Split(formName, "$")[1]
				operator := strings.Split(formName, "$")[2]

				filterIdN, _ := strconv.Atoi(filterId)

				subQuery := ""
				alias := ""
				by_child_attr := 0
				attrCode := ""
				filter := ""
				dtlsqlStr := `select
				ea.code as order_by,
				coalesce(fs.default_order_asc,0) order_asc,
				fsd.alias,
				coalesce(fsd.by_child_attr,0) by_child_attr,
				(select concat('main.',ea.code) from entities e,entity_attrs ea
				  where ea.entity_id=e.id and ea.id=fsd.entity_attr_id and e.id=fs.entity_id and coalesce(fsd.by_child_attr,0)=0) attr_code,
				fsd.subquery from 
				filter_set_dtls fsd
				join filter_sets fs on fs.id=fsd.set_id
				left join entity_attrs ea on ea.id=fs.default_order_attr_id  
				left join entity_attrs ea_fsd on ea_fsd.id=fsd.entity_attr_id
			  where
				(fsd.id=? or ea_fsd.code=?) and fs.entity_id=?
				limit 1
				`

				if GetDbDriverType() == orm.DROracle {
					dtlsqlStr = `select
					ea.code as order_by,
					coalesce(fs.default_order_asc,0) order_asc,
					fsd.alias,
					coalesce(fsd.by_child_attr,0) by_child_attr,
					(select 'main.'||ea.code from entities e,entity_attrs ea
					  where ea.entity_id=e.id and ea.id=fsd.entity_attr_id and e.id=fs.entity_id and coalesce(fsd.by_child_attr,0)=0) attr_code,
					fsd.subquery from 
					filter_set_dtls fsd
					join filter_sets fs on fs.id=fsd.set_id					
					left join entity_attrs ea on ea.id=fs.default_order_attr_id  
					left join entity_attrs ea_fsd on ea_fsd.id=fsd.entity_attr_id
				  where
					(fsd.id=? or ea_fsd.code=?) and fs.entity_id=? and rownum=1`
				}

				err := cached.O().Raw(DbBindReplace(dtlsqlStr), filterIdN, filterId, entityId).QueryRow(&default_order_by, &default_order_asc, &alias, &by_child_attr, &attrCode, &subQuery)

				if err != nil {
					log.Println("error! query " + err.Error())
				}
				if strings.Trim(subQuery, "") == "" {

					//log.Println("operator = "+operator)
					if operator == "like" {
						filter = attrCode + " like ?"
					} else if operator == "gteq" {
						filter = attrCode + " >= ?"
					} else if operator == "lteq" {
						filter = attrCode + " <= ?"
					} else if operator == "gteq" {
						filter = attrCode + " >= ?"
					} else if operator == "gt" {
						filter = attrCode + " > ?"
					} else if operator == "lt" {
						filter = attrCode + " < ?"
					} else if operator == "today_year" {
						filter = "DATE_FORMAT(" + attrCode + ",'%m%d') = DATE_FORMAT(now(),'%m%d')"
					} else if operator == "from_age" {
						filter = attrCode + " <=   DATE_SUB(now(),INTERVAL  " + formValue[0] + " YEAR) "
					} else if operator == "to_age" {
						filter = attrCode + " >=   DATE_SUB(now(),INTERVAL  " + formValue[0] + " YEAR) "
					} else if operator == "eq" {
						filter = attrCode + " = ? "
					} else if operator == "in" {

						if !CheckListFilterSRegexpBool(formValue[0]) {
							return found_flt, "", filterArray, formArray, errors.New("QueryFilterBuild sql INJECTION EVENT")
						}

						addonFilter := " 1 = 0 or "

						if strings.Contains(formValue[0], "-1") {
							//filter = "  (" + attrCode + " is null " + attrCode + " in " + formValue[0] + " )"
							addonFilter += " " + attrCode + " is null or "
						}
						if strings.Contains(formValue[0], "-3") {
							//filter = "  (" + attrCode + " = "+strconv.Itoa(int(userId))+" or " + attrCode + " in " + formValue[0] + " )"
							addonFilter += "   " + attrCode + " = " + strconv.Itoa(int(userId)) + " or "
						}
						if strings.Contains(formValue[0], "-2") {
							addonFilter += "   " + attrCode + " is not null or "
						}
						filter = " ( " + addonFilter + " " + attrCode + " in " + formValue[0] + ")"
					} else if strings.HasPrefix(operator, "func") {
						subQueryFilterFunc := ""
						is_need_data := 0
						o.Raw(DbBindReplace("select sub_query,is_need_data from filter_funcs where id=?"), strings.TrimLeft(operator, "func")).QueryRow(&subQueryFilterFunc, &is_need_data)
						if is_need_data == 1 && formValue[0] == "" {
							filter = "1 = 1"
						} else {
							if strings.Contains(subQueryFilterFunc, "{{attr}}") {
								subQueryFilterFunc = strings.Replace(subQueryFilterFunc, "{{attr}}", attrCode, -1)
								filter = subQueryFilterFunc
							} else {
								filter = attrCode + " " + subQueryFilterFunc
							}
						}
						//log.Println("filter ==== " + filter + " value ==== " + formValue[0])
					}

				} else {
					if formValue[0] == "" {
						filter = "1=1"
					} else {

						subQueryFilterFunc := ""
						o.Raw(DbBindReplace(DbBindReplace("select sub_query from filter_funcs where id=?")), strings.TrimLeft(operator, "func")).QueryRow(&subQueryFilterFunc)
						filter = subQuery

						if CheckListFilterSRegexpBool(formValue[0]) {
							filter = strings.Replace(filter, "#", formValue[0], -1)
						}
						filter = strings.Replace(filter, "{{filterfunc}}", subQueryFilterFunc, -1)
						filter = strings.Replace(filter, "{{attr}}", alias, -1)
						//log.Println("filter ==== " + filter + " value ==== " + formValue[0])
					}
				}

				if strings.Contains(filter, "?") {
					if strings.Contains(formValue[0], "|") {
						s := strings.Split(formValue[0], "|")
						//vs := s(interface{})
						filterArray = append(filterArray, s[0], s[1])
					} else {
						filterArray = append(filterArray, formValue[0])
					}
				}

				if filter != "" {
					sqlStr_filter += " and " + filter
				}

			}

		}

		if found_flt && orderBy == "" && default_order_by != "" {
			orderBy = default_order_by
			orderAsc = default_order_asc

		}

		if found_flt == false && orderBy == "" && default_order_by == "" {

			sqlStrFindDefOrder := `select concat('main.',ea.code),fs.default_order_asc from entities e
			join pages p on p.entity_id = e.id
			join filter_sets fs on fs.id =p.filter_set_id
			join entity_attrs ea on ea.id = fs.default_order_attr_id
			where e.id=?`
			if GetDbDriverType() == orm.DROracle {
				sqlStrFindDefOrder = `select 'main.'||ea.code,fs.default_order_asc from entities e
				join pages p on p.entity_id = e.id
				join filter_sets fs on fs.id =p.filter_set_id
				join entity_attrs ea on ea.id = fs.default_order_attr_id
				where e.id=?`
			}

			err = o.Raw(DbBindReplace(sqlStrFindDefOrder), entityId).QueryRow(&orderBy, &orderAsc)

		}

		if form.Get("selectContains") != "" {

			if GetDbDriverType() != orm.DROracle {
				sqlStr_filter += " and " + lookupExpr + " like concat('%',?,'%') "
			} else {
				sqlStr_filter += " and " + lookupExpr + " like '%'||?||'%'"
			}
			filterArray = append(filterArray, form.Get("selectContains"))
		}

		sqlStr = strings.Replace(sqlStr, "%filter%", sqlStr_filter, -1)

		if orderBy != "" && orderBy != "null" {

			if !strings.Contains(orderBy, ".") {

				if GetDbDriverType() == orm.DROracle {

					if orderAsc != "1" {
						orderBy = `"` + orderBy + `"` + " desc"
					} else if orderAsc == "1" {
						orderBy = `"` + orderBy + `"` + " asc"
					}

				} else {
					if orderAsc != "1" {
						orderBy = orderBy + " desc"
					} else if orderAsc == "1" {
						orderBy = orderBy + " asc"
					}

				}
			} else {
				if orderAsc != "1" {
					orderBy = orderBy + " desC"
				} else if orderAsc == "1" {
					orderBy = orderBy + " asC"
				}
			}

			sqlStr = strings.Replace(sqlStr, "%order%", "order by "+orderBy, -1)
		} else {
			sqlStr = strings.Replace(sqlStr, "%order%", "", -1)
		}

	}

	i := 0
	for formName, _ := range form {
		if strings.HasPrefix(formName, "param") {
			i++
			formArray = append(formArray, form.Get("param"+strconv.Itoa(i)))
		} else if strings.HasPrefix(formName, "named") {
			log.Println("named", formName[5:], formName, form.Get(formName))
			//formArray = append(formArray, sql.Named(formName[5:], form.Get(formName)))
			formArray = append(formArray, sql.Named("id", 1))
		}
	}

	sqlStr = strings.Replace(sqlStr, ":user_id", strconv.Itoa(int(userId)), -1)
	//sqlStr = strings.Replace(sqlStr, ":lang", GetLanguage2() ), -1)

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("sqlStr =" + sqlStr)
	}

	//log.Println(formArray)
	//log.Println(filterArray)

	return found_flt, sqlStr, filterArray, formArray, nil
}
