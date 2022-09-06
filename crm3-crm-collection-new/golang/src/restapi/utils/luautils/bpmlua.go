package luautils

import (
	"encoding/json"
	"fmt"
	"os"

	lua "github.com/Shopify/go-lua"

	"log"
	"runtime/debug"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"

	//"git.dar.kz/crediton-3/src-mfo/src/pkg"
	//"git.dar.kz/crediton-3/src-mfo/src/lib/lua/stateorm"
	"net/http"

	"errors"
	"strings"
)

type LuaContext struct {
	Name string
	o    orm.Ormer
	req  *http.Request
}

//BindLocalVariablesToByPoint
/*
func (context *InstanceContext) BindLocalVariablesToByPoint(req *http.Request, instanceId, pointId, userId int64) error {

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(`select v.code as "name",'' "value",dt.code as "type"
		from
		bp_points po,
		bp_process_vars v,
		data_types dt
		where
		v.data_type_id=dt.id
		and
		po.process_id=v.process_id
		and po.id=?`), pointId).QueryRows(&rows)
	if err != nil {
		return err
	}

	var action_entity_ids []int

	_, err = context.O.Raw(utils.DbBindReplace(`select p.action_entity_id from bp_points po
  join bp_processes p on p.id=po.process_id
where po.id=?`), pointId).QueryRows(&action_entity_ids)

	if err != nil {
		return err
	}

	vars := make(map[string]interface{})
	sys := make(map[string]interface{})
	for index, element := range rows {
		//isArray := false
		rows[index].Value, _, err = context.GetProcessVarByInstance(instanceId, element.Name)
		if err != nil {
			log.Println("BindLocalVariablesToLua GetProcessVarByInstance")
			return err
		}
		vars[element.Name] = rows[index].Value
	}
	sys["instance_id"] = instanceId
	sys["user_id"] = userId

	if len(action_entity_ids) > 0 {
		sys["action_entity_id"] = action_entity_ids[0]
	}

	var reqArrGet = make(map[string]string)

	if req != nil {
		for k, v := range req.Form {
			reqArrGet[k] = v[0]
		}
		var reqArrHeader = make(map[string]string)
		for k, v := range req.Header {
			reqArrHeader[k] = v[0]
		}
	}
	var requestMap = make(map[string]interface{})
	requestMap["get"] = reqArrGet
	requestMap["user_id"] = userId
	DeepPush(context.Lua, requestMap)
	context.Lua.SetGlobal("request")

	DeepPush(context.Lua, vars)
	context.Lua.SetGlobal("var")
	//log.Println("var = ")
	//log.Println(vars)

	DeepPush(context.Lua, sys)
	context.Lua.SetGlobal("sys")

	context.Lua.PushInteger(int(userId))
	context.Lua.SetGlobal("user_id")

	utils.ClearInterface(&vars)

	return nil
}
*/

