package luautils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"errors"
	"reflect"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	lua "github.com/Shopify/go-lua"
)

type DMLRequestItem struct {
	Values      []orm.Params `json:"values"`
	lastData    []orm.Params
	TableName   string   `json:"table_name"`
	UpsertAttrs []string `json:"upsert_attrs"`
	Action      string   `json:"action"`
	origAction  string   `json:"orig_action"`
}

type DMLRequest struct {
	Items []DMLRequestItem `json:"items"`
}

type UpdateResponse_Item struct {
	TableName    string `json:"table_name"`
	Action       string `json:"action"`
	lastData     []orm.Params
	LastInsertId int64  `json:"last_insert_id"`
	SysUUID      string `json:"sys$uuid"`
}

type TRqErrorAttrs struct {
	TableName string `json:"table_name"`
	Title     string `json:"title"`
	Code      string `json:"code"`
}

type UpdateResponse struct {
	Items        []UpdateResponse_Item `json:"items"`
	Error        int64                 `json:"error"`
	ErrorText    string                `json:"error_text"`
	ErrorCode    string                `json:"error_code"`
	ErrorRecom   string                `json:"error_recom"`
	ErrorDscr    string                `json:"error_dscr"`
	RqErrorAttrs []TRqErrorAttrs       `json:"rq_error_attrs"`
}

func setDefaultUserId(o orm.Ormer, tableName string, userId int64, lastInsertId int64, values orm.Params) {
	//log.Println("setDefaultUserId start")
	var sqls []string
	_, err := cached.O().Raw(utils.DbBindReplace(`select ea.code from entity_attrs ea,entities e where
					e.code=? and ea.entity_id=e.id and ea.is_default_current_user = 1`), tableName).QueryRows(&sqls)

	if err != nil {
		log.Println("setDefaultUserId err1 " + err.Error())
	}
	for _, sql := range sqls {

		//Don't Worry tableName has been checked
		o.Raw("update "+tableName+" set "+sql+" = ? where id= ?", userId, lastInsertId).Exec()

		if err != nil {
			log.Println("setDefaultUserId err2 " + err.Error())
		}

	}

}

func setOnUpdateByUserVar(o orm.Ormer, tableName string, userId int64, Id int) error {
	//log.Println("setDefaultUserId start")
	var sqls []string

	sqlGet := `select concat('update ',e.code ,'  set ',ea.code,' = (select main.',uea.code,' from users main where main.id=XXX) where id=XXX ') from
	entity_attrs ea,
	entities e,
	user_vars uv,
	entity_attrs uea where
   e.code=? and ea.entity_id=e.id
  and ea.on_update_user_var_id = uv.id and uea.id=uv.user_entity_attr_id
  `
	if utils.GetDbDriverType() == orm.DROracle {
		sqlGet = `select 'update '||e.code ||'  set '||ea.code||' = (select main.'||uea.code||' from users main where main.id=XXX) where id=XXX ' from
	entity_attrs ea,
	entities e,
	user_vars uv,
	entity_attrs uea where
   e.code=? and ea.entity_id=e.id
  and ea.on_update_user_var_id = uv.id and uea.id=uv.user_entity_attr_id
  `

	}
	_, err := cached.O().Raw(utils.DbBindReplace(sqlGet), tableName).QueryRows(&sqls)

	if err != nil {
		log.Println("setOnUpdateByUserVar err222 ", tableName, err.Error())
		return err
	}

	for _, sql := range sqls {

		sql = strings.Replace(sql, "XXX", "?", -1)

		//Don't Worry tableName has been checked

		_, err = o.Raw(utils.DbBindReplace(sql), userId, Id).Exec()

		if err != nil {
			log.Println(sql)
			log.Println("setOnUpdateByUserVar err2 " + err.Error())
			return err
		}

	}

	return nil

}

