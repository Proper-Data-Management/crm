package luautils

import (
	"errors"
	"log"
	"os"
	"runtime/debug"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	lua "github.com/Shopify/go-lua"

	"strings"
)

func BindInputVariablesToLuaValidator(input interface{}, user_id int64, state *lua.State) error {

	o := orm.NewOrm()
	o.Using("default")
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("BindInputVariablesToLuaValidator input", input)
	}

	DeepPush(state, input)
	state.SetGlobal("input")

	state.PushInteger(int(user_id))
	state.SetGlobal("user_id")

	return nil
}

func BindOutputErrorVariablesToLuaValidator(state *lua.State, script LuaResp) error {

	o := orm.NewOrm()
	o.Using("default")

	state.Global("error")

	value := state.ToBoolean(1)
	if value {
		return errors.New(script.Title)
	}
	return nil
}

func runSQLValidator(request orm.Params, user_id int64, script string) error {

	log.Println(request)

	var arr_columns []string
	var arr_values []interface{}
	for k, v := range request {
		if k == "title" {
			arr_columns = append(arr_columns, "? `"+k+"` ")
			arr_values = append(arr_values, v)
		}
	}

	o := orm.NewOrm()
	o.Using("default")
	sql := "select count(*) cnt from (select " + strings.Join(arr_columns, ",") + " from dual) as main where 1=1 and title like ''"
	//sql = "with (select 1  from (select ? title) main) as s select count(1)cnt, s from dual where s.title=''";
	log.Println("runSQLValidator sql " + sql)
	//cnt := []int64;
	_, err := o.Raw(utils.DbBindReplace(sql), arr_values).Exec()

	if err != nil {
		log.Println("runSQLValidator err " + err.Error())
		return errors.New("runSQLValidator err " + err.Error())
	}
	//	if len(cnt)>0 {
	//		return errors.New("Validate Error")
	//	}

	return nil
}

func runLuaValidator(o orm.Ormer, action string, request interface{}, user_id int64, script LuaResp) error {
	l := lua.NewState()
	lua.OpenLibraries(l)

	//loadLuas(l)

	RegisterAPI(l, o)

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("Startttttt3333")
	}

	//****************************
	//log.Println("Script Body "+row.Script)
	//l.PushString("KRUTO GOI!!!!")
	//l.SetGlobal("guid")

	err := BindInputVariablesToLuaValidator(request, user_id, l)

	sys := make(map[string]interface{})
	//sys["instance_id"] = instanceId

	sys["user_id"] = user_id
	sys["action"] = action

	DeepPush(l, sys)
	l.SetGlobal("sys")

	DeepPush(l, request)
	l.SetGlobal("input")

	if err != nil {
		log.Println("error bind variables  " + err.Error())
		debug.PrintStack()
		//o.Rollback()
		return err
	}

	if err := lua.DoString(l, script.Script); err != nil {
		log.Println("error lua  " + err.Error())
		log.Println("script = " + script.Script)
		debug.PrintStack()
		utils.ErrorWriteUser("BindOutputErrorVariablesToLuaValidator", "nothing", user_id, err)
		return err
	}

	err = BindOutputErrorVariablesToLuaValidator(l, script)

	if err != nil {
		log.Println("error on " + err.Error() + " " + script.Script + " " + script.Code)
		log.Println("script = " + script.Script)
		debug.PrintStack()
		return err
	}
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("LUA SCRIPT POINT DONE")
	}

	return nil
}

type LuaResp struct {
	Script string `json:"script"`
	Title  string `json:"title"`
	Code   string `json:"code"`
	Recom  string `json:"recom"`
	Dscr   string `json:"dscr"`
}