//Раньше перед запуском LUA скрипта почему-то запускался метод  BindLocalVariablesToLua,
//который биндил все поля подрял не учитывая is_input
/*func (context *InstanceContext) BindInputVariablesToLua2(req *http.Request, instanceId, point int64, userId int64) error {

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(`select v.code as "name",'' "value",dt.code as "type"
		from
		    bp_points po,
			bp_process_vars v,
			bp_processes p,
			data_types dt,
			bp_point_vars pv
		where pv.point_id=po.id and pv.var_id=v.id and pv.is_input =1 and
		v.data_type_id=dt.id and p.id=v.process_id and po.process_id=p.id
		and po.id=?
		`), point).QueryRows(&rows)
	if err != nil {
		return err
	}

	action_entity_id := 0

	err = context.O.Raw(utils.DbBindReplace(`select p.action_entity_id from bp_instances i
  join bp_processes p on p.id=i.process_id
where i.id=?`), instanceId).QueryRow(&action_entity_id)

	if err != nil {
		return err
	}

	vars := make(map[string]interface{})
	sys := make(map[string]interface{})
	for index, element := range rows {
		//isArray := false
		rows[index].Value, _, err = context.GetProcessVarByInstance(instanceId, element.Name)
		if err != nil {
			log.Println("BindLocalVariablesToLua GetProcessVarByInstance")
			return err
		}
		vars[element.Name] = rows[index].Value
	}
	sys["instance_id"] = instanceId
	sys["user_id"] = userId
	sys["action_entity_id"] = action_entity_id

	var reqArrGet = make(map[string]string)

	if req != nil {
		for k, v := range req.Form {
			reqArrGet[k] = v[0]
		}
		var reqArrHeader = make(map[string]string)
		for k, v := range req.Header {
			reqArrHeader[k] = v[0]
		}
	}
	var requestMap = make(map[string]interface{})
	requestMap["get"] = reqArrGet
	requestMap["user_id"] = userId
	DeepPush(context.Lua, requestMap)
	context.Lua.SetGlobal("request")

	DeepPush(context.Lua, vars)
	context.Lua.SetGlobal("var")
	//log.Println("var = ")
	//log.Println(vars)

	DeepPush(context.Lua, sys)
	context.Lua.SetGlobal("sys")

	context.Lua.PushInteger(int(userId))
	context.Lua.SetGlobal("user_id")

	utils.ClearInterface(&vars)

	return nil
}*/

//Третяя версия, которая берет переменные из памяти
func (context *InstanceContext) BindInputVariablesToLua3(req *http.Request, instanceId, point int64, userId int64) error {

	action_entity_id := int64(0)
	err := context.O.Raw(utils.DbBindReplace(`select p.action_entity_id from bp_instances i
  join bp_processes p on p.id=i.process_id
where i.id=?`), instanceId).QueryRow(&action_entity_id)

	if err != nil {
		return err
	}

	vars := make(map[string]interface{})
	sys := make(map[string]interface{})
	for _, element := range context.InstanceVars[instanceId] {
		//isArray := false
		if err != nil {
			log.Println("BindLocalVariablesToLua GetProcessVarByInstance")
			return err
		}
		vars[element.Name] = element.Value
	}
	sys["instance_id"] = instanceId
	sys["user_id"] = userId
	sys["action_entity_id"] = action_entity_id

	var reqArrGet = make(map[string]string)

	if req != nil {
		for k, v := range req.Form {
			reqArrGet[k] = v[0]
		}
		var reqArrHeader = make(map[string]string)
		for k, v := range req.Header {
			reqArrHeader[k] = v[0]
		}
	}
	var requestMap = make(map[string]interface{})
	requestMap["get"] = reqArrGet
	requestMap["user_id"] = userId
	DeepPush(context.Lua, requestMap)
	context.Lua.SetGlobal("request")

	DeepPush(context.Lua, vars)
	context.Lua.SetGlobal("var")
	//log.Println("var = ")
	//log.Println(vars)

	DeepPush(context.Lua, sys)
	context.Lua.SetGlobal("sys")

	context.Lua.PushInteger(int(userId))
	context.Lua.SetGlobal("user_id")

	utils.ClearInterface(&vars)

	return nil
}

