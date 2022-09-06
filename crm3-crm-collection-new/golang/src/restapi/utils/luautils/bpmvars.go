package luautils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	jsoniter "github.com/json-iterator/go"
)

func (context *InstanceContext) BindOutputVariablesFromSubProcess(childInstanceId int64) error {

	type instSrcDstValue struct {
		SrcInstanceId int64       `json:"src_instance_id"`
		Src           string      `json:"src"`
		Dst           string      `json:"dst"`
		Value         interface{} `json:"value"`
		Len           int         `json:"len"`
	}

	rows := []instSrcDstValue{}
	values := []NameValue{}

	_, err := context.O.Raw(utils.DbBindReplace(`select i.parent_id as "src_instance_id", v_src.code as "src",
	 v.code as "dst",'' as "value",v.len as "len"
  from
  bp_instances i
  join bp_instances ip on ip.id=i.parent_id
  join bp_points po on po.process_id = ip.process_id
  join bp_point_datamaps pdm on pdm.point_id = po.id
  join bp_process_vars v_src on v_src.id = pdm.var_src_out
  join bp_process_vars v on v.id = pdm.var_dst and v.is_output = 1
  join data_types dt on dt.id =v.data_type_id
  where i.id=? and po.is_active=1`), childInstanceId).QueryRows(&rows)
	if err != nil {
		//Need new code
		err = errors.New("ERROR E-BPM-00019. BindOutputVariablesFromSubProcess" + err.Error())
		log.Println(err)
		return err
	}

	for _, element := range rows {
		for _, child := range context.InstanceVars[childInstanceId] {
			if element.Dst == child.Name {
				values = append(values, NameValue{Name: element.Src, Value: child.Value, Len: element.Len})
			}
		}
		if os.Getenv("CRM_VERBOSE_SQL") == "1" {
			log.Println("BindOutputVariablesFromSubProcess value")
		}
		if err != nil {
			err = errors.New("ERROR E-BPM-000020. BindOutputVariablesFromSubProcess" + err.Error())
			return err
		}
	}

	if len(rows) > 0 {
		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		   log.Println("BindOutputVariablesFromSubProcess", values)
		}
		context.SetVariablesToContext(rows[0].SrcInstanceId, values)
	}
	if err != nil {
		return err
	}
	return nil
}

//Установка переменных из переменной в контекст
//Все переменные должны быть проинициализированы
func (context *InstanceContext) SetVariablesToContext(instanceId int64, input []NameValue) error {
	
	
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
	   log.Println("SetVariablesToContext before", context.InstanceVars[instanceId])
	}	
	
	for _, v1 := range input {
		for k2, v2 := range context.InstanceVars[instanceId] {
			if v1.Name == v2.Name {
				context.InstanceVars[instanceId][k2].Value = v1.Value
				context.InstanceVars[instanceId][k2].Len = v1.Len
				
				if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				   log.Println("SetVariablesToContext", instanceId , v1.Name, "--", v1.Value)
				}
		
				
				
				
			}
		}
	}
	//log.Println("after", context.InstanceVars[instanceId])
	return nil
}