func DoLuaValidator(o orm.Ormer, action string, tableName string, userId int64, lastInsertId int64, values orm.Params) (string, string, string, string, error) {

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("doInsertEvent start")
	}

	var sqls []LuaResp
	_, err := o.Raw(utils.DbBindReplace(`select ev.lua_script as "script",
	ev.title as "title",
	ev.code as "code",
	ev.recom as "recom",
	ev.dscr as "dscr" from entity_validators ev,entities e,entity_validator_types evt where e.id=ev.entity_id and e.code=? and ev.type_id=evt.id and evt.code='lua'`), tableName).QueryRows(&sqls)

	if err != nil {
		log.Println("DoLuaValidator err1 " + err.Error())
		return "", "", "", "", err
	}

	for _, lua_script := range sqls {

		err = runLuaValidator(o, action, values, userId, lua_script)

		if err != nil {
			return lua_script.Code, lua_script.Title, lua_script.Recom, lua_script.Dscr, err
		}

	}
	return "", "", "", "", nil

}

func DoSQLValidator(o orm.Ormer, tableName string, userId int64, lastInsertId int64, values orm.Params) (string, string, error) {

	//log.Println("DoSQLValidator start")
	type sqlResp struct {
		Query string `json:"query"`
		Title string `json:"title"`
		Code  string `json:"code"`
	}
	var sqls []sqlResp
	_, err := o.Raw(utils.DbBindReplace(`select ev.sql_query as "query",ev.title as "title",ev.code as "code" 
	from entity_validators ev,entities e,entity_validator_types evt where e.id=ev.entity_id and e.code=? and ev.type_id=evt.id and evt.code='sql'`), tableName).QueryRows(&sqls)

	if err != nil {
		log.Println("DoSQLValidator err1 " + err.Error())
		return "", "", err
	}

	for _, sql_query := range sqls {
		err = runSQLValidator(values, userId, sql_query.Query)
		if err != nil {
			return sql_query.Code, sql_query.Title, err
		}
	}
	return "", "", nil

}

func DMLRqCheck(action string, o orm.Ormer, tableName string, userId int64, lastInsertId int64, values orm.Params) ([]TRqErrorAttrs, error) {

	var codes []TRqErrorAttrs

	var rqErrorAttrs []TRqErrorAttrs
	_, err := cached.O().Raw(utils.DbBindReplace("select ? as \"table_name\", title as \"title\", code  as \"code\" from entity_attrs ea where ea.entity_id=(select id from entities e where e.code=?) and ea.rq = 1"), tableName, tableName).QueryRows(&codes)
	if err != nil {
		return rqErrorAttrs, err
	}
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("DMLRqCheck Values", values)
	}

	errText := ""
	for _, v := range codes {
		_, filled := values[v.Code]
		print("filled1",filled,values[v.Code])
		if filled && (values[v.Code] == "" || values[v.Code] == nil) {		
			rqErrorAttrs = append(rqErrorAttrs, v)
			errText = "RQATTRS"			
		}else if ! filled && action == "insert"{
			rqErrorAttrs = append(rqErrorAttrs, v)
			errText = "RQATTRS"		
		}

	}

	if errText != "" {
		log.Println("RQATTRS",rqErrorAttrs)
		return rqErrorAttrs, errors.New(errText)
	}
	return rqErrorAttrs, nil
}

func DMLValidate(action string, o orm.Ormer, tableName string, userId int64, lastInsertId int64, values orm.Params) ([]TRqErrorAttrs, string, string, string, string, error) {

	var rqErrorAttrs []TRqErrorAttrs
	str1, str2, recom, dscr, err := DoLuaValidator(o, action, tableName, userId, lastInsertId, values)
	if err != nil {
		return rqErrorAttrs, str1, str2, recom, dscr, err
	}

	str1, str2, err = DoSQLValidator(o, tableName, userId, lastInsertId, values)
	if err != nil {
		return rqErrorAttrs, str1, str2, recom, dscr, err
	}

	rqErrorAttrs, err = DMLRqCheck(action, o, tableName, userId, lastInsertId, values)
	if err != nil {
		return rqErrorAttrs, "required_variable", err.Error(), "", "", err
	}
	return nil, "", "", "", "", nil

}