func (context *InstanceContext) BindInputVariablesToLua(instanceId, pointId, userId, jrnId int64, state *lua.State) error {

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("BindInputVariablesToLua", instanceId, pointId, userId, jrnId)
	}

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(`select v.code as "name",'' "value",dt.code as "type"
	from	bp_point_vars pv,
			bp_process_vars v,
			data_types dt  where v.id=pv.var_id and pv.is_input=1 and pv.point_id=? and v.data_type_id=dt.id`), pointId).QueryRows(&rows)
	if err != nil {
		return err
	}

	for index, element := range rows {

		isArray := false
		rows[index].Value, isArray, err = context.GetProcessVarByInstance(instanceId, element.Name)
		if err != nil {
			return err
		}
		if rows[index].Type == "struct" || isArray {

			//log.Println("WWW")
			//log.Println(rows[index].Value)
			//decoder := json.NewDecoder(strings.NewReader(rows[index].Value))
			//var request  interface {}
			//_ = decoder.Decode(&request)
			DeepPush(state, rows[index].Value)
			//log.Println(request)
		} else {
			state.PushString(rows[index].Value.(string))
		}
		state.SetGlobal(element.Name)
	}

	state.PushInteger(int(userId))
	state.SetGlobal("user_id")

	if jrnId != 0 {

		//		err = WritePointJRNDTL(instanceId, userId, pointId, jrnId, rows)
		//		if err != nil {
		//			log.Println(err)
		//			debug.PrintStack()
		//			return err
		//		}
	}

	return nil
}

/*
func (context *InstanceContext) InitContextDefaultVariables(instanceId, processId, userId int64) error {


	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("InitContextDefaultVariables", instanceId, processId, userId)
	}

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(

		`select v.code as "name",dt.code "type",'' "value" from
		bp_processes pr
		join bp_process_vars v on v.process_id=pr.id
		join data_types dt on dt.id=v.data_type_id
		where pr.id=?
		`), processId).QueryRows(&rows)
	if err != nil {
		return err
	}

	if err != nil {
		fmt.Println("error on ReadContextVariableToLua", err)
		return err
	}

	for _, row := range rows {

		//log.Println("key", key)
		if context.InstanceVars[instanceId] == nil {
			context.InstanceVars[instanceId] = append(context.InstanceVars[instanceId],
				NameValue{Name: row.Name, Value: ""})
		}

		//if context.InstanceVars[instanceId][key].Value == nil {
		//	context.InstanceVars[instanceId][key].Value = ""
		//}

	}

	return nil
}
*/

//Считываем из Lua в Контекст
func (context *InstanceContext) ReadContextVariableToLua(instanceId, processId, userId int64) error {

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("ReadContextVariableToLua", instanceId, processId, userId)
	}

	var reqArrGet = make(map[string]string)

	if context.Req != nil {
		for k, v := range context.Req.Form {
			reqArrGet[k] = v[0]
		}
		var reqArrHeader = make(map[string]string)
		for k, v := range context.Req.Header {
			reqArrHeader[k] = v[0]
		}
	}

	var action_entity_ids []int

	_, err := context.O.Raw(utils.DbBindReplace(`select p.action_entity_id from bp_processes p
  
where p.id=?`), processId).QueryRows(&action_entity_ids)

	rows := []NameValue{}
	vars := make(map[string]interface{})
	sys := make(map[string]interface{})
	sys["instance_id"] = instanceId
	sys["user_id"] = userId

	if len(action_entity_ids) > 0 {
		sys["action_entity_id"] = action_entity_ids[0]
	}

	_, err = context.O.Raw(utils.DbBindReplace(

		`select v.code as "name",dt.code "type",'' "value",v.len as "len" from
		bp_processes pr
		join bp_process_vars v on v.process_id=pr.id
		join data_types dt on dt.id=v.data_type_id
		where pr.id=?
		`), processId).QueryRows(&rows)
	if err != nil {
		return err
	}

	if err != nil {
		fmt.Println("error on ReadContextVariableToLua", err)
		return err
	}

	//Default Values, например, строка, число всегда ""
	for _, element := range rows {
		//log.Println("SSSS", element)
		vars[element.Name] = ""
	}

	for key, element := range context.InstanceVars[instanceId] {
		vars[element.Name] = fmt.Sprintf("%v", context.InstanceVars[instanceId][key].Value)
	}
	//log.Println("vars", vars)

	if vars != nil {
		DeepPush(context.Lua, vars)
		context.Lua.SetGlobal("var")
	}

	DeepPush(context.Lua, sys)
	context.Lua.SetGlobal("sys")

	var requestMap = make(map[string]interface{})
	requestMap["get"] = reqArrGet
	requestMap["user_id"] = userId
	DeepPush(context.Lua, requestMap)
	context.Lua.SetGlobal("request")

	return nil
}