func setDefaultByUserVar(o orm.Ormer, tableName string, userId int64, lastInsertId int64) {
	//log.Println("setDefaultUserId start")
	var sqls []string

	sqlSet := `select concat('update ',e.code ,'  set ',ea.code,' = (select main.',uea.code,' from users main where main.id=XXX) where id=XXX') from
	entity_attrs ea,
	entities e,
	user_vars uv,
	entity_attrs uea where
   e.code=? and ea.entity_id=e.id
  and ea.default_user_var_id = uv.id and uea.id=uv.user_entity_attr_id
  `

	if utils.GetDbDriverType() == orm.DROracle {

		sqlSet = `select 'update '||e.code||'  set '||ea.code||' = (select main.'||uea.code||' from users main where main.id=XXX) where id=XXX' from
	entity_attrs ea,
	entities e,
	user_vars uv,
	entity_attrs uea where
   e.code=? and ea.entity_id=e.id
  and ea.default_user_var_id = uv.id and uea.id=uv.user_entity_attr_id
  `

	}
	_, err := cached.O().Raw(utils.DbBindReplace(sqlSet), tableName).QueryRows(&sqls)

	if err != nil {
		log.Println("setDefaultByUserVar err1 " + err.Error())
	}
	for _, sql := range sqls {

		//Don't Worry tableName has been checked
		sql = strings.Replace(sql, "XXX", "?", -1)
		o.Raw(utils.DbBindReplace(sql), userId, lastInsertId).Exec()
		//log.Println("s!!!!")

		if err != nil {
			log.Println("setDefaultByUserVar err2 " + err.Error())
		}

	}

}

/*
func doInsertEvent(o orm.Ormer, tableName string,userId int64, lastInsertId int64,values orm.Params){

	//log.Println("doInsertEvent start")
	var sqls []string
	_,err := o.Raw(`select eva.data from event_actions eva,events ev,entities e,event_action_types evat
where e.id=ev.entity_id and eva.event_id=ev.id and evat.id=eva.type_id
and e.code=? and evat.code='sqlexec'
and ev.on_insert = '1' order by eva.nn`,tableName).QueryRows(&sqls)

	if err != nil {
		log.Println("doInsertEvent err1 " + err.Error())
	}
	for _,sql := range sqls {
		for fieldName,fieldValue := range values {
			//i,err := strconv.Atoi(fieldValue.(string))
			//			if err!=nil{
			//				log.Println("skipping non integer value fieldName="+fieldValue.(string))
			//			}

			//else {
			sql = strings.Replace(sql, "{{insert." + fieldName+"}}", "'"+fmt.Sprintf("%v",fieldValue)+"'", -1)
			//}
		}
		sql = strings.Replace(sql, "{{user_id}}", strconv.Itoa(int(userId)), -1)
		sql = strings.Replace(sql, "{{last_insert_id}}", strconv.Itoa(int(lastInsertId)), -1)
		_, err = o.Raw(utils.DbBindReplace(sql)).Exec()
		if err != nil {
			log.Println("doInsertEvent err2 " + err.Error())
		}
	}

	sqls = sqls [:0]

}
*/
//Замена uuid на ID.
func replaceUUIDtoID(o orm.Ormer, tableName string, values orm.Params) (orm.Params, error) {

	//	o := orm.NewOrm()
	//	o.Using("default")

	delim := utils.GetDbStringDelimiter()

	for fieldName, fieldValue := range values {
		if fieldValue != nil && !strings.HasPrefix(fieldName, "_") && fieldName != "sys$uuid" {
			//log.Println("tableName==="+tableName)
			//log.Println("FieldName==="+fieldName)
			//log.Println("fieldValue===")
			//log.Println(fieldValue)

			if reflect.TypeOf(fieldValue).String() == "string" && utils.IsValidUUID(fieldValue.(string)) {
				linkTable := ""
				err := cached.O().Raw(utils.DbBindReplace("select code from entities where id=(select ea.entity_link_id from entity_attrs ea,entities ee where ea.code=? and ea.entity_id=ee.id and ee.code=?)"), fieldName, tableName).QueryRow(&linkTable)
				if err != nil {
					log.Println("replaceUUIDtoID not found " + err.Error() + "fieldName, tableName =>" + fieldName + ", " + tableName)
					//return values,nil
				} else {
					id := 0
					sql := "select id from " + delim + linkTable + delim + " where sys$uuid=?"
					err = o.Raw(utils.DbBindReplace(sql), fieldValue).QueryRow(&id)
					values[fieldName] = id
					log.Println("replace = " + fieldValue.(string) + " " + sql)
					log.Println(id)
					//log.Println(err)
					//return values,nil
				}
			} else {
				//log.Println("not valid "+tableName+" "+fieldName+" "+reflect.TypeOf(fieldValue).String())
			}
		}
	}
	return values, nil
}