func (context *InstanceContext) ReadAllVariablesFromDB(instanceId, processId int64, processTable string, userId int64) error {

	parent_instance_id := int64(0)
	parent_process_id := int64(0)
	parent_point_id := int64(0)
	err := context.O.Raw(utils.DbBindReplace("select i.parent_id,i.parent_point_id,pi.process_id  from bp_instances i join bp_instances pi on pi.id=i.parent_id where i.id=?"), instanceId).QueryRow(&parent_instance_id, &parent_point_id, &parent_process_id)
	//log.Println("ReadAllVariablesFromDB", 1)
	if parent_instance_id != 0 && parent_point_id != 0 && context.InstanceTables[parent_instance_id] == "" {
		//log.Println("ReadAllVariablesFromDB", 2)
		parentProcessTable, err := context.GetProcessTableByProcessId(parent_process_id)

		if err != nil {
			log.Println("ReadAllVariablesFromDB processTable " + err.Error())
			log.Println(instanceId)
			log.Println(processTable)
			log.Println(userId)
			return err
		}

		err = context.ReadContextVariableToLua(parent_instance_id, parent_process_id, userId)
		if err != nil {
			err = errors.New("ERROR E-BPM-00002. ReadAllVariablesFromDB 1 " + err.Error())
			log.Println(err)
			return err
		}

		err = context.ReadAllVariablesFromDB(parent_instance_id, parent_process_id, parentProcessTable, userId)
		if err != nil {
			err = errors.New("ERROR E-BPM-00002. ReadAllVariablesFromDB 2 " + err.Error())
			log.Println(err)
			return err
		}

	}

	context.InstanceTables[instanceId] = processTable

	arr := []orm.Params{}
	_, err = context.O.Raw(utils.DbBindReplace("select * from "+processTable+" where id$=?"), instanceId).Values(&arr)
	if err != nil {
		err = errors.New("ERROR E-BPM-00002. ReadAllVariablesFromDB" + err.Error())
		log.Println(err)
		return err
	}

	tempVars := []NameValue{}
	_, err = context.O.Raw(utils.DbBindReplace(`select
	v.code "name", dt.code "type",'' "value", v.len "len"
from
	bp_processes p
	join bp_process_vars v on v.process_id = p.id
	join data_types dt on dt.id=v.data_type_id
where p.id=?`), processId).QueryRows(&tempVars)
	if err != nil {
		err = errors.New("ERROR E-BPM-00002. ReadAllVariablesFromDB" + err.Error())
		log.Println(err)
		return err
	}

	for k, v := range tempVars {
		if len(arr) > 0 {
			if arr[0][v.Name] == nil {
				//Специфика присваивания переменных по умолчанию
				//Если переменная не задана, то будет пустая строка ""
				tempVars[k].Value = ""
			} else {
				tempVars[k].Value = arr[0][v.Name]
			}

			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("DBREADED", v.Name, tempVars[k].Value, tempVars[k].Len)
			}

		} else {
			//Специфика присваивания переменных по умолчанию
			//Если переменная не задана, то будет пустая строка ""
			tempVars[k].Value = ""
		}
		//log.Println("tempvar", v.Name, v.Value)
		//context.InstanceVars[instanceId] = append(context.InstanceVars[instanceId], NameValue{Name: v, Value: arr[k]})
	}
	context.InstanceVars[instanceId] = tempVars
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("DBREADED", instanceId, context.InstanceVars[instanceId])
	}
	return err
}