//Считываем из Lua в Контекст
func (context *InstanceContext) ReadFromLuaToContext(instanceId, pointId, userId int64) error {

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("ReadFromLuaToContext", instanceId, pointId, userId)
	}

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(

		`select v.code as "name",dt.code "type",'' "value",v.len as "len" from
		bp_processes pr
		join bp_points p on p.process_id=pr.id  
		join bp_process_vars v on v.process_id=pr.id
		join data_types dt on dt.id=v.data_type_id
		where p.id=?
		`), pointId).QueryRows(&rows)
	if err != nil {
		return err
	}

	context.Lua.Global("var")

	v, err := PullTable(context.Lua, context.Lua.Top())
	data, err := json.Marshal(v)
	//log.Println("string(data)", string(data), v)
	var s map[string]interface{}
	err = json.Unmarshal(data, &s)
	//log.Println("string(data)", s)

	if err != nil {
		fmt.Println("error on BindOutputVariablesFromLua", err)
		return err
	}

	context.InstanceVars[instanceId] = rows
	for index, _ := range context.InstanceVars[instanceId] {
		if context.InstanceVars[instanceId][index].Type == "struct" {
			//data, err := json.Marshal(s)
			//log.Println("string(data)", string(data), s[rows[index].Name])

			context.InstanceVars[instanceId][index].Value = s[rows[index].Name]
			if err != nil {
				fmt.Println("ReadFromLuaToContext data Error ", err.Error())
			}
		} else {
			context.InstanceVars[instanceId][index].Value = s[rows[index].Name]
		}
		//data, err := json.Marshal(rows[index].Value)

	}

	return nil
}

/*
//OPTIMIZE
func (context *InstanceContext) BindOutputVariablesFromLua(instanceId, pointId, userId int64) error {


	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("BindOutputVariablesFromLua", instanceId, processId, userId)
	}

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace("select v.code as \"name\",dt.code \"type\",'' \"value\" from bp_point_vars pv,bp_process_vars v,data_types dt where dt.id=v.data_type_id and v.id=pv.var_id and pv.is_output=1 and point_id=?"), pointId).QueryRows(&rows)
	if err != nil {
		return err
	}

	//i := 0

	//type mytype []map[string]interface{}
	//var v mytype

	context.Lua.Global("var")

	v, err := PullInterfaceTable(context.Lua, context.Lua.Top())
	//fmt.Println("parsed v = ",v )
	//err = json.Unmarshal([]byte(l.(string)),&v)
	if err != nil {
		fmt.Println("error on BindOutputVariablesFromLua", err)
		return err
	}

	for index, element := range rows {
		rows[index].Value = v[rows[index].Name]
		//fmt.Println("parsed ",rows[index].Name ," ",rows[index].Value )

		if rows[index].Type == "struct" {
			//data, err := json.Marshal(rows[index].Value)
			if err != nil {
				fmt.Println("data err", err.Error())
			}
			//fmt.Println("data=", rows[index].Name, data)
			//json.Unmarshal(data, &rows[index].Value)
			//fmt.Println("&rows[index].Value=", rows[index].Value)
		}
		err = context.SetProcessVarByInstance(instanceId, element.Name, rows[index].Value)
		if err != nil {
			return err
		}

	}



	//	err = WritePointJRNDTL(instanceId,userId,pointId,rows)
	//
	//	if err!=nil{
	//		log.Println(err)
	//		debug.PrintStack()
	//		return err
	//	}
	return nil
}
*/
func (context *InstanceContext) LuaBPMSStartProcess(L *lua.State) int {
	processCode, ok := L.ToString(1)

	if !ok {

		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString("")
		L.PushInteger(0)
		return 5
	}
	userId, ok := L.ToInteger(2)
	if !ok {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString("")
		L.PushInteger(0)
		return 5
	}
	var inputNameValue []NameValue

	arr, err := PullInterfaceTable(L, 3)

	if err != nil {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString("")
		L.PushInteger(0)
		return 5
	}

	for k, v := range arr {
		inputNameValue = append(inputNameValue, NameValue{Name: k, Value: v})
	}

	pid, err := context.GetProcessIdByProcessCode(processCode)
	context.Lua = lua.NewState()
	context.InstanceVars = make(map[int64][]NameValue, 0)
	context.InstanceTables = make(map[int64]string)
	lua.OpenLibraries(context.Lua)
	RegisterAPI(context.Lua, context.O)
	RegisterBPMLUaAPI(nil, context.Lua, context.O)

	outputNameValue, instance, task, _, err := context.CreateInstance(nil, pid, int64(userId), inputNameValue, 0)

	if err != nil {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 5

	}
	output := make(map[string]interface{})

	for _, v := range outputNameValue {
		output[v.Name] = v.Value
	}

	DeepPush(L, output)
	L.PushString(instance)
	L.PushString(task)
	L.PushString("")
	L.PushInteger(0)
	return 5
}