type TDataTypes struct {
	Code         string `json:"code"`
	DataTypeCode string `json:"data_type_code"`
}

func getDataTypeByAttrCode(code string, dataTypes []TDataTypes) string {
	//log.Println("getDataTypeByAttrCode",dataTypes)
	for _, v := range dataTypes {
		if v.Code == code {
			return v.DataTypeCode
		}
	}
	return ""
}

func TriggerSync(o orm.Ormer, user_id int64, t DMLRequest, resP UpdateResponse) (UpdateResponse, error) {

	//BEGIN EVENT PROCESSING
	for _, valueEvent := range resP.Items {

		var ps []int64
		_, err := cached.O().Raw(utils.DbBindReplace("select p.id from bp_processes p, entities e,entity_event_types eet where p.action_entity_id=e.id and e.code=? and eet.id=p.entity_event_type_id and p.is_active=1 and (eet.code='after_upsert' and ? in ('insert','update') or eet.code=?) "), valueEvent.TableName, valueEvent.Action, "after_"+valueEvent.Action).QueryRows(&ps)
		if err != nil {
			log.Println("ERROR  UPDATE RESTAPI 1 1 " + err.Error())
			log.Println("details:")
			log.Println("action", valueEvent.Action)
			return resP, err
		}
		if len(ps) > 0 {
			for _, v := range ps {
				if len(valueEvent.lastData) > 0 {
					eventBPParams := []NameValue{NameValue{Name: "pk", Value: strconv.Itoa(int(valueEvent.LastInsertId))}, NameValue{Name: "old", Value: valueEvent.lastData[0]}}

					context := InstanceContext{}
					context.InstanceVars = make(map[int64][]NameValue)
					context.InstanceTables = make(map[int64]string)
					context.O = o
					context.Lua = lua.NewState()
					lua.OpenLibraries(context.Lua)
					RegisterAPI(context.Lua, o)
					RegisterBPMLUaAPI(nil, context.Lua, o)

					_, _, _, _, err := context.CreateInstance(nil, v, user_id, eventBPParams, 0)
					if err != nil {

						resP.ErrorText = err.Error()
						//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 24 " + err.Error() + err2.Error()
						//	return resP, err2
						//}

						log.Println("ERROR UPDATE RESTAPI EVENT RUN BP " + err.Error())
						log.Println("details:")
						log.Println(valueEvent)
						return resP, err
					}
				} else {
					eventBPParams := []NameValue{NameValue{Name: "pk", Value: strconv.Itoa(int(valueEvent.LastInsertId))}, NameValue{Name: "old", Value: "{}"}}
					context := InstanceContext{}
					context.InstanceVars = make(map[int64][]NameValue)
					context.InstanceTables = make(map[int64]string)
					context.O = o
					context.Lua = lua.NewState()
					lua.OpenLibraries(context.Lua)
					RegisterAPI(context.Lua, o)
					RegisterBPMLUaAPI(nil, context.Lua, o)
					_, _, _, _, err := context.CreateInstance(nil, v, user_id, eventBPParams, 0)
					if err != nil {
						resP.ErrorText = err.Error()
						//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
						//if err2 != nil {
						//	resP.ErrorText = err.Error()
						//	log.Println(err2.Error())
						//
						//	return resP, err2
						//}

						log.Println("ERROR UPDATE RESTAPI EVENT RUN BP " + err.Error())
						log.Println("details:")
						log.Println(valueEvent)
						return resP, err
					}
				}

			}
		}

	}
	return resP, nil
}