func (context *InstanceContext) BindInputVariablesToSubProcess(instanceId, pointId, userId int64) ([]NameValue, error) {

	type srcDstValue struct {
		Src    string      `json:"src"`
		Dst    string      `json:"dst"`
		Value  interface{} `json:"value"`
		IsVar  int64       `json:"is_var"`
		IsExpr int64       `json:"is_expr"`
		Expr   string      `json:"expr"`
	}

	//
	rows := []srcDstValue{}
	values := []NameValue{}

	_, err := context.O.Raw(utils.DbBindReplace(`select v.code as "src", v2.code as "dst"  ,'' "value",
	pdm.is_var as "is_var",pdm.is_expr as "is_expr",pdm.expr as "expr" from
	bp_point_datamaps pdm
left join bp_process_vars v on v.id=pdm.var_src and pdm.is_var = 1
left join bp_process_vars v2 on v2.is_input = 1  and v2.id = pdm.var_dst
where pdm.point_id=?`), pointId).QueryRows(&rows)

	/*_, err := context.O.Raw(utils.DbBindReplace(`select v.code src, v2.code as dst  ,'' "value",pdm.is_var,pdm.is_expr,pdm.expr
	 from bp_process_vars v,bp_points po,
	  bp_process_vars v2,bp_point_datamaps pdm,
	  data_types dt,bp_instances i

	where v.data_type_id=dt.id and
	      i.process_id=v.process_id and i.id=?
	      and pdm.point_id=po.id
	      and po.subprocess_id = v2.process_id and v2.id = pdm.var_dst
	      and po.id = ? and po.process_id=i.process_id
	      and v2.is_input = 1
	      and po.is_active=1
	      and (v.id=pdm.var_src and pdm.is_var = 1 or pdm.is_expr = 1)`), instanceId, pointId).QueryRows(&rows)*/
	if err != nil {
		//Need new code
		err = errors.New("ERROR E-BPM-00021. BindInputVariablesToSubProcess" + err.Error())
		log.Println(err)
		return values, err
	}
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("SubProcess Vars start")
	}

	for _, element := range rows {

		if element.IsVar == 1 {
			//value, _, err := context.(instanceId, element.Src)

			for _, v := range context.InstanceVars[instanceId] {
				if v.Name == element.Src {
					values = append(values, NameValue{Name: element.Dst, Value: v.Value})
				}
			}

			if err != nil {
				err = errors.New("ERROR E-BPM-00022. BindInputVariablesToSubProcess" + err.Error())
				return values, err
			}
		} else if element.IsExpr == 1 {
			value, err := context.LuaExpression(element.Expr, instanceId, pointId, userId)
			//log.Println("VALUE", element.Expr, value)
			values = append(values, NameValue{Name: element.Dst, Value: value})
			if err != nil {
				log.Println("ERROR E-BPM-00022. BindInputVariablesToSubProcess 2 ", element, err)
				return values, err
			}
		}

		//rows[index].Value ,_, err = context.GetProcessVarByInstance(instanceId,element.)
		if err != nil {
			log.Println("ERROR E-BPM-00022. BindInputVariablesToSubProcess 3", err)
			return values, err
		}
	}
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("SubProcess Vars finish", rows)
	}
	return values, nil
}

func (context *InstanceContext) SaveVariables() error {

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("SaveVariables variables",context.InstanceVars)
	}

	for k, _ := range context.InstanceVars {
		var values []interface{}
		sqlU := "update " + context.InstanceTables[k] + " set "
		sqlI := "insert into " + context.InstanceTables[k] + "  "
		sqlIF := ""
		sqlIV := ""
		for kk, v := range context.InstanceVars[k] {

			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("VERBOSE: save variable full", v.Value, v.Type, kk)
			}

			if v.Value != nil {
				if v.Type == "struct" {
					b, _ := json.Marshal(v.Value)
					if utils.GetDbDriverType() == orm.DROracle {
						values = append(values, b)
					} else {
						values = append(values, string(b))
					}
					/*} else if utils.GetDbDriverType() == orm.DROracle && (v.Type == "text" || v.Type == "longtext") {
					values = append(values, []byte(fmt.Sprintf("%v", v.Value)))
					*/
				} else {
					values = append(values, fmt.Sprintf("%v", v.Value))
				}
			} else {
				values = append(values, "")
			}

			if utils.GetDbDriverType() == orm.DROracle {

				if v.Type == "varchar" && v.Len > 0 {
					sqlU = sqlU + v.Name + " = substr(?,1," + strconv.Itoa(v.Len) + ") "
					sqlIV = sqlIV + "  substr(?,1," + strconv.Itoa(v.Len) + "), "
				} else {
					sqlU = sqlU + v.Name + " = ? "
					sqlIV = sqlIV + " ?, "
				}
				sqlIF = sqlIF + " " + v.Name + " , "

			} else {
				sqlU = sqlU + v.Name + " = nullif(?,'') "
				sqlIF = sqlIF + " " + v.Name + " , "
				sqlIV = sqlIV + " nullif(?,''), "
			}
			if kk != len(context.InstanceVars[k])-1 {
				sqlU = sqlU + ", "
			}

		}

		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("VERBOSE: save variable before", values)
		}

		values = append(values, fmt.Sprintf("%v", k))

		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("VERBOSE: save variable after", values)
		}

		sqlU = sqlU + " where id$ = ?"
		sqlI = sqlI + " (" + sqlIF + " id$) values (" + sqlIV + " ?)"
		//log.Println("save variable update", sqlU)
		//log.Println("save variable insert", sqlI)

		pr, err := context.O.Raw(utils.DbBindReplace(sqlI)).Prepare()
		defer pr.Close()
		if err == nil {

			_, err = pr.Exec(values...)
		}
		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("VERBOSE: save variable", sqlI, values)
		}
		if err != nil {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				if !strings.Contains(err.Error(), "ORA-00001") {
					log.Println("Error on SaveVariables. Try Update", k, sqlI, err)
				}
			}
			//_, err = context.O.Raw(utils.DbBindReplace(sqlU), values).Exec()
			pr, err = context.O.Raw(utils.DbBindReplace(sqlU)).Prepare()
			defer pr.Close()
			_, err = pr.Exec(values...)

			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("VERBOSE: save variable", sqlU, values)
			}
			if err != nil {
				log.Println("Error on SaveVariables Update", sqlU, err)
				return err
			}
		}
		//log.Println("save variable insert result", err)
	}

	return nil

}