func (context *InstanceContext) LuaBPMSStartProcess2(L *lua.State) int {
	processCode, ok := L.ToString(1)

	if !ok {

		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString("")
		L.PushInteger(0)
		return 5
	}
	userId, ok := L.ToInteger(2)
	if !ok {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("error")
		L.PushString("error")
		L.PushInteger(0)
		return 5
	}
	var inputNameValue []NameValue

	arr, err := PullInterfaceTable(L, 3)

	if err != nil {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString("error")
		L.PushInteger(0)
		return 5
	}

	parent_id, ok := L.ToInteger(4)
	if !ok {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString("error")
		L.PushInteger(0)
		return 5
	}

	for k, v := range arr {
		inputNameValue = append(inputNameValue, NameValue{Name: k, Value: v})
	}

	pid, err := context.GetProcessIdByProcessCode(processCode)
	context.Lua = lua.NewState()
	context.InstanceVars = make(map[int64][]NameValue, 0)
	context.InstanceTables = make(map[int64]string)
	lua.OpenLibraries(context.Lua)
	RegisterAPI(context.Lua, context.O)
	RegisterBPMLUaAPI(nil, context.Lua, context.O)

	outputNameValue, instance, task, _, err := context.CreateInstance(nil, pid, int64(userId), inputNameValue, int64(parent_id))

	if err != nil {
		L.PushNil()
		L.PushInteger(0)
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 5

	}
	output := make(map[string]interface{})

	for _, v := range outputNameValue {
		output[v.Name] = v.Value
	}

	DeepPush(L, output)
	L.PushString(instance)
	L.PushString(task)
	L.PushString("")
	L.PushInteger(0)
	return 5
}

//DML С игнорированием прав доступов кроме валидаций
func (context *LuaContext) LuaDMLI(L *lua.State) int {

	action, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("LuaDMLI err bind 1")
		L.PushInteger(2)
		return 3
	}

	tableName, ok := L.ToString(2)
	if !ok {
		L.PushNil()
		L.PushString("LuaDMLI err bind 1")
		L.PushInteger(2)
		return 3
	}

	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("LuaDMLI err bind 3")
		L.PushInteger(2)
		return 3
	}

	t := DMLRequest{}
	i := DMLRequestItem{}
	i.TableName = tableName
	i.Action = action
	arrInterface, err := PullInterfaceTable(L, 4)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}
	i.Values = append(i.Values, arrInterface)

	t.Items = append(t.Items, i)

	s, err := DML(context.o, int64(userId), t, true)

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	s, err = TriggerSync(context.o, int64(userId), t, s)

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	lastId := 0
	for _, v := range s.Items {
		if tableName == v.TableName {
			lastId = int(v.LastInsertId)
		}
	}
	L.PushInteger(lastId)
	L.PushString("")
	L.PushInteger(0)

	return 3
}