func TriggerSyncBefore(o orm.Ormer, user_id int64, t DMLRequest) (UpdateResponse, error) {

	//BEGIN EVENT PROCESSING
	resP := UpdateResponse{}
	//log.Println("get data t items", t)
	for _, valueEvent := range t.Items {

		//log.Println("get data 2", valueEvent)

		for _, valueEvent2 := range valueEvent.Values {

			var ps []int64
			_, err := cached.O().Raw(utils.DbBindReplace("select p.id from bp_processes p, entities e,entity_event_types eet where p.action_entity_id=e.id and e.code=? and eet.id=p.entity_event_type_id and p.is_active=1 and (eet.code='before_upsert' and ? in ('insert','update') or eet.code=?) "), valueEvent.TableName, valueEvent.Action, "before_"+valueEvent.Action).QueryRows(&ps)
			//log.Println("get data ", ps, valueEvent.TableName, valueEvent.Action, "before_"+valueEvent.Action)

			if err != nil {
				log.Println("ERROR  UPDATE RESTAPI 1  2 " + err.Error())
				log.Println("details:")
				log.Println("action", valueEvent.Action)
				return resP, err
			}
			if len(ps) > 0 {
				for _, v := range ps {

					eventBPParams := []NameValue{NameValue{Name: "pk", Value: valueEvent2["id"]}, NameValue{Name: "old", Value: valueEvent2}}

					context := InstanceContext{}
					context.O = o
					context.Lua = lua.NewState()
					lua.OpenLibraries(context.Lua)
					RegisterAPI(context.Lua, o)
					RegisterBPMLUaAPI(nil, context.Lua, o)
					log.Println("start process", eventBPParams)
					_, _, _, _, err := context.CreateInstance(nil, v, user_id, eventBPParams, 0)
					if err != nil {

						resP.ErrorText = err.Error()
						//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 24 " + err.Error() + err2.Error()
						//	return resP, err2
						//}

						log.Println("ERROR UPDATE RESTAPI EVENT RUN BP " + err.Error())
						log.Println("details:")
						log.Println(valueEvent)
						return resP, err

					}
				}

			}
		}

	}
	return resP, nil
}