//Требует оптимизации. Надо сделать копию искать по pointId
func (context *InstanceContext) OnlyOutPutVars(values []NameValue, PointId int64) ([]NameValue, error) {

	rows := []NameValue{}
	outputValues := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(`select prv.code "name",prv.len "len"
  from
  bp_point_vars pov
  join bp_process_vars prv on prv.id=pov.var_id
  where
  pov.is_output=1 and
  pov.point_id=?`), PointId).QueryRows(&rows)
	if err != nil {
		err = errors.New("ERROR E-BPM-00002.1 OnlyOutPutVars" + err.Error())
		log.Println(err)
		return outputValues, err
	}

	for _, element := range rows {

		for _, value := range values {
			if value.Name == element.Name {
				value.Len = element.Len
				outputValues = append(outputValues, value)
			}
		}
	}
	return outputValues, nil
}

//Требует оптимизации. Надо сделать копию искать по pointId
func (context *InstanceContext) BindGlobalOuputVariablesToVar(instanceId, ProcessId int64) ([]NameValue, error) {

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace(`select v.code "name", dt.code "type",'' "value", v.len "len"
	from
	bp_processes pr
	join bp_process_vars v on v.process_id=pr.id
	join data_types dt on v.data_type_id=dt.id
	where v.is_output=1 and pr.id=?`), ProcessId).QueryRows(&rows)
	if err != nil {
		err = errors.New("ERROR E-BPM-00002. BindGlobalOuputVariablesToVar" + err.Error())
		log.Println(err)
		return rows, err
	}
	//log.Println("XXAA", context.InstanceVars[instanceId])
	for index, element := range rows {

		for _, vars := range context.InstanceVars[instanceId] {
			if vars.Name == element.Name {
				rows[index].Value = vars.Value
				rows[index].Len = vars.Len
			}
		}

		if err != nil {
			log.Println("Error on BindGlobalOuputVariablesToVar ", err)
			return rows, err
		}
	}
	return rows, nil
}