func (context *LuaContext) LuaDML(L *lua.State) int {

	action, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("LuaDML err bind 1")
		L.PushInteger(2)
		return 3
	}

	tableName, ok := L.ToString(2)
	if !ok {
		L.PushNil()
		L.PushString("LuaDML err bind 1")
		L.PushInteger(2)
		return 3
	}

	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("LuaDML err bind 3")
		L.PushInteger(2)
		return 3
	}

	t := DMLRequest{}
	i := DMLRequestItem{}
	i.TableName = tableName
	i.Action = action
	arrInterface, err := PullInterfaceTable(L, 4)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}
	i.Values = append(i.Values, arrInterface)

	t.Items = append(t.Items, i)

	s, err := DML(context.o, int64(userId), t, false)

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	s, err = TriggerSync(context.o, int64(userId), t, s)

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	lastId := 0
	for _, v := range s.Items {
		if tableName == v.TableName {
			lastId = int(v.LastInsertId)
		}
	}
	L.PushInteger(lastId)
	L.PushString("")
	L.PushInteger(0)

	return 3
}

//Upsert с игнорированием ошибки на доступы кроме валидаций
func (context *LuaContext) LuaUpsertI(L *lua.State) int {

	upsertKeys, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("LuaUpsertI err bind 1")
		L.PushInteger(2)
		return 3
	}

	tableName, ok := L.ToString(2)
	if !ok {
		L.PushNil()
		L.PushString("LuaUpsertI err bind 1")
		L.PushInteger(2)
		return 3
	}

	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("LuaUpsert err bind 3")
		L.PushInteger(2)
		return 3
	}

	t := DMLRequest{}
	i := DMLRequestItem{}
	i.TableName = tableName
	i.Action = "upsert"
	i.UpsertAttrs = strings.Split(upsertKeys, ";")
	arrInterface, err := PullInterfaceTable(L, 4)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}
	i.Values = append(i.Values, arrInterface)

	t.Items = append(t.Items, i)

	s, err := DML(context.o, int64(userId), t, true)
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("LuaUpsert", t)
	}
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	lastId := 0
	for _, v := range s.Items {
		if tableName == v.TableName {
			lastId = int(v.LastInsertId)
		}
	}
	L.PushInteger(lastId)
	L.PushString("")
	L.PushInteger(0)

	return 3
}

func (context *LuaContext) LuaUpsert(L *lua.State) int {

	upsertKeys, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("LuaUpsert err bind 1")
		L.PushInteger(2)
		return 3
	}

	tableName, ok := L.ToString(2)
	if !ok {
		L.PushNil()
		L.PushString("LuaUpsert err bind 1")
		L.PushInteger(2)
		return 3
	}

	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("LuaUpsert err bind 3")
		L.PushInteger(2)
		return 3
	}

	t := DMLRequest{}
	i := DMLRequestItem{}
	i.TableName = tableName
	i.Action = "upsert"
	i.UpsertAttrs = strings.Split(upsertKeys, ";")
	arrInterface, err := PullInterfaceTable(L, 4)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}
	i.Values = append(i.Values, arrInterface)

	t.Items = append(t.Items, i)

	s, err := DML(context.o, int64(userId), t, false)
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("LuaUpsert", t)
	}
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	lastId := 0
	for _, v := range s.Items {
		if tableName == v.TableName {
			lastId = int(v.LastInsertId)
		}
	}
	L.PushInteger(lastId)
	L.PushString("")
	L.PushInteger(0)

	return 3
}