func DML(o orm.Ormer, user_id int64, t DMLRequest, ignoreGrants bool) (UpdateResponse, error) {

	delim := utils.GetDbStringDelimiter()
	type TRestrictEdits struct {
		RestrictEdit      int
		EntityTitle       string
		AttrTitle         string
		AttrCode          string
		RestrictEditError int
		AllowEdit2Roles   string
	}

	//decoder := json.NewDecoder(req.Body)
	//var t DMLRequest
	var resP UpdateResponse
	//err := decoder.Decode(&t)
	//	if err!=nil{
	//		return
	//	}
	resP.Error = 1

	//o := orm.NewOrm()
	//o.Using("default")

	//savepointName := "DML" + utils.Uuid2()
	//rollbackSql := "ROLLBACK TO " + savepointName
	//_, err := o.Raw("SAVEPOINT " + savepointName).Exec()
	//if err != nil {
	//	resP.ErrorText = err.Error()
	//	return resP, err
	//}

	//defer o.Rollback()

	for items := range t.Items {

		t.Items[items].origAction = t.Items[items].Action

		for _, value := range t.Items[items].Values {
			var arr []interface{}
			updateColumns := ""
			insertColumns := ""
			insertValues := ""

			entityId := int64(0)
			err := cached.O().Raw(utils.DbBindReplace("select id from entities where code=?"), t.Items[items].TableName).QueryRow(&entityId)

			if err != nil {
				resP.ErrorText = err.Error()
				//o.Rollback()

				//OPT
				//_, err2 := o.Raw(rollbackSql).Exec()
				//if err2 != nil {
				//	resP.ErrorText = "Error 5 1 " + err.Error() + err2.Error()
				//	return resP, err
				//}

				return resP, err
			}

			var dataTypes []TDataTypes
			_, err = cached.O().Raw(utils.DbBindReplace(`select ea.code as "code", dt.code as "data_type_code" from data_types dt,entity_attrs ea where ea.entity_id=? and ea.data_type_id=dt.id`), entityId).QueryRows(&dataTypes)

			if err != nil {
				resP.ErrorText = err.Error()
				//o.Rollback()

				//OPT
				//_, err2 := o.Raw(rollbackSql).Exec()
				//if err2 != nil {
				//	resP.ErrorText = "Error 5 " + err.Error() + err2.Error()
				//	return resP, err
				//}

				return resP, err
			}

			value, _ = replaceUUIDtoID(o, t.Items[items].TableName, value)
			//log.Println(t.Items[items].Action+ " - "+t.Items[items].TableName)

			if len(t.Items[items].UpsertAttrs) > 0 && t.Items[items].origAction == "upsert" {
				upCond := "1 = 1"
				upId := int64(0)
				var upVals []interface{}
				for _, upsertAttr := range t.Items[items].UpsertAttrs {
					if utils.CheckFieldRegexp(upsertAttr) != nil {
						resP.ErrorText = "Upsert Attribute Field Check Fail"

						//OPT
						//_, err2 := o.Raw(rollbackSql).Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 6 " + err.Error()
						//	return resP, err
						//}
						return resP, err
					}
					if os.Getenv("CRM_DEBUG_SQL") == "1" {
						log.Println("upsertAttr", upsertAttr)
					}
					upCond = upCond + " and " + upsertAttr + " = ?"
					upVals = append(upVals, value[upsertAttr])

				}
				upQl := "select id from " + delim + t.Items[items].TableName + delim + "  where " + upCond
				//log.Println("upql",upQl)
				err = o.Raw(utils.DbBindReplace(upQl), upVals).QueryRow(&upId)
				if !utils.IsNoRowFound(err) && err != nil {
					resP.ErrorText = "Upsert Fail " + err.Error()
					//OPT
					//_, err2 := o.Raw(rollbackSql).Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 7 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				} else if utils.IsNoRowFound(err) && err != nil {
					t.Items[items].Action = "insert"
					value["id"] = 0

				} else if err == nil {
					t.Items[items].Action = "update"
					value["id"] = upId
				}

			}

			restrictEdits := []TRestrictEdits{}

			if !ignoreGrants {

				_, err3 := cached.O().Raw(utils.DbBindReplace(`
			select
ea.restrict_edit as "restrict_edit",
e.title as "entity_title",
ea.title as "attr_title",
ea.code "attr_code",
ea.restrict_edit_error as "restrict_edit_error",
ea.allow_edit2roles as "allow_edit2roles"
from entity_attrs ea
join entities e on ea.entity_id=e.id
where
e.code=?
and (ea.restrict_edit=1 or ea.is_formula=1)`), t.Items[items].TableName).QueryRows(&restrictEdits)

				if err3 != nil {
					resP.ErrorText = "Error on get 3 " + err3.Error()
					return resP, err
				}
			}

			for fieldName, fieldValue := range value {

				skipEdit := false
				for _, restrictValue := range restrictEdits {
					if restrictValue.AttrCode == fieldName {

						if restrictValue.RestrictEditError == 1 && restrictValue.RestrictEdit == 1 {
							resP.ErrorText = "Access denied for edit field `" + restrictValue.AttrCode + "`"
							//OPT
							//_, err2 := o.Raw(rollbackSql).Exec()
							//if err2 != nil {
							//	resP.ErrorText = "Error 71 " + err.Error() + err2.Error()
							//	resP.ErrorCode = "71"
							//	return resP, err
							//}
							return resP, err
						} else {
							skipEdit = true
						}
					}
				}

				if skipEdit {
					continue
				}

				if fieldName != "sys$uuid" {
					err := utils.CheckFieldRegexp(fieldName) //Check SQL Injection
					if err != nil {
						resP.ErrorText = err.Error()
						//OPT
						//_, err2 := o.Raw(rollbackSql).Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 8 " + err.Error() + err2.Error()
						//	return resP, err
						//}
						return resP, err
					}
				}

				datatype_code := getDataTypeByAttrCode(fieldName, dataTypes)
				if !strings.HasSuffix(fieldName, "$") && !strings.HasPrefix(fieldName, "_") && datatype_code != "current_and_on_update_datetime" {
					if fieldValue == nil {
						updateColumns = updateColumns + " " + delim + fieldName + delim + "=NULL,"
						insertColumns = insertColumns + fieldName + ","
						insertValues = insertValues + "NULL,"
					} else {

						if fieldName != "id" {
							updateColumns = updateColumns + " " + delim + fieldName + delim + "=?,"
							insertColumns = insertColumns + fieldName + ","
							insertValues = insertValues + "?,"

							if datatype_code == "longtext" && utils.GetDbDriverType() == orm.DROracle {
								arr = append(arr, []byte(fieldValue.(string)))
							} else {
								arr = append(arr, fieldValue)
							}

							//arr = append(arr, fieldValue)
						}

					}
				}
			}

			updateColumns = strings.TrimRight(updateColumns, ",")
			insertColumns = strings.TrimRight(insertColumns, ",")
			insertValues = strings.TrimRight(insertValues, ",")

			//lastInsertId := int64(0)

			if t.Items[items].Action == "insert" {

				lastInsertId := int64(0)

				sql := "insert into " + t.Items[items].TableName + " ( " + insertColumns + " ) values ( " + insertValues + ")"

				lastInsertId, err = utils.DbInsert(o, sql, arr...)

				if os.Getenv("CRM_DEBUG_SQL") == "1" {
					log.Println("insert sql=" + sql)
				}

				if err != nil {
					resP.ErrorText = err.Error()
					//OPT
					//_, err2 := o.Raw(rollbackSql).Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 9 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				/*
					if utils.GetDbDriverType() == orm.DRPostgres {

						sql := "insert into " + t.Items[items].TableName + " ( " + insertColumns + " ) values ( " + insertValues + ") returning id"
						log.Println("insert sql=" + sql)
						err = o.Raw(utils.DbBindReplace(sql), arr).QueryRow(&lastInsertId)
					}else {
						sql := "insert into " + t.Items[items].TableName + " ( " + insertColumns + " ) values ( " + insertValues + ")"
						log.Println("insert sql=" + sql)
						i, err := o.Raw(utils.DbBindReplace(sql), arr).Exec()

						if err!=nil{
							resP.ErrorText = err.Error()
							_,err2 := o.Raw("ROLLBACK TO DML").Exec()
							if err2!=nil{
								resP.ErrorText = "Error 9 "+ err.Error() + err2.Error()
								return resP,err
							}
							return resP,err
						}

						lastInsertId,err = i.LastInsertId()
					}*/

				if err != nil {
					resP.ErrorText = err.Error()
					//OPT
					//_, err2 := o.Raw(rollbackSql).Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 9 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}
				if os.Getenv("CRM_DEBUG_SQL") == "1" {
					log.Println("lastInsertId=", lastInsertId)
				}
				value["id"] = lastInsertId

				sysUUID := ""
				o.Raw(utils.DbBindReplace("select sys$uuid from  "+t.Items[items].TableName+" where id=?"), lastInsertId).QueryRow(&sysUUID)
				if os.Getenv("CRM_DEBUG_SQL") == "1" {
					log.Println("sys$UUID=" + sysUUID)
				}
				setDefaultUserId(o, t.Items[items].TableName, user_id, lastInsertId, value)

				setDefaultByUserVar(o, t.Items[items].TableName, user_id, lastInsertId)
				//log.Println("sql="+sql)

				if !ignoreGrants {
					err = utils.CheckGrantOfEntity(o, user_id, t.Items[items].TableName, "is_insert", lastInsertId)

					if err != nil {
						resP.ErrorText = err.Error()
						//OPT
						//_, err2 := o.Raw(rollbackSql).Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 10 " + err.Error() + err2.Error()
						//	return resP, err
						//}
						return resP, err
					}
				}

				rqErrorAttrs, errCode, errText, errRecom, errDscr, err := DMLValidate(t.Items[items].Action, o, t.Items[items].TableName, user_id, lastInsertId, value)
				if err != nil {
					resP.RqErrorAttrs = rqErrorAttrs
					resP.ErrorText = errText
					resP.ErrorCode = errCode
					resP.ErrorRecom = errRecom
					resP.ErrorDscr = errDscr
					//OPT
					//_, err2 := o.Raw(rollbackSql).Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 11 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				/*doInsertEvent(o,t.Items[items].TableName,user_id,lastInsertId,value);
				if err!=nil{
					resP.ErrorText = errText
					_,err2 := o.Raw("ROLLBACK TO DML").Exec()
					if err2!=nil{
						resP.ErrorText = "Error 12 "+ err.Error() + err2.Error()
						return resP,err
					}
					return resP,err
				}*/
				resP.Items = append(resP.Items, UpdateResponse_Item{SysUUID: sysUUID, LastInsertId: lastInsertId, TableName: t.Items[items].TableName, Action: t.Items[items].Action})

			} else if t.Items[items].Action == "update" {

				sql := "update " + t.Items[items].TableName + " main  set " + updateColumns + " where main.id=? " + utils.UpdateLimitByEntityCode(o, t.Items[items].TableName, user_id)
				if os.Getenv("CRM_DEBUG_SQL") == "1" {
					log.Println("update sql=" + sql)
				}

				sql_lastData := "select * from " + delim + t.Items[items].TableName + delim + " where id=?"
				_, err := o.Raw(utils.DbBindReplace(sql_lastData), value["id"]).Values(&t.Items[items].lastData)

				//log.Println("lastData =")
				//log.Println(t.Items[items].lastData)
				if err != nil {
					resP.ErrorText = err.Error()
					log.Println("Error 13", sql_lastData, value["id"])
					//OPT
					//_, err2 := o.Raw(rollbackSql).Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 13 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}
				idInt, err := strconv.Atoi(fmt.Sprintf("%v", value["id"]))

				if !ignoreGrants {
					err = utils.CheckGrantOfEntity(o, user_id, t.Items[items].TableName, "is_update", int64(idInt))
					if err != nil {
						resP.ErrorText = err.Error()
						log.Println("Error 14 1 ", err.Error())
						//OPT
						//_, err2 := o.Raw(rollbackSql).Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 14 " + err.Error() + err2.Error()
						//	return resP, err
						//}
						return resP, err
					}
				}

				/*Warning
				if RestCheckDBPanicDetails(err ,res , errCode, errText, o ) {
					o.Rollback()
					return
				}*/

				pr, err := o.Raw(utils.DbBindReplace(sql)).Prepare()
				defer pr.Close()
				//_, err = o.Raw(utils.DbBindReplace(sql), arr, value["id"]).Exec()
				if err != nil {
					log.Println("Error 16.1 SQL = ", sql)
					resP.ErrorText = err.Error()
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 16 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}
				arr = append(arr, value["id"])
				_, err = pr.Exec(arr...)

				if err != nil {
					log.Println("Error 16.2 SQL = ", sql,err.Error(),arr)
					resP.ErrorText = err.Error()
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 16 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				rqErrorAttrs, errCode, errText, errRecom, errDscr, err := DMLValidate(t.Items[items].Action, o, t.Items[items].TableName, user_id, int64(idInt), value)

				if err != nil {

					if errText == "" {
						errText = err.Error()
					}
					resP.RqErrorAttrs = rqErrorAttrs
					resP.ErrorText = errText
					resP.ErrorCode = errCode
					resP.ErrorRecom = errRecom
					resP.ErrorDscr = errDscr
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 15 " + err.Error() + err2.Error()
					//	return resP, err
					//}

					return resP, err
				}

				err = setOnUpdateByUserVar(o, t.Items[items].TableName, user_id, idInt)

				if err != nil {

					if errText == "" {
						errText = err.Error()
					}

					resP.ErrorText = errText
					resP.ErrorCode = errCode
					resP.ErrorRecom = errRecom
					resP.ErrorDscr = errDscr
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 151 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				//log.Println("idInt=")
				//log.Println(idInt)
				//go DoTriggerOnUpdateOne(o, entityId, user_id, idInt)

				/*rAffected,err := rs.RowsAffected()
				if RestCheckDBPanic(err ,res ,o ) {
					return
				}
				if rAffected == 0{
					WriteRestCheckPanic("Access Denied "+strconv.Itoa(int(rAffected)),"403",res)
					return
				}*/

				resP.Items = append(resP.Items, UpdateResponse_Item{LastInsertId: int64(idInt), TableName: t.Items[items].TableName, Action: t.Items[items].Action, lastData: t.Items[items].lastData})

			} else if t.Items[items].Action == "delete" {

				if value["id"] == nil {
					err = errors.New("Cannot delete empty PK value")
					resP.ErrorText = err.Error()
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 17 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				idInt, err := strconv.Atoi(fmt.Sprintf("%v", value["id"]))

				if !ignoreGrants {
					err = utils.CheckGrantOfEntity(o, user_id, t.Items[items].TableName, "is_delete", int64(idInt))
					if err != nil {
						resP.ErrorText = err.Error()
						//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
						//if err2 != nil {
						//	resP.ErrorText = "Error 18 " + err.Error() + err2.Error()
						//	return resP, err
						//}
						return resP, err
					}
				}

				sql := "delete from " + t.Items[items].TableName + " where id=?"
				log.Println("delete sql=" + sql)
				_, err = o.Raw(utils.DbBindReplace(sql), value["id"]).Exec()

				if err != nil {
					resP.ErrorText = err.Error()
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 19 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				resP.Items = append(resP.Items, UpdateResponse_Item{LastInsertId: int64(idInt), TableName: t.Items[items].TableName, Action: t.Items[items].Action})

			}

			///LOGGING

			if utils.CheckEntityFullAudit(o, t.Items[items].TableName) {
				/*lid := ""
				if t.Items[items].Action == "insert" {
					lid = strconv.Itoa(int(lastInsertId))
					log.Println("sss", lid, lastInsertId, value["id"])
				} else {

				}*/
				//lid := fmt.Sprintf("%v", value["id"])
				iid, err := utils.DbInsert(o, utils.DbBindReplace("insert into table_logs (user_id,table_name,action,pk,"+utils.GetDbStringDelimiter()+"version"+utils.GetDbStringDelimiter()+") values (?,?,?,?,(select value from params where code='version'))"), user_id, t.Items[items].TableName, t.Items[items].Action, value["id"])

				if err != nil {
					log.Println("Error on insert into table_log_dtls 1", user_id, t.Items[items].TableName, t.Items[items].Action, value["id"])
					resP.ErrorText = err.Error()
					//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
					//if err2 != nil {
					//	resP.ErrorText = "Error 20 " + err.Error() + err2.Error()
					//	return resP, err
					//}
					return resP, err
				}

				for fieldName, fieldValue := range value {

					if fieldName != "sys$uuid" {
						err := utils.CheckFieldRegexp(fieldName) //Check SQL Injection
						if err != nil {
							resP.ErrorText = err.Error()
							//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
							//if err2 != nil {
							//	resP.ErrorText = "Error 21 " + err.Error() + err2.Error()
							//	return resP, err
							//}
							return resP, err
						}
					}

					if !strings.HasPrefix(fieldName, "_") {

						var last_value interface{}
						if len(t.Items[items].lastData) > 0 {
							last_value = t.Items[items].lastData[0][fieldName]

						}

						if fieldValue == nil {

							_, err := o.Raw(utils.DbBindReplace("insert into table_log_dtls (col,val,oldval,log_id) values (?,NULL,?, ?)"), fieldName, last_value, iid).Exec()
							if err != nil {
								log.Println("Error on insert into table_log_dtls 2", fieldName, iid)
								resP.ErrorText = err.Error()
								//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
								//if err2 != nil {
								//	resP.ErrorText = "Error 22 " + err.Error() + err2.Error()
								//	return resP, err
								//}
								return resP, err
							}
						} else {
							if fieldName != "id" {
								_, err := o.Raw(utils.DbBindReplace("insert into table_log_dtls (col,val,oldval,log_id) values (?,?,?,?)"), fieldName, fmt.Sprintf("%v", fieldValue), last_value, iid).Exec()
								if err != nil {
									log.Println("Error on insert into table_log_dtls 3", fieldName, iid, err.Error())
									resP.ErrorText = err.Error()
									//_, err2 := o.Raw("ROLLBACK TO DML").Exec()
									//if err2 != nil {
									//	resP.ErrorText = "Error 23 " + err.Error() + err2.Error()
									//	return resP, err
									//}
									return resP, err
								}
							}
						}
					}
				}
			}

			///LOGGING
		}
	}
	//o.Commit()

	//END EVENT PROCESSING

	//o.Raw("RELEASE SAVEPOINT DML").Exec()
	resP.Error = 0
	resP.ErrorText = "OK"

	//jsonData, err := json.Marshal(resP)
	//fmt.Fprint(res,string(jsonData))

	//utils.ClearInterface(&resP)
	//utils.ClearInterface(&jsonData)
	utils.ClearInterface(&t)

	return resP, nil

}