func (context *InstanceContext) BindVariablesByFormToVar(instanceId, formId int64) ([]NameValue, error) {

	rows := []NameValue{}
	_, err := context.O.Raw(utils.DbBindReplace("select v.code as \"name\",'' \"value\",dt.code as \"type\",v.len as \"len\" from bp_process_form_vars pv,bp_process_vars v,data_types dt  where v.id=pv.var_id and pv.form_id=? and v.data_type_id=dt.id"), formId).QueryRows(&rows)
	if err != nil {
		err = errors.New("ERROR E-BPM-00003. Error on BindVariablesByFormToVar " + err.Error())
		log.Println(err.Error())
		return rows, err
	}
	//log.Println("test123")
	//log.Println(formId)
	//log.Println(rows)
	for index, element := range rows {
		rows[index].Value, _, err = context.GetProcessVarByInstance(instanceId, element.Name)
		rows[index].Len = element.Len
		if err != nil {
			err = errors.New("ERROR E-BPM-00004. Error on BindVariablesByFormToVar.GetProcessVarByInstance " + err.Error())
			log.Println(err.Error())
			return rows, err
		}
	}
	return rows, nil
}

func (context *InstanceContext) BindInputGlobalVariablesToVar(instanceId, pointId int64) ([]NameValue, error) {

	rows := []NameValue{}
	sql := "select v.code as \"name\",'' as \"value\",dt.code as \"type\",v.len as \"len\" from bp_point_vars pv,bp_process_vars v,data_types dt  where v.id=pv.var_id and pv.is_input=1 and point_id=? and v.data_type_id=dt.id"
	_, err := context.O.Raw(utils.DbBindReplace(sql), pointId).QueryRows(&rows)
	if err != nil {
		err = errors.New("ERROR E-BPM-00005. Error on BindInputGlobalVariablesToVar " + sql + " " + err.Error())
		log.Println(err, instanceId, pointId)
		return rows, err
	}
	for index, element := range rows {

		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("BindInputGlobalVariablesToVar", element, pointId)
		}
		rows[index].Value, _, err = context.GetProcessVarByInstance(instanceId, element.Name)
		rows[index].Len = element.Len
		if err != nil {

			err = errors.New("ERROR E-BPM-00006. Error on BindInputGlobalVariablesToVar.GetProcessVarByInstance " +
				err.Error())
			//log.Println(element)
			return rows, err
		}
	}
	return rows, nil
}

func (context *InstanceContext) SetGlobalVarsByInstance(instanceId int64) error {

	/*for _, element := range arr {
		err := context.SetProcessVarByInstance(instanceId, element.Name, element.Value)
		if err != nil {
			err = errors.New("ERROR E-BPM-00006. Error on SetGlobalVarsByInstance.SetProcessVarByInstance " + err.Error())
			log.Println(err.Error())
			return err
		}
	}*/
	return nil

}

//func CreatePointJRN(instanceId,userId int64) (int64,error){
//
//	point,err := getPointByInstance(instanceId)
//	if err!=nil{
//		return 0,err
//	}
//	lInsertId := int64(0)
//
//
//	o := orm.NewOrm()
//	o.Using("default")
//	if userId == 0 {
//		rs,err := o.Raw("insert into bp_point_jrns (point_id,instance_id,exec_at) values (?,?,NOW())", point, instanceId).Exec()
//		if err!=nil{
//			debug.PrintStack()
//			log.Println(err)
//		}
//		lInsertId,err = rs.LastInsertId()
//		if err!=nil{
//			debug.PrintStack()
//			log.Println(err)
//		}else {
//			return lInsertId, err
//		}
//
//
//	}else{
//		rs,err := o.Raw("insert into bp_point_jrns (point_id,instance_id,exec_at,user_id) values (?,?,NOW(),?)", point, instanceId,userId).Exec()
//		if err!=nil{
//			debug.PrintStack()
//			log.Println(err)
//		}
//		lInsertId,err = rs.LastInsertId()
//		if err!=nil{
//			debug.PrintStack()
//			log.Println(err)
//		}else {
//			return lInsertId, err
//		}
//	}
//
//	return 0,nil
//}