func (context *InstanceContext) LuaTerminateByInstanceId(L *lua.State) int {
	instanceId, ok := L.ToInteger(1)
	if !ok {
		L.PushString("Input Identified 1 Expected")
		L.PushInteger(1)
		return 2
	}

	v, err := PullTable(L, 2)

	if err != nil {
		L.PushString("Input Identified 2 Expected")
		L.PushInteger(2)
		return 2
	}

	context.ErrorJson = v

	log.Println("EEEEEEEE", context.ErrorJson)

	err2 := errors.New("Process Terminated")
	err = context.TerminateByInstanceId(int64(instanceId), err2)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(3)
		return 2
	}
	L.PushString("")
	L.PushInteger(0)
	return 2
}
func (context *LuaContext) LuaBPMSRunManualTask(L *lua.State) int {

	task, ok := L.ToString(1)
	if !ok {
		L.PushString("")
		L.PushString("LuaBPMSRunManualTask err bind task (1 param)")
		L.PushInteger(2)
		return 3
	}

	userId, ok := L.ToInteger(2)
	if !ok {
		L.PushString("")
		L.PushString("LuaBPMSRunManualTask err bind userId (2 param)")
		L.PushInteger(2)
		return 3
	}

	var input []NameValue

	arr, err := PullInterfaceTable(L, 3)

	if err != nil {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	}
	for k, v := range arr {
		input = append(input, NameValue{Name: k, Value: v})
	}

	instanceContext := InstanceContext{O: context.o}
	instanceContext.InstanceVars = make(map[int64][]NameValue)
	instanceContext.InstanceTables = make(map[int64]string, 0)

	instanceContext.Lua = lua.NewState()
	lua.OpenLibraries(instanceContext.Lua)
	RegisterAPI(instanceContext.Lua, context.o)
	RegisterBPMLUaAPI(context.req, instanceContext.Lua, context.o)

	pointId, err := instanceContext.GetPointByTask(task)

	if err != nil {
		context.o.Rollback()
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	_, err = instanceContext.CheckRqOutputIntPointVars(pointId, input)

	if err != nil {
		context.o.Rollback()
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(4)
		return 3
	}

	newTask, err := instanceContext.ManualExecInstanceByTask(context.req, task, int64(userId), input)
	if err != nil {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	} else {
		L.PushString(newTask)
		L.PushString("")
		L.PushInteger(0)
		return 3
	}

	return 3
}

func RegisterBPMLUaAPI(req *http.Request, l *lua.State, o orm.Ormer) {

	luaContext := LuaContext{}
	luaContext.o = o
	InstanceContext := InstanceContext{}
	InstanceContext.O = o
	l.Register("BPMSRunManualTask", luaContext.LuaBPMSRunManualTask)
	l.Register("TerminateByInstanceId", InstanceContext.LuaTerminateByInstanceId)

	l.Register("BPMSStartProcess", InstanceContext.LuaBPMSStartProcess)
	l.Register("BPMSStartProcess2", InstanceContext.LuaBPMSStartProcess2)
	l.Register("DML", luaContext.LuaDML)
	l.Register("DMLI", luaContext.LuaDMLI)
	l.Register("Upsert", luaContext.LuaUpsert)
	l.Register("UpsertI", luaContext.LuaUpsertI)
	//stateorm.Open(l, o)
	//pkg.Open(l)

}

func (context *InstanceContext) runLuaScriptTask(o orm.Ormer, req *http.Request, instanceId, point, userId int64, script string) error {

	//log.Println("Startttttt")
	//log.Println(point)
	//****************************
	//log.Println("Script Body "+row.Script)
	//l.PushString("KRUTO GOI!!!!")
	//l.SetGlobal("guid")

	//	jrnId,err := CreatePointJRN(instanceId,userId)
	//	if err != nil {
	//		log.Println("error CreatePointJRN  "+err.Error())
	//		debug.PrintStack()
	//		return err
	//	}

	//err := context.BindInputVariablesToLua3(req, instanceId, point, userId)
	//err := context.BindInputVariablesToLua2(req, instanceId, point, userId)
	//err := context.BindInputVariablesToLua(instanceId, point, userId, 0, l)

	//(instanceId, pointId, userId, jrnId int64, state *lua.State)
	//err = BindInputVariablesToLua(instanceId,point,userId,jrnId,l)
	//if err != nil {
	//	log.Println("runLuaScriptTask error bind variables  " + err.Error())
	//	debug.PrintStack()
	//	//o.Rollback()
	//	return err
	//}

	if err := lua.DoString(context.Lua, script); err != nil {
		log.Println("error lua  " + err.Error())
		log.Println("script = " + script)
		debug.PrintStack()
		//o.Rollback()
		utils.ErrorWriteUser("BPMLuaScriptError2", "nothing", userId, err)
		//l = nil
		return err
	}
	//	l.Global("aaa")
	//	x,_:=l.ToString(1)
	//	log.Println("URA!!! "+x+" VATA EMES")
	//****************************

	//err = context.BindOutputVariablesFromLua(instanceId, point, userId)

	//if err != nil {
	//	log.Println("error set output var  " + err.Error())
	//	debug.PrintStack()
	//	return err
	//}
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("LUA SCRIPT POINT DONE")
	}
	return nil
}

