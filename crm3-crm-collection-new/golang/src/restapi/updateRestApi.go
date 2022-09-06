package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	"github.com/julienschmidt/httprouter"
)

type UpdateRequest struct {
	Items     []orm.Params `json:"items"`
	TableName string       `json:"table_name"`
}

func UpdateRestApi(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if RestCheckAuth(res, req) {
		return
	}

	stringDelimiter := utils.GetDbStringDelimiter()

	decoder := json.NewDecoder(req.Body)
	var t UpdateRequest
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	o := orm.NewOrm()
	o.Using("default")
	o.Begin()

	for _, element := range t.Items {
		var arr []interface{}
		updateColumns := ""
		insertColumns := ""
		insertValues := ""

		for fieldName, fieldValue := range element {

			if fieldName != "sys$uuid" {
				err := utils.CheckFieldRegexp(fieldName) //Check SQL Injection
				if RestCheckDBPanic(err, res, o) {
					o.Rollback()
					return
				}
			}

			if fieldName == "_table_name_" {
				t.TableName = regexp.QuoteMeta(fieldValue.(string))

				if RestCheckDBPanic(err, res, o) {
					o.Rollback()
					return
				}

			} else if !strings.HasPrefix(fieldName, "_") {

				if fieldValue == nil {
					updateColumns = updateColumns + " " + stringDelimiter + fieldName + stringDelimiter + "=NULL,"
					insertColumns = insertColumns + fieldName + ","
					insertValues = insertValues + "NULL,"
				} else {
					if fieldName != "id" {
						updateColumns = updateColumns + " " + stringDelimiter + fieldName + stringDelimiter + "=?,"
						insertColumns = insertColumns + fieldName + ","
						insertValues = insertValues + "?,"
						arr = append(arr, fieldValue)
					}

				}
			}
		}

		updateColumns = strings.TrimRight(updateColumns, ",")
		insertColumns = strings.TrimRight(insertColumns, ",")
		insertValues = strings.TrimRight(insertValues, ",")
		//log.Println("id ="+ element["id"].(string) )
		if element["id"] == nil || element["id"] == "0" {
			sql := "insert into " + t.TableName + " ( " + insertColumns + " ) values ( " + insertValues + ")"

			if os.Getenv("CRM_DEBUG_SQL") == "1" {
				log.Println("insert sql=" + sql)
			}
			i, err := o.Raw(utils.DbBindReplace(sql), arr).Exec()
			if RestCheckDBPanic(err, res, o) {
				o.Rollback()
				return
			}
			lastInsertId, err := i.LastInsertId()
			log.Println("sql=" + sql)
			if RestCheckDBPanic(err, res, o) {
				o.Rollback()
				return
			}

			err = utils.CheckGrantOfEntity(o, utils.UserId(req), t.TableName, "is_insert", lastInsertId)

			//doInsertEvent(o,t.TableName,utils.UserId(req),lastInsertId,element);
			luautils.DoLuaValidator(o, "insert", t.TableName, utils.UserId(req), lastInsertId, element)

			if err != nil {
				fmt.Fprint(res, "{\"error\":\"1\"}")
				log.Println(err)
			} else {
				fmt.Fprint(res, "{\"error\":\"0\", \"items\":[{ \"id\":"+strconv.Itoa(int(lastInsertId))+"}] }")
			}
		} else {
			sql := "update " + t.TableName + " set " + updateColumns + " where id=?"
			//log.Println("update sql="+sql)

			log.Println("ELEMENT=")
			log.Println(element["id"])
			idInt, err := strconv.Atoi(element["id"].(string))

			err = utils.CheckGrantOfEntity(o, utils.UserId(req), t.TableName, "is_update", int64(idInt))

			luautils.DoLuaValidator(o, "update", t.TableName, utils.UserId(req), int64(idInt), element)
			_, err = o.Raw(utils.DbBindReplace(sql), arr, element["id"]).Exec()
			if err != nil {
				fmt.Fprint(res, "{\"error\":\"1\"}")
				log.Println(err)
			} else {
				fmt.Fprint(res, "{\"error\":\"0\"}")
			}
		}

		//o.QueryTable(t.TableName).Filter("id", element["id"]).Update(element)
	}

	o.Commit()

}