//
//func WritePointJRNDTL(instanceId,userId,point,lInsertId int64,input []NameValue) error{
//
//
//	o := orm.NewOrm()
//	o.Using("default")
//
//	//Журналирование переменных
//	for _,element := range input{
//
//		log.Println("@@@@")
//		log.Println(lInsertId)
//		log.Println(point)
//		log.Println(element.Name)
//		log.Println(element.Value)
//		log.Println("####")
//
//
//		isJrn := 0;
//		err := o.Raw(`select count(1) cnt from bp_point_vars pv, bp_process_vars v
//		where pv.var_id=v.id and pv.point_id = ? and v.code=? and pv.is_jrn=1`,point, element.Name).QueryRow(&isJrn);
//
//		if err != nil {
//			log.Println(err)
//			debug.PrintStack()
//			return err
//		}
//
//		log.Println("NEED JRN element.Name =")
//		log.Println(isJrn)
//
//		if isJrn != 0 {
//
//			if  element.Value != nil && reflect.TypeOf(element.Value).Kind() != reflect.Map  {
//				_, err := o.Raw(`insert into bp_point_jrn_vars (jrn_id,point_var_id,value) values
//				 (?,
//				 (select pv.id from bp_point_vars pv, bp_process_vars v
//				   where pv.var_id=v.id and pv.point_id = ? and v.code=?
//				 ),?)`, lInsertId, point, element.Name, element.Value.(string)).Exec()
//				if err != nil {
//					log.Println(err)
//					debug.PrintStack()
//					return err
//				}
//			} else {
//				jsonData, _ := json.Marshal(element.Value)
//				_, err := o.Raw(`insert into bp_point_jrn_vars (jrn_id,point_var_id,value) values
//				 (?,
//				 (select pv.id from bp_point_vars pv, bp_process_vars v
//				   where pv.var_id=v.id and pv.point_id = ? and v.code=?
//				 ),?)`, lInsertId, point, element.Name, string(jsonData)).Exec()
//				if err != nil {
//					log.Println(err)
//					debug.PrintStack()
//					return err
//				}
//			}
//
//		}
//	}
//	//Конец Журналирования переменных
//	return nil;
//
//}
//func WritePointJRN(instanceId,userId int64,input []NameValue) error{
//
//	point,err := getPointByInstance(instanceId)
//	if err!=nil{
//		return err
//	}
//
//	lInsertId,err := CreatePointJRN(instanceId,userId)
//
//	if err!=nil{
//		return err
//	}
//
//	WritePointJRNDTL(instanceId,userId,point,lInsertId,input)
//	return nil
//}