func (context *InstanceContext) LuaExpression(expression string, instanceId int64, pointId int64, userId int64) (string, error) {

	resStr := "__result"

	/*
		//OPTIMIZE
		err := context.BindLocalVariablesToByPoint(nil, instanceId, pointId, userId)
		if err != nil {
			log.Println("LuaExpression error bind variables  " + err.Error())
			debug.PrintStack()
			return "", err
		}*/

	if err := lua.DoString(context.Lua, resStr+" = "+expression); err != nil {
		log.Println("LuaExpression error lua  " + err.Error())
		log.Println("LuaExpression script = " + resStr + " = " + expression)
		debug.PrintStack()
		utils.ErrorWriteUser("BPMLuaConditionError", "nothing", userId, err)
		return "", errors.New("Lua Condition error of \"" + resStr + " = " + expression + "\"")
	}

	context.Lua.Global(resStr)
	result, ok := context.Lua.ToString(context.Lua.Top())
	if !ok {
		return "", errors.New("Error on Bind String BPMLuaConditionError")
	}

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("LUA Expression DONE " + resStr + " = " + expression)
		log.Println(result)
	}
	return result, nil
}

func (context *InstanceContext) LuaCondition(o orm.Ormer, req *http.Request, instanceId int64, point, userId int64, flow int64, condition string) (bool, error) {

	resStr := "__result"

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("Start Lua Condition")
	}

	if condition == "" {
		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("LuaCondition Skip", flow)
		}
		return true, nil
	}

	/*
		//OPTIMIZE
		err := context.BindLocalVariablesToByPoint(req, instanceId, point, userId)
		if err != nil {
			log.Println("luaCondition error bind variables  " + err.Error())
			debug.PrintStack()
			return false, err
		}*/

	if err := lua.DoString(context.Lua, resStr+" = "+condition); err != nil {
		log.Println("luaCondition error lua  " + err.Error())
		log.Println("luaCondition script = " + resStr + " = " + condition)
		debug.PrintStack()
		utils.ErrorWriteUser("BPMLuaConditionError", "nothing", userId, err)
		return false, errors.New("Lua Condition error of \"" + resStr + " = " + condition + "\"")
	}

	context.Lua.Global(resStr)
	result := context.Lua.ToBoolean(context.Lua.Top())
	/*log.Println("result", result, condition)

	for i := 0; i < 10; i++ {
		log.Println("result", i, context.Lua.ToValue(i))
	}*/

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("LUA CONDITION POINT DONE " + resStr + " = " + condition)
		log.Println(result)
	}
	return result, nil
}