func (context *InstanceContext) SetProcessVarByInstance(instanceId int64, varCode string, varValue interface{}) error {

	//	processId,err := GetProcessIdByInstance(instanceId);
	//	if err!=nil{
	//		return err
	//	}
	processCode, err := context.GetProcessCodeByInstance(instanceId)

	if err != nil {
		err = errors.New("ERROR E-BPM-00008. Error on SetGlobalVarsByInstance.SetProcessVarByInstance " + err.Error())
		log.Println(err.Error())
		return err

	}

	//context.O.Raw("LOCK TABLE  i$"+processCode+" IN EXCLUSIVE MODE").Exec()
	//defer context.O.Raw("RollBack")
	//context.O.Raw(utils.DbBindReplace("select * from bp_instances z2 where id=? limit 1 for update"), instanceId).Exec()
	//LOCK TABLE  i$"+processCode+" IN SHARE MODE

	if varValue != nil && reflect.TypeOf(varValue).Kind() == reflect.Map {

		json := jsoniter.ConfigCompatibleWithStandardLibrary
		varValueMap, _ := json.Marshal(&varValue)
		varValue = string(varValueMap)
	}

	//log.Println("varCode="+varCode)
	//log.Println("varValue="+varValue.(string))
	//log.Println("instanceId="+strconv.Itoa(int(instanceId)))
	//log.Println("processCode="+processCode)

	isArray := 0
	dataType := 0
	err = context.O.Raw(utils.DbBindReplace("select  v.is_array,v.data_type_id from bp_instances i,bp_process_vars v where i.id=? and v.process_id=i.process_id and v.code=?"), instanceId, varCode).QueryRow(&isArray, &dataType)
	if err != nil {
		err = errors.New("ERROR E-BPM-00001. Process Variable not found. on getType on SetProcessVarByInstance varCode = " + varCode + " instanceID=" + strconv.Itoa(int(instanceId)) + err.Error())
		log.Println(err.Error())
		return err
	}

	//log.Println("reflect.ValueOf(varValue).Kind()")
	//log.Println(reflect.ValueOf(varValue).Kind())
	if isArray == 1 && reflect.TypeOf(varValue).Kind() != reflect.String {
		//log.Println(varCode+".varvalue_bef")
		//log.Println(varValue)
		varValueMap, _ := json.Marshal(varValue)
		varValue = string(varValueMap)
		//log.Println(varCode+".varvalue_after")
		//log.Println(varValue)
	}

	//context.O.Raw("LOCK TABLE i$"+processCode+" WRITE").Exec()

	sql := "update  i$" + processCode + " set " + varCode + "=NULLIF(?,'')  where id$=?"

	if utils.GetDbDriverType() == orm.DROracle {
		sql = "update  i$" + processCode + " set " + varCode + "=?  where id$=?"
	}

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("ttt", instanceId)
	}
	_, err = context.O.Raw(utils.DbBindReplace(sql), varValue, instanceId).Exec()
	if err != nil {
		err = errors.New("ERROR E-BPM-00009. Error on SetProcessVarByInstance ----" + sql + "---- " + err.Error())
		log.Println("E-BPM-00009", instanceId, err.Error())
		return err
	}

	cnt := 0

	//_, err = context.O.Raw(utils.DbBindReplace("select * from bp_instances z1 limit 1 for update")).Exec()

	err = context.O.Raw(utils.DbBindReplace("select count(1) from i$"+processCode+" where id$=?"), instanceId).QueryRow(&cnt)

	//cnt = 1
	//
	//data :=0

	if err != nil {
		err = errors.New("ERROR E-BPM-00010. Error on select count(1) " + processCode + " " + err.Error())
		log.Println(err.Error())
		return err
	}
	if cnt == 0 {

		sql := "insert into  i$" + processCode + " (" + varCode + ",id$) values (NULLIF(?,''),?)"

		if utils.GetDbDriverType() == orm.DROracle {
			sql = "insert into  i$" + processCode + " (" + varCode + ",id$) values (?,?)"
		}

		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("instance insert...", instanceId)
			log.Println("inseeert", sql, varValue)
		}

		_, err := context.O.Raw(utils.DbBindReplace(sql), varValue, instanceId).Exec()

		//context.O.Raw("UNLOCK TABLES").Exec()

		if err != nil {
			err = errors.New("ERROR E-BPM-00011. Error on insert into i$" + processCode + " " + err.Error())
			log.Println(err.Error())
			return err
		}

		//context.O.Commit()
		//context.O.Begin()

	}

	return nil

}

func (context *InstanceContext) GetProcessVarByInstance(instanceId int64, varCode string) (interface{}, bool, error) {

	var varValue string

	//	processId,err := GetProcessIdByInstance(instanceId);
	//	if err!=nil{
	//		return varValue,false,err
	//	}

	processCode, err := context.GetProcessCodeByInstance(instanceId)
	if err != nil {
		log.Println("Error GetProcessCodeByInstance", instanceId, err)
		return varValue, false, err
	}

	err = context.O.Raw(utils.DbBindReplace("select "+varCode+" from i$"+processCode+" where id$=?"), instanceId).QueryRow(&varValue)
	if os.Getenv("CRM_VERBOSE_SQL") == "1" {
		log.Println("GetProcessVarByInstance 1 ", processCode, varValue)
	}

	isArray := 0
	dbDataType := "string"
	dataType := ""
	err = context.O.Raw(utils.DbBindReplace("select  v.is_array,dt.db_data_type,dt.code from bp_instances i,bp_process_vars v,data_types dt where i.id=? and v.process_id=i.process_id and v.code=? and v.data_type_id=dt.id"), instanceId, varCode).QueryRow(&isArray, &dbDataType, &dataType)
	if err != nil {
		err = errors.New("ERROR E-BPM-00012. Error on getType on SetProcessVarByInstance varcode=" + varCode + " " + err.Error())
		log.Println(err.Error())
		return "", false, err
	}

	if isArray == 1 && varValue != "" {
		//log.Println("varValue +++ "+varValue)

		if os.Getenv("CRM_VERBOSE_SQL") == "1" {
			log.Println("GetProcessVarByInstance 2", varValue, processCode, varValue)
		}

		var varValue2 []interface{}
		err = json.Unmarshal([]byte(varValue), &varValue2)

		if err != nil {
			err = errors.New("ERROR E-BPM-00013. Error on getType isArray json on SetProcessVarByInstance varValue=" + varValue + " " + err.Error())
			log.Println(err.Error())
			return "", false, err
		} else {
			return varValue2, isArray == 1, nil
		}
		//varValue = varValue2
	} else if dataType == "struct" {
		//log.Println("varValue +++ "+varValue)
		var varValue2 interface{}
		if varValue == "" && isArray == 1 {
			varValue = "[]"
		} else if varValue == "" {
			varValue = "{}"
		}
		err = json.Unmarshal([]byte(string(varValue)), &varValue2)

		if err != nil {
			err = errors.New("ERROR E-BPM-00014. Error on getType struct json on SetProcessVarByInstance varValue=" + varValue + " " + err.Error())
			log.Println(err.Error())
			return "", false, err
		} else {
			return varValue2, isArray == 0, nil
		}
	} else if dbDataType == "integer" && varValue != "" {
		//fmt.Println("parsing"," ",varCode," ",varValue)
		i, err := strconv.Atoi(varValue)
		if err != nil {
			err = errors.New("ERROR E-BPM-00015. Error parse code varcode=" + varCode + ",dbDataType=" + dbDataType + " " + err.Error())
			log.Println(err.Error())
			return 0, false, err
		} else {
			return i, false, nil
		}
	} else if dbDataType == "double" && varValue != "" {
		f, err := strconv.ParseFloat(varValue, 64)
		if err != nil {
			err = errors.New("ERROR E-BPM-00016. Error parse code varcode=" + varCode + ",dbDataType=" + dbDataType + " " + err.Error())
			log.Println(err.Error())
			return 0, false, err
		} else {
			return f, false, nil
		}
	}

	if os.Getenv("CRM_VERBOSE_SQL") == "1" {
		log.Println("GetProcessVarByInstance 3", varValue, processCode, varValue)
	}
	return varValue, false, nil

	//return varValueInt,nil

}

func (context *InstanceContext) GetProcessTableByProcessId(processId int64) (string, error) {

	tableName := ""
	err := context.O.Raw(utils.DbBindReplace("select code from bp_processes where id=?"), processId).QueryRow(&tableName)
	return "i$" + tableName, err
}

func (context *InstanceContext) GetProcessIdByProcessCode(processCode string) (int64, error) {

	processId := int64(0)
	err := context.O.Raw(utils.DbBindReplace("select id from bp_processes where code=?"), processCode).QueryRow(&processId)
	return processId, err
}

func (context *InstanceContext) FindInstanceIdByUniqueVar(processCode, name string, value interface{}) (int64, error) {

	instanceId := int64(0)
	err := context.O.Raw(utils.DbBindReplace("select id$ from i$"+processCode+" where "+name+"=?"), value).QueryRow(&instanceId)
	if err != nil {
		return 0, utils.AddLogError(err, "E-BPM-00017", "Error on FindInstanceIdByUniqueVar processCode %v name %v", processCode, name)
	} else {
		return instanceId, err
	}

}
