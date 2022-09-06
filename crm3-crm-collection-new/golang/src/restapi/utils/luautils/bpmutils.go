package luautils

import (
	"errors"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	lua "github.com/Shopify/go-lua"

	"fmt"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
)

type InstanceContext struct {
	O               orm.Ormer
	Req             *http.Request
	Lua             *lua.State
	ErrorJson       interface{}
	InstanceVars    map[int64][]NameValue
	InstanceTables  map[int64]string
	ParentInstances map[int64]int64
	RootUUID        string
}

type NameValue struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
	Len   int         `json:"len"`
}

type RunProcessRequest struct {
	ProcessCode              string      `json:"processCode"`
	Task                     string      `json:"task"`
	FindInstanceByUniqueVars []NameValue `json:"findInstanceByUniqueVars"`
	ManualTaskPointCode      string      `json:"manualTaskPointCode"`
	Input                    []NameValue `json:"input"`
}

type RunUserTaskRequest struct {
	ProcessCode              string      `json:"processCode"`
	Task                     string      `json:"task"`
	FindInstanceByUniqueVars []NameValue `json:"findInstanceByUniqueVars"`
	UserTaskPointCode        string      `json:"userTaskPointCode"`
	Input                    []NameValue `json:"input"`
}

////Создание процесса
//func CreateInstanceByBPCode(process_code string,userId int64, input []NameValue) ([]NameValue,int64,int64,error) {
//
//	result,err := getProcessIDByCode(process_code)
//	if err!=nil{
//		return nil,0,0,err
//	}else {
//		return CreateInstance(result, userId, input)
//	}
//}

//Создание процесса

func (context *InstanceContext) GetTaskId(task string) (int64, error) {

	taskId := int64(0)
	err := context.O.Raw(utils.DbBindReplace("select id  from bp_tasks where sys$uuid=?"), task).QueryRow(&taskId)
	return taskId, err

}

func (context *InstanceContext) GetTaskUUID(taskId int64) (string, error) {

	taskUUID := ""
	err := context.O.Raw(utils.DbBindReplace("select sys$uuid from bp_tasks where id=?"), taskId).QueryRow(&taskUUID)
	return taskUUID, err

}

func (context *InstanceContext) checkInputVar(process_id, userId int64, input []NameValue) error {

	var rqNames []string
	_, err := cached.O().Raw(utils.DbBindReplace("select code from bp_process_vars where is_input = 1 and process_id=?"), process_id).QueryRows(&rqNames)
	//log.Printf("checkInputVar rqNames %v",rqNames)
	//log.Printf("checkInputVar input %v",input)
	if err != nil {
		return err
	}

	for _, v := range input {
		ok := false
		for _, si := range rqNames {

			if v.Name == si {
				if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
					log.Printf("checkInputVar name='%v' value='%v'", v.Name, v.Value)
				}
				ok = true
			}

		}
		if !ok {
			return errors.New(fmt.Sprintf("%v is not Input Variable", v.Name))
		}
	}
	return nil

}

func (context *InstanceContext) checkRqInputVar(process_id, userId int64, input []NameValue) error {

	var rqNames []string
	_, err := cached.O().Raw(utils.DbBindReplace("select code from bp_process_vars where is_input = 1 and is_rq = 1 and process_id=?"), process_id).QueryRows(&rqNames)
	//log.Printf("checkRqInputVar rqNames %v",rqNames)
	//log.Printf("checkRqInputVar input %v",input)
	if err != nil {
		return err
	}
	for _, si := range rqNames {
		ok := false
		for _, v := range input {
			if v.Name == si {
				//log.Printf("checkRqInputVar name='%v' value='%v'",v.Name,v.Value)
				ok = true
			}
			if v.Name == si && v.Value == "" {
				log.Println("Error 1. ", v.Name, "Required Input Variable. ProcessId", process_id, "input", input)
				return errors.New(fmt.Sprintf("%v is Required Input Variable", v.Name))
			}

		}
		if !ok {
			log.Println("Error 2. ", si, "Required Input Variable. ProcessId", process_id, "input", input)
			return errors.New(fmt.Sprintf("%v is Required Input Variable", si))
		}
	}
	return nil

}

func (context *InstanceContext) CreateInstance(req *http.Request, process_id, userId int64, input []NameValue, parent_id int64) ([]NameValue, string, string, int64, error) {

	//processCode,err := context.GetProcessCodeById(process_id)

	if req == nil {
		//log.Println("Create Instance Req Is Null")
	}

	err := context.checkRqInputVar(process_id, userId, input)

	if err != nil {
		debug.PrintStack()
		utils.ErrorWriteUser("BPMError", "CreateInstance process_id="+strconv.Itoa(int(process_id)), userId, err)
		return nil, "", "", int64(0), err
	}

	err = context.checkInputVar(process_id, userId, input)

	if err != nil {
		debug.PrintStack()
		utils.ErrorWriteUser("BPMError", "CreateInstance process_id="+strconv.Itoa(int(process_id)), userId, err)
		return nil, "", "", int64(0), err
	}

	//o := orm.NewOrm()
	//o.Using("default")

	processTitle, err := context.GetProcessTitleByProcessId(process_id)
	if err != nil {
		debug.PrintStack()
		return nil, "", "", int64(0), err
	}

	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("=======================================Creating Process ", processTitle, "=======================================")
	}
	//pointId, err := findFirstPoint(process_id)

	newInstance := utils.Uuid()

	if parent_id != 0 {
		context.RootUUID = context.GetRootUUID(parent_id)
	} else {
		context.RootUUID = newInstance
	}

	sql := "insert into bp_instances (root_uuid,parent_id,sys$uuid,is_finished,process_id,is_terminated,created_by) values (?,nullif(?,0),?,0,?,?,nullif(?,0))"
	if utils.GetDbDriverType() == orm.DROracle {
		sql = "insert into bp_instances (root_uuid,parent_id,sys$uuid,is_finished,process_id,is_terminated,created_by,start_ts) values (?,nullif(?,0),?,0,?,?,nullif(?,0),systimestamp)"
	}
	instanceId, err := utils.DbInsert(context.O, utils.DbBindReplace(sql), context.RootUUID, parent_id, newInstance, process_id, 0, userId)
	//Меняем только после создания. Как видим ссылку parent_id сохраняем

	if err != nil {
		log.Println("Error on CreateInstance GetProcessTableByProcessId", err)
		return nil, "", "", int64(0), err
	}

	tableName, err := context.GetProcessTableByProcessId(process_id)

	if err != nil {
		log.Println("Error on CreateInstance GetProcessTableByProcessId", err)
		return nil, "", "", int64(0), err
	}

	err = context.ReadAllVariablesFromDB(instanceId, process_id, tableName, userId)

	if err != nil {
		//debug.PrintStack()
		log.Println("CreateInstance 6.0.2 ReadAllVariablesFromDB " + err.Error())
		log.Println(instanceId)
		log.Println(tableName)
		log.Println(userId)
		return nil, "", "", int64(0), err
	}

	//Need Merge
	//context.InstanceVars[instanceId] = input
	err = context.SetVariablesToContext(instanceId, input)
	if err != nil {
		//debug.PrintStack()
		log.Println("CreateInstance SetVariablesToContext 6.0.7 "+err.Error(), instanceId)
		return nil, "", "", int64(0), err
	}

	//context.InstanceVars[instanceId] = input
	context.InstanceTables[instanceId] = tableName

	//context.InitContextDefaultVariables(instanceId, process_id, userId)

	if err != nil {
		log.Println("Error on CreateInstance", err)
		return nil, "", "", int64(0), err
	}

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("VERBOSE:=======================================Created Instance", instanceId)
	}

	err = context.ReadContextVariableToLua(instanceId, process_id, userId)

	if err != nil {
		log.Println("Error on ReadContextVariableToLua", err)
		return nil, "", "", int64(0), err
	}

	instance := newInstance

	point, err := context.findFirstPoint(process_id)
	if err != nil {
		//debug.PrintStack()
		log.Println("error CreateInstance.findFirstPoint " + err.Error())
		utils.ErrorWriteUser("BPMError", "process_id="+strconv.Itoa(int(process_id)), userId, err)
		return nil, "", "", int64(0), err
	}
	log.Println("========================ZZZZ", point)
	taskId, err := context.execInstance(req, instanceId, point, userId)

	if err != nil {
		log.Println("execInstance err1" + err.Error())
		return nil, "", "", int64(0), err
	}

	isTerminated := 0
	terminateText := ""
	err = context.O.Raw(utils.DbBindReplace("select error_text,is_terminated from bp_instances where id=?"), instanceId).QueryRow(&terminateText, &isTerminated)
	if isTerminated == 1 {
		return nil, "", "", int64(0), errors.New(terminateText)
	}
	if err != nil {
		log.Println("execInstance err2" + err.Error())
		return nil, "", "", int64(0), err
	}

	//log.Println("vata emes start", context.InstanceVars[instanceId])

	context.SaveVariables()
	output, _ := context.BindGlobalOuputVariablesToVar(instanceId, process_id)
	//fmt.Println ("output",output)
	//context.O.Begin()
	context.O.Raw("savepoint after_create_instance_" + strconv.Itoa(int(instanceId))).Exec()
	return output, instance, taskId, int64(0), err
}

//Создание подпроцесса
func (context *InstanceContext) CreateSubInstance(req *http.Request, process_id, parent_instance_id int64, parent_point_id int64, userId int64, input []NameValue) ([]NameValue, int64, string, error) {

	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("creating subprocess")
	}
	//pointId, err := findFirstPoint(process_id)
	sql := "insert into bp_instances (root_uuid,is_finished,process_id,is_terminated,parent_id,parent_point_id, created_by) values (?,0,?,?,?,?,nullif(?,0))"

	if utils.GetDbDriverType() == orm.DROracle {
		sql = "insert into bp_instances (root_uuid,is_finished,process_id,is_terminated,parent_id,parent_point_id, created_by,start_ts) values (?,0,?,?,?,?,nullif(?,0),systimestamp)"

	}

	roomUUID := context.GetRootUUID(parent_instance_id)

	instanceId, err := utils.DbInsert(context.O, utils.DbBindReplace(sql), roomUUID,
		process_id, 0, parent_instance_id, parent_point_id, userId)

	if err != nil {
		log.Println("Error on CreateSubInstance", err)
		return nil, int64(0), "", err
	}

	//context.InstanceVars[instanceId] = input

	tableName, err := context.GetProcessTableByProcessId(process_id)

	if err != nil {
		log.Println("Error on CreateSubInstance GetProcessTableByProcessId", err)
		return nil, int64(0), "", err
	}
	context.InstanceTables[instanceId] = tableName

	//log.Println("SSSSSSSSSSSSS", input)
	err = context.ReadAllVariablesFromDB(instanceId, process_id, tableName, userId)

	err = context.SetVariablesToContext(instanceId, input)

	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance SetVariablesToContext 6.0.7 "+err.Error(), instanceId)
		return nil, 0, "", err
	}

	err = context.ReadContextVariableToLua(instanceId, process_id, userId)

	if err != nil {
		log.Println("Error on CreateSubInstance ReadContextVariableToLua", err)
		return nil, int64(0), "", err
	}

	if err != nil {
		debug.PrintStack()
		utils.ErrorWriteUser("BPMError", "process_id="+strconv.Itoa(int(process_id)), userId, err)
		return nil, int64(0), "", err
	}

	/*err = context.SetGlobalVarsByInstance(instanceId,input)
	if err != nil{
		log.Println("Error on runSubProcess 2 "+err.Error())
		return nil,int64(0),"",err
	}*/

	//instanceId, err := rs.LastInsertId()
	//Warning! Bug!
	point, err := context.findFirstPoint(process_id)
	if err != nil {
		debug.PrintStack()
		utils.ErrorWriteUser("BPMError", "process_id="+strconv.Itoa(int(process_id)), userId, err)
		return nil, int64(0), "", err
	}

	task, err := context.execInstance(req, instanceId, point, userId)
	/*output, _ := context.BindGlobalOuputVariablesToVar(instanceId)
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("output", output)
	}*/
	return nil, instanceId, task, err
}

func (context *InstanceContext) GetProcessTitleByProcessId(processId int64) (string, error) {
	processTitle := ""
	err := cached.O().Raw(utils.DbBindReplace(`select title from bp_processes p where p.id=?`), processId).QueryRow(&processTitle)
	if err != nil {
		log.Println("Error on GetProcessTitleByProcessId", processId, err)
		return "", err
	}
	return processTitle, nil
}

func (context *InstanceContext) GetRootUUID(instanceId int64) string {

	rootUUID := ""
	err := context.O.Raw(utils.DbBindReplace(`select root_uuid from bp_instances i where i.id=?`), instanceId).QueryRow(&rootUUID)
	if err != nil {
		//debug.PrintStack()
		//utils.ErrorWrite("BPMError","GetProcessCodeByInstance instanceId="+strconv.Itoa(int(instanceId)),err)
		log.Println("Error on GetRootUUID", instanceId, err)
		return rootUUID
	}

	return rootUUID
}

func (context *InstanceContext) GetProcessCodeByInstance(instanceId int64) (string, error) {

	processCode := ""
	err := context.O.Raw(utils.DbBindReplace(`select p.code from bp_instances i,bp_processes p where i.id=? and p.id=i.process_id`), instanceId).QueryRow(&processCode)
	if err != nil {
		//debug.PrintStack()
		//utils.ErrorWrite("BPMError","GetProcessCodeByInstance instanceId="+strconv.Itoa(int(instanceId)),err)
		log.Println("Error on GetProcessCodeByInstance", instanceId, err)
		return "", err
	}

	return processCode, err
}

func (context *InstanceContext) GetProcessIdByPointId(pointId int64) (int64, error) {

	processId := int64(0)
	err := cached.O().Raw(utils.DbBindReplace(`select process_id from bp_points where id=?`), pointId).QueryRow(&processId)
	if err != nil {
		debug.PrintStack()
		utils.ErrorWrite("BPMError", "GetProcessIdByPointId pointId="+strconv.Itoa(int(pointId)), err)
		return 0, err
	}

	return processId, err
}

func (context *InstanceContext) GetProcessIdByInstance(instanceId int64) (int64, error) {

	processId := int64(0)
	err := context.O.Raw(utils.DbBindReplace(`select process_id from bp_instances where id=?`), instanceId).QueryRow(&processId)
	if err != nil {
		debug.PrintStack()
		utils.ErrorWrite("BPMError", "GetProcessIdByInstance instanceId="+strconv.Itoa(int(instanceId)), err)
		return 0, err
	}
	log.Println("GetProcessIdByInstance: intance id = " + strconv.Itoa(int(instanceId)))
	log.Println("GetProcessIdByInstance: process id = " + strconv.Itoa(int(processId)))

	return processId, err
}

func (context *InstanceContext) getNextPointsFromParallelInclusiveGateway(req *http.Request, instanceId int64, task string, point, user_id int64) ([]int64, error) {

	var result []int64
	//pt, err := context.getPointTypeCodeByTask(task)
	pt, err := context.getPointTypeCodeByPointId(point)
	if err != nil {
		debug.PrintStack()
		utils.ErrorWrite("BPMError", "getNextPointsFromParallelInclusiveGateway getPointTypeCodeByPointId"+strconv.Itoa(int(point)), err)
		return result, err
	}
	if pt != "parallelgateway" && pt != "inclusivegateway" {

		debug.PrintStack()
		return result, errors.New("only parallelgateway or inclusive gets next points")
	}

	type typeCond struct {
		Cond  string `json:"cond"`
		Point int64  `json:"point"`
		SfId  int64  `json:"sf_id"`
	}

	var conds = []typeCond{}
	//cached.ClearCache() // example

	//context.O
	_, err = cached.O().Raw(utils.DbBindReplace(`select cond as "cond",ps2.point_id as "point",sf.id as "sf_id"  from
			  bp_sequence_flows sf,
			  bp_point_sfs ps1,
			  bp_point_sfs ps2
where ps1.sf_id=sf.id and
      ps1.point_id=? and
      ps1.is_incoming=0 and
      ps2.is_incoming=1 and sf.is_active=1 and ps1.is_active=1 and ps2.is_active=1 and
      ps2.sf_id=ps1.sf_id`), point).QueryRows(&conds)

	if err != nil {
		return result, err
	}

	for _, element := range conds {

		b, err := context.LuaCondition(context.O, req, instanceId, element.Point, user_id, element.SfId, element.Cond)
		if err != nil {
			return result, err
		}
		if b {

			if os.Getenv("CRM_DEBUG_BPMS") == "1" {
				log.Println("getNextPointsFromParallelInclusiveGateway ")
				log.Println(element.Point)
			}
			result = append(result, element.Point)
		}
	}

	if len(result) == 0 {
		return result, errors.New("No Conditions")
	} else {
		return result, nil
	}
}

func (context *InstanceContext) getNextPointFromExclusiveGateway(req *http.Request, instanceId int64, pointId int64, user_id int64) (int64, error) {
	pt, err := context.getPointTypeCodeByPointId(pointId)
	if err != nil {
		debug.PrintStack()
		utils.ErrorWrite("BPMError", "getNextPointFromExclusiveGateway ", err)
		return int64(0), err
	}
	if pt != "exclusivegateway" {
		debug.PrintStack()
		return int64(0), errors.New("only exclusivegateway can calculate next point")
	}

	type typeCond struct {
		Cond  string `json:"cond"`
		Point int64  `json:"point"`
		SfId  int64  `json:"sf_id"`
	}

	var conds = []typeCond{}
	//context.O.Raw
	_, err = cached.O().Raw(utils.DbBindReplace(`select cond as "cond",ps2.point_id as "point",sf.id as "sf_id"  from
			  bp_sequence_flows sf,
			  bp_point_sfs ps1,
			  bp_point_sfs ps2
where ps1.sf_id=sf.id and
      ps1.point_id=? and
      ps1.is_incoming=0 and
      ps2.is_incoming=1 and sf.is_active=1 and ps1.is_active=1 and ps2.is_active=1 and
      ps2.sf_id=ps1.sf_id`), pointId).QueryRows(&conds)

	if err != nil {
		return int64(0), err
	}
	for _, element := range conds {

		b, err := context.LuaCondition(context.O, req, instanceId, element.Point, user_id, element.SfId, element.Cond)
		if err != nil {
			return 0, err
		}
		if b {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("getNextPointFromExclusiveGateway ", element.Point)
			}
			return element.Point, nil
		}

	}

	return 0, errors.New("No Conditions")
}

func (context *InstanceContext) checkCondString(cond string) (bool, error) {
	if strings.TrimSpace(cond) == "" {
		return true, nil
	}

	cnt := 0
	err := context.O.Raw(utils.DbBindReplace("select count(1) cnt from dual where " + cond)).QueryRow(&cnt)
	return cnt > 0, err
}

func (context *InstanceContext) isLoopPoint(point int64) (bool, error) {

	var isLoops []int64
	_, err := cached.O().Raw(utils.DbBindReplace("select p.is_loop from bp_points p where p.id=?"), point).QueryRows(&isLoops)
	return (len(isLoops) > 0 && isLoops[0] == 1), err
}

func runServiceTask(instanceId int64) error {
	log.Println("SERVICE TASK DONE")
	return nil
}

func test(L *lua.State) int {
	log.Println("hello world! from go!!!")
	return 0
}

func (context *InstanceContext) runSubProcess(req *http.Request, instanceId int64, pointId, userId int64) (string, error) {
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("========runSubProcess POINT START ", instanceId, pointId)
	}

	type typeRows struct {
		SubProcess int64  `json:"subprocess"`
		Title      string `json:"title"`
	}
	var row []typeRows

	_, err := cached.O().Raw(utils.DbBindReplace(`select p.subprocess_id as "sub_process",p.title as "title" from bp_points p where p.id=?`), pointId).QueryRows(&row)
	if err != nil {
		log.Println("Error on runSubProcess 1 " + err.Error())
		return "", err
	}

	if len(row) == 0 {
		log.Println("Error on runSubProcess 2  " + err.Error())
		return "", err
	}

	//	err = WritePointJRN(instanceId,userId,nil)
	//	if err!=nil{
	//		log.Println(err)
	//		debug.PrintStack()
	//		return err
	//	}

	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("=========================================Doing runSubProcess `" + row[0].Title + "`========================================= ")
	}
	//pointId, err := context.GetPointByTask(parentTask)
	input, _ := context.BindInputVariablesToSubProcess(instanceId, pointId, userId)

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("subprocess input", instanceId, pointId, userId, input)
	}

	_, _, task, err := context.CreateSubInstance(req, row[0].SubProcess, instanceId, pointId, userId, input)
	if err != nil {
		log.Println("Error on runSubProcess 1 " + err.Error())
		return "", err
	}
	//context.BindGlobalOuputVariablesToVar()
	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("=======================================Done runSubProcess `" + row[0].Title + "`================================")
	}
	return task, nil
}

func (context *InstanceContext) runParallelInclusive(req *http.Request, instanceId int64, task string, point, userId int64, pointType string) (string, error) {

	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("runParallelInclusive task = ", task)
	}
	gwTask := task

	cnt := 0
	err := context.O.Raw(utils.DbBindReplace("select count(1) from bp_point_waits where instance_id=? and point_id=?"), instanceId, point).QueryRow(&cnt)
	if err != nil {
		log.Println("error runParallelInclusive 1")
		return "", err
	}
	if cnt == 0 {
		log.Println("task = " + task)
		_, err := context.O.Raw(utils.DbBindReplace("insert into bp_point_waits (instance_id,point_id,is_wait) (select ?,point_id,1 from bp_point_sfs where point_id=? and is_incoming=1 and is_active=1)"), instanceId, point).Exec()
		if err != nil {
			log.Println("error runParallelInclusive 2")
			return "", err
		}
	}

	sql := "update bp_point_waits set is_wait=0 where  instance_id=? and point_id=? and is_wait=1  limit 1"
	if utils.GetDbDriverType() == orm.DRPostgres {
		sql = "update bp_point_waits set is_wait=0 where id= (select id from bp_point_waits where instance_id=? and point_id=? and is_wait=1 limit 1)"
	}
	_, err = context.O.Raw(utils.DbBindReplace(sql), instanceId, point).Exec()
	if err != nil {
		log.Println("error runParallelInclusive 2.1")
		return "", err
	}

	if pointType == "parallelgateway" {
		err = context.O.Raw(utils.DbBindReplace("select count(1) from bp_point_waits where instance_id=? and is_wait=1"), instanceId).QueryRow(&cnt)
		if err != nil {
			log.Println("error runParallelInclusive 3")
			return "", err
		}
		if cnt > 0 {
			log.Println("breaking parallel task = " + task)
			context.closeTask(task, point,userId)

			err = context.O.Raw(utils.DbBindReplace("select t.sys$uuid from bp_tasks t,bp_points p,bp_point_types pt where t.instance_id=? and t.point_id=p.id and p.type_id=pt.id and pt.code in ('usertask','endevent') and t.is_open=1 "), instanceId).QueryRow(&task)
			log.Println("found task = " + task)
			return task, nil
		}
	}

	res, err := context.getNextPointsFromParallelInclusiveGateway(req, instanceId, task, point, userId)

	log.Println("task  gw gw = " + task)

	log.Println(res)
	if err != nil {
		return "", err
	} else {
		context.closeTask(task, point,userId)
		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("CLOSE " + task)
		}
		for _, v := range res {

			task, err = context.execInstance(req, instanceId, int64(v), userId)

			if pointType == "inclusivegateway" {
				_, err = context.O.Raw(utils.DbBindReplace("insert into bp_task_auto_closes (task_id,parent_id) values ((select id from bp_tasks where sys$uuid=?),(select id from bp_tasks where sys$uuid=?))"), task, gwTask).Exec()
				if err != nil {
					return "", err
				}
			}

		}

		//		taskId,err := GetTaskId(gwTask)
		//		if err!=nil {
		//			return "", err
		//		}

		//		_,err = o.Raw("update bp_tasks set is_open=0 where parent_id=? and sys$uuid<>?", taskId,task).Exec()
		//		if err!=nil {
		//			return "", err
		//		}else{
		//			log.Println("deleted "+task)
		//		}

	}

	err = context.O.Raw(utils.DbBindReplace("select t.sys$uuid from bp_tasks t,bp_points p,bp_point_types pt where t.instance_id=? and t.point_id=p.id and p.type_id=pt.id and pt.code in ('usertask','endevent') and t.is_open=1 and t.user_id=?"), instanceId, userId).QueryRow(&task)
	//Возможно, это глючное решение
	if utils.IsNoRowFound(err) {
		return task, nil
	} else {
		return task, err
	}
}
func (context *InstanceContext) runScriptTask(req *http.Request, instanceId, point, userId int64) error {
	//log.Println("SCRIPT POINT START ")

	type typeRows struct {
		Code   string `json:"code"`
		Title  string `json:"title"`
		Script string `json:"script"`
		Type   string `json:"type"`
	}
	var row []typeRows

	_, err := cached.O().Raw(utils.DbBindReplace(`select
	p.code as "code",p.title as "title",p.script_txt as "script",st.code as "type" from bp_points p,
	bp_script_types st
	where p.id=? and p.script_type_id=st.id`), point).QueryRows(&row)
	if err != nil {
		log.Println("error on get script", err)
		return err
	}
	if len(row) == 0 {
		log.Println("nothing todo. Script not Set")
		return nil
	}

	if err != nil {
		log.Println("Error on runScriptTask", err)
		return err
	}
	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("=======================================Doing " + row[0].Title + "========================================= ")
	}
	if row[0].Type == "lua" {
		err = context.runLuaScriptTask(context.O, req, instanceId, point, userId, row[0].Script)
		if err != nil {
			log.Println("runLuaScriptTask ERROR!!! "+row[0].Code, " ", err.Error())
			//context.O.Rollback()
		} else {
			//o.Commit()
		}
		return err
	} else if row[0].Type == "sql_query_row" {
		return nil
		//runSQLQueryRowScriptTask(instanceId, point, row[0].Script)
	}

	log.Println("script nothing todo")

	utils.ClearInterface(&row)
	return nil
}

func (context *InstanceContext) GetPointIdByAlias(alias string) (int64, error) {

	var points []int64
	_, err := cached.O().Raw(utils.DbBindReplace("select id from bp_points where alias=?"), alias).QueryRows(&points)
	if err == nil {
		return points[0], err
	}
	return int64(0), err
}

func (context *InstanceContext) GetPointIdByCode(code string) (int64, error) {

	var points []int64
	err := cached.O().Raw(utils.DbBindReplace("select id from bp_points where code=?"), code).QueryRow(&points)
	if err == nil {
		return points[0], err
	}
	return int64(0), err
}

//Проверка обязательности выходных полей
func (context *InstanceContext) CheckRqOutputIntPointVars(pointId int64, input []NameValue) (map[string]string, error) {

	errorsMaps := make(map[string]string)
	var vars []string
	hasErrors := false
	_, err := cached.O().Raw(utils.DbBindReplace(`select pv.code from bp_point_vars pov
	join bp_process_vars pv on pov.var_id=pv.id
	where
	pov.point_id=? and pov.is_rq=1
	`), pointId).QueryRows(&vars)
	if err != nil {
		return errorsMaps, err
	}
	for _, v := range vars {
		ok := false
		for _, inputValue := range input {
			if inputValue.Name == v && inputValue.Value != nil && fmt.Sprintf("%v", inputValue.Value) != "" {
				//log.Println("inputValue")
				//log.Println(inputValue)
				ok = true
			}
		}
		if !ok {
			hasErrors = true
			errorsMaps[v] = "Required Variable2"
			log.Println(v, " Required")
		}
	}

	if hasErrors {
		return errorsMaps, errors.New("Required Variable")
	} else {
		return errorsMaps, nil
	}
}

func (context *InstanceContext) TerminateByInstanceId(instanceId int64, err error) error {

	err_text := "UNKNOWN ERROR"
	if err != nil {
		err_text = err.Error()
	}
	_, err2 := context.O.Raw(utils.DbBindReplace("update bp_instances set is_terminated = 1,error_text = substr(?,1,250) where id=?"), err_text, instanceId).Exec()
	if err2 != nil {
		return err2
	}
	_, err3 := context.O.Raw(utils.DbBindReplace("update bp_tasks set is_open = 0, is_terminated = 1,error_text = substr(?,1,250) where instance_id=? and is_open=1"), err_text, instanceId).Exec()
	if err3 != nil {
		return err3
	}

	context.O.Raw("rollback to after_create_instance_" + strconv.Itoa(int(instanceId))).Exec()

	return nil
}

func (context *InstanceContext) TaskIsOpen(task string) bool {

	id := int64(0)
	err := context.O.Raw(utils.DbBindReplace("select id from bp_tasks where sys$uuid = ? and is_open = 1"), task).QueryRow(&id)
	return id != 0 && err == nil

}

func (context *InstanceContext) GetInstanceFinishedByUUID(instanceUUID string) bool {
	is_finished := 1
	context.O.Raw(utils.DbBindReplace("select coalesce(is_finished,0) as is_finished from bp_instances where sys$uuid = ?"), instanceUUID).QueryRow(&is_finished)
	return is_finished == 1
}

func (context *InstanceContext) GetInstanceUUID(instanceId int64) (string, error) {
	result := ""
	err := context.O.Raw(utils.DbBindReplace("select sys$uuid from bp_instances where id = ?"), instanceId).QueryRow(&result)
	return result, err
}

func (context *InstanceContext) ManualExecInstanceByTask(req *http.Request, task string, userId int64, input []NameValue) (string, error) {

	if !context.TaskIsOpen(task) {
		return "", errors.New("TASK " + task + " IS CLOSED")
	}
	instanceId, pointId, err := context.GetInstanceAndPointByTask(task)
	if err != nil {
		debug.PrintStack()
		return "", err
	}

	process_id, err := context.GetProcessIdByPointId(pointId)

	if err != nil {
		debug.PrintStack()
		return "", err
	}

	context.RootUUID = context.GetRootUUID(instanceId)

	err = context.ReadContextVariableToLua(instanceId, process_id, userId)

	if err != nil {
		log.Println("Error on ManualExecInstanceByTask/ReadContextVariableToLua", err)
		return "", err
	}

	loop, err := context.isLoopPoint(pointId)
	if err != nil {
		debug.PrintStack()
		log.Println("ManualExecInstanceByTask isLoopPointByTask 3 " + err.Error())
		return "", err
	}

	pt, err := context.getPointTypeByTask(task)
	if err != nil {
		debug.PrintStack()
		return "", err
	}

	if !loop && pt == "scripttask" {
		if pt != "inclusivegateway" && pt != "parallelgateway" && pt != "intermediatecatchevent" && pt != "usertask" && pt != "manualtask" && pt != "subprocess" {
			return "", errors.New("manual exec only for manualTask and userTask or loop scripttask. break")
		}
	}

	/*
		//OPTIMIZE
		err = context.SetGlobalVarsByInstance(instanceId, input)
		if err != nil {
			debug.PrintStack()
			log.Println("ManualExecInstance 4 " + err.Error())
			return nil, "", err
		}*/

	//	point,err := gotoNextPoint(instanceId,userId)
	//	if err!=nil{
	//		debug.PrintStack()
	//		log.Println("ManualExecInstance 5 " + err.Error())
	//		return 0,err
	//	}

	processTable, err := context.GetProcessTableByProcessId(process_id)

	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance 6.0.8 processTable " + err.Error())
		log.Println(instanceId)
		log.Println(processTable)
		log.Println(userId)
		return "", err
	}

	err = context.ReadAllVariablesFromDB(instanceId, process_id, processTable, userId)

	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance 6.0.2 ReadAllVariablesFromDB " + err.Error())
		log.Println(instanceId)
		log.Println(processTable)
		log.Println(userId)
		return "", err
	}

	nextPointId, err := context.getNextPoint(req, instanceId, pointId, userId)
	if err != nil {
		debug.PrintStack()
		log.Println("ManualExecInstance 5 " + err.Error())
		return "", err
	}

	context.closeTask(task, pointId,userId)

	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance.GetProcessTableByProcessId 6.01 " + err.Error())
		log.Println(instanceId)
		log.Println(nextPointId)
		return "", err
	}

	//Need Merge
	//context.InstanceVars[instanceId] = input
	onlyOutput, err := context.OnlyOutPutVars(input, pointId)

	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance.OnlyOutPutVars 6.02 " + err.Error())
		log.Println(instanceId)
		log.Println(nextPointId)
		return "", err
	}

	err = context.SetVariablesToContext(instanceId, onlyOutput)
	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance SetVariablesToContext 6.0.7 "+err.Error(), instanceId)
		return "", err
	}
	context.InstanceTables[instanceId] = processTable
	err = context.ReadContextVariableToLua(instanceId, process_id, userId)

	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance 6.1 " + err.Error())
		log.Println(instanceId)
		log.Println(nextPointId)
		return "", err
	}

	task, err = context.execInstance(req, instanceId, nextPointId, userId)
	if err != nil {
		//debug.PrintStack()
		log.Println("ManualExecInstance 6 " + err.Error())
		log.Println(instanceId)
		log.Println(nextPointId)
		return "", err
	}
	//output, _ := context.BindGlobalOuputVariablesToVar(instanceId)

	context.SaveVariables()
	return task, nil
}

func (context *InstanceContext) closeTask(task string, pointId int64,userId int64) error {

	pt, err := context.getPointTypeCodeByPointId(pointId)

	if err != nil {
		log.Println("closeTask getPointTypeCodeByPointId error", err)
		return err
	}

	if pt != "intermediatecatchevent" && pt != "usertask" && pt != "manualtask" && pt != "subprocess" {
		return nil
	}

	closeSql := "update bp_tasks set closed_at=now(),closed_by=nullif(?,0), is_open=0 where sys$uuid=?"
	
	if utils.GetDbDriverType() == orm.DROracle {
		closeSql = "update bp_tasks set closed_at=sysdate,closed_by=nullif(?,0), is_open=0 where sys$uuid=?"
	}
	
	_, err = context.O.Raw(utils.DbBindReplace(closeSql), userId, task).Exec()
	if err != nil {
		log.Println("Error closeTaskByInstance 1 "+err.Error(), task)
		return err
	}

	taskId, err := context.GetTaskId(task)

	/*if err != nil {
		log.Println("Error closeTaskByInstance  2 "+err.Error(), task)
		return err
	}*/

	//Автоматически подчищать для Inclusive GateWay

	parent_id := int64(0)
	err = context.O.Raw(utils.DbBindReplace(`select parent_id from bp_task_auto_closes where task_id=?`), taskId).QueryRow(&parent_id)

	if err != nil {
		if !utils.IsNoRowFound(err) {
			log.Println("Error closeTaskByInstance 3 "+err.Error(), taskId)
			return err
		}
	}

	if parent_id != 0 {
		_, err = context.O.Raw(utils.DbBindReplace(`update bp_tasks set is_open = 0 where id in (select c.task_id from bp_task_auto_closes c where c.parent_id=?)`), parent_id).Exec()

		if err != nil {
			log.Println("Error closeTaskByInstance 4 "+err.Error(), parent_id)
			return err
		}
	}

	_, err = context.O.Raw(utils.DbBindReplace(`delete from bp_task_auto_closes where parent_id=?`), parent_id).Exec()

	return err

}

//Получаем Актеров
func (context *InstanceContext) getActorsByPointInstance(instanceId, point, currentUserId int64) ([]string, error) {

	//processEntity := ""
	//actorType := ""
	//actorVarCode := ""
	//actorUserId := int64(0)
	//actorRoleId := int64(0)

	type actorStruct struct {
		ActorVarCode  string
		ActorUserId   string
		ActorRoleId   string
		ActorType     string
		ProcessEntity string
	}

	var actorData []actorStruct

	sql := `select
	v.code as "actor_var_code", p.actor_user_id as "actor_user_id",
	p.actor_role_id as "actor_role_id",
	at.code as "actor_type",
	concat('i$',pr.code) as "process_entity"
	from bp_points p
	join  bp_processes pr on pr.id=p.process_id
	join bp_actor_types at  on at.id=p.actor_type_id
	left join bp_process_vars v on v.id=p.actor_var_id
	 where  p.id=?`

	_, err := cached.O().Raw(utils.DbBindReplace(sql), point).QueryRows(&actorData)
	if err != nil {
		log.Println("getActorByPointInstance err1 " + err.Error())
		return []string{}, err
	}

	if len(actorData) == 0 {
		return []string{}, errors.New("No role info found")
	}
	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("ActorType", actorData[0].ActorType)
	}

	if actorData[0].ActorType == "actor" {
		var sql_query []string

		//log.Println("test")

		sql := `select
		case when pa.is_role = 1 then
			concat('select group_concat(user_id) from user_roles ur,roles r where r.id=',pa.role_id,' and r.id=ur.role_id')
			 when pa.is_query = 1 then
		pa.sql_query end  from bp_actors pa,bp_points p where pa.id=p.actor_id and p.id=?`

		if utils.GetDbDriverType() == orm.DROracle {

			sql = `select
			case when pa.is_role = 1 then
				TO_CLOB('SELECT  LISTAGG(user_id, '','') WITHIN GROUP (ORDER BY user_id) "user_id" from user_roles ur,roles r
where r.id=1 and r.id=ur.role_id')
				 when pa.is_query = 1 then
			pa.sql_query end  from bp_actors pa,bp_points p where pa.id=p.actor_id and p.id=?`

		}
		//Этот запрос кешировать нельзя, ведь тут users, которые динамические
		_, err = context.O.Raw(utils.DbBindReplace(sql), point).QueryRows(&sql_query)
		if err != nil {
			log.Println("getActorByPointInstance err2 " + err.Error())
			return []string{}, err
		}
		if len(sql_query) == 0 {
			return []string{}, errors.New("getActorByPointInstance len(sql_query) == 0")
		}

		sql_query[0] = strings.Replace(sql_query[0], ":user_id", strconv.Itoa(int(currentUserId)), -1)

		usersString := ""
		sql = "select (" + sql_query[0] + " ) \"user\" from bp_instances instance left join " + actorData[0].ProcessEntity + " var on var.id$=instance.id where instance.id=? "
		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("getActorByPointInstance SQL=", sql)
		}
		err = context.O.Raw(utils.DbBindReplace(sql), instanceId).QueryRow(&usersString)
		if err != nil {
			log.Println("getActorByPointInstance err3 " + sql + " \r\n" + err.Error())
			return []string{}, err
		}
		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("getActorByPointInstance Actor=", usersString)
		}

		return strings.Split(usersString, ","), err
	} else if actorData[0].ActorType == "user" {
		log.Println("actor user", actorData[0].ActorUserId)
		return []string{actorData[0].ActorUserId}, nil
	} else if actorData[0].ActorType == "role" {
		var userRolesArr []string
		context.O.Raw(utils.DbBindReplace("select user_id from user_roles where role_id=?"), actorData[0].ActorRoleId).QueryRows(&userRolesArr)
		return userRolesArr, nil
	} else if actorData[0].ActorType == "var" {
		value := ""
		context.O.Raw(utils.DbBindReplace("select "+actorData[0].ActorVarCode+" from "+actorData[0].ProcessEntity+" where id$=?"), instanceId).QueryRow(&value)

		return []string{value}, nil
	} else {
		return []string{}, errors.New("No Actor Found")
	}
}

func (context *InstanceContext) setTimerAtToTimerEventTask(task string) error {

	sql := `update bp_tasks t
	join bp_points po on po.id=t.point_id
  set t.timer_at =
  DATE_ADD(
	  DATE_ADD(
		  DATE_ADD(
			  now(),
			  interval coalesce(po.timer_day,0) day),
		  interval coalesce(po.timer_hour,0) hour),
	  interval coalesce(po.timer_min,0) minute)

  where po.by_timer=1 and t.sys$uuid=?`

	if utils.GetDbDriverType() == orm.DROracle {
		sql = `update bp_tasks t

	  set t.timer_at =
	  (select
	  SYSDATE + coalesce(po.timer_day,0)
	   + coalesce(po.timer_hour,0)/24
	   + coalesce(po.timer_min,0)/24/60
		from  bp_points po
	  where po.id=t.point_id
	   and po.by_timer=1
	   )

	   where t.sys$uuid=?`
	}
	_, err := context.O.Raw(utils.DbBindReplace(sql), task).Exec()

	//Опасный коммит, который приводил к ошибкам
	//_, err = context.O.Raw("commit").Exec()

	return err

}
func (context *InstanceContext) startEvent(instanceId, point, userId, status int64, pointTypeCode string) {
	if status != 0 {
		type TBpTaskStartEvent = struct {
			Title  string
			Script string
		}

		bpTaskStartEvents := []TBpTaskStartEvent{}
		_, err := cached.O().Raw(`select title as "title",script as "script" from bp_task_start_events where is_active=1 order by nn`).QueryRows(&bpTaskStartEvents)
		if err == nil {

			input := make(map[string]interface{})

			for _, v := range bpTaskStartEvents {

				if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
					log.Println("VERBOSE:Doing bpTaskStartEvents", v)
				}

				input["user_id"] = userId
				input["instance_id"] = instanceId
				input["point_id"] = point
				//input["task_uuid"] = taskUUID
				input["point_type_code"] = pointTypeCode
				DeepPush(context.Lua, input)
				context.Lua.SetGlobal("input")

				err = lua.DoString(context.Lua, v.Script)
				if err != nil {
					log.Println("Error on lua DoString bpTaskStartEvents", err)
				}
			}
		} else {
			log.Println("Error on bpTaskStartEvents", err)
		}
	}

}
func (context *InstanceContext) createTask(instanceId, point, currentUserId int64) (string, error) {

	pt, status, err := context.getPointTypeAndWFStatus(point)

	if err != nil {
		log.Println("createTask err getPointType " + err.Error())
		return "", err
	}

	userId := currentUserId

	newTaskUUID := ""

	if pt == "endevent" {
		context.O.Raw(utils.DbBindReplace("update bp_instances i set finish_point_id= ? where id= ?"), point, instanceId).Exec()
		//context.startEvent(instanceId, point, userId, status, pt)
	} else if pt == "inclusivegateway" || pt == "parallelgateway" ||

		//pt == "endevent" ||  OPTIMIZE

		pt == "intermediatecatchevent" || pt == "usertask" || pt == "manualtask" {
		//|| pt == "subprocess"

		newTaskUUID = utils.Uuid()

		lid, err := utils.DbInsert(context.O, utils.DbBindReplace("insert into bp_tasks (sys$uuid,instance_id,point_id,user_id,is_open) values (?,?,?,nullif(?,0),?)"), newTaskUUID, instanceId, point, userId, 1)
		if err != nil {
			log.Println("error on createTask 55555 " + err.Error())
			return "", err
		}
		if pt == "usertask" {
			usersString, err := context.getActorsByPointInstance(instanceId, point, currentUserId)
			if err != nil {
				return "", err
			}
			for _, v := range usersString {
				if v != "" && v != "0" {
					_, err := context.O.Raw(utils.DbBindReplace("insert into bp_task_actors (task_id,user_id) values (?,?)"), lid, v).Exec()
					if err != nil {
						log.Println("create Task Error " + err.Error())
						return "", err
					}
				}
			}
		}

	}

	context.startEvent(instanceId, point, userId, status, pt)

	return newTaskUUID, nil
}

//Выполняем текущий шаг
func (context *InstanceContext) execInstance(req *http.Request, instanceId, point, userId int64) (string, error) {

	if req == nil {
		//log.Println("execInstance req is null")
	}
	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("=======================================Exec Instance Start =======================================", point)
	}

	//if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
	//	log.Println("input: ", input)
	//}

	//err := context.SetGlobalVarsByInstance(instanceId, input)
	//if err != nil {
	//	log.Println("SetGlobalVarsByInstance 66 ", err)
	//	debug.PrintStack()
	//	log.Println("execInstance 1 " + err.Error())
	//	return "", err
	//}

	//log.Print("execInstance point "+strconv.Itoa(int(point)))
	pt, err := context.getPointType(point)
	if err != nil {
		debug.PrintStack()
		log.Println("execInstance 2 " + err.Error())
		return "", err
	}

	loop, err := context.isLoopPoint(point)
	if err != nil {
		debug.PrintStack()
		log.Println("execInstance 3 " + err.Error())
		return "", err
	}

	//taskId,err := createTask(instanceId,point,userId)

	if pt == "usertask" || pt == "manualtask" {
		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("detect ", pt, " break")
		}
		taskId, err := context.createTask(instanceId, point, userId)
		return taskId, err
	}

	if pt == "intermediatecatchevent" {
		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("detect 2 ", pt, " break")
		}
		taskId, err := context.createTask(instanceId, point, userId)
		err = context.setTimerAtToTimerEventTask(taskId)
		return taskId, err
	}

	taskId, err := context.createTask(instanceId, point, userId)
	if err != nil {
		log.Println("createTask err5 " + err.Error())
		return "", err
	}

	if loop && pt == "scripttask" {
		log.Println("detect loop service task. break")
		return taskId, nil
	}
	//
	//	name,err := getPointNameByInstance (instanceId)
	//	if err!=nil{
	//
	//		debug.PrintStack()
	//		log.Println("execInstance 4 " + err.Error())
	//		return 0,err
	//	}
	//	log.Println("executing "+pt+ ".... "+name)
	//	log.Println(loop)

	if pt == "startevent" {
		context.closeTask(taskId, point,userId)
	}

	if pt == "endevent" {
		context.closeTask(taskId, point,userId)
		if os.Getenv("CRM_VERBOSE_SQL") == "1" {
			log.Println("closed", pt)
		}
	}

	if pt == "servicetask" {
		runServiceTask(instanceId)
	}
	if pt == "scripttask" {
		err := context.runScriptTask(req, instanceId, point, userId)
		if err != nil {
			//log.Println(err)
			log.Println("execInstance 4.1 " + err.Error())
			debug.PrintStack()

			err2 := context.TerminateByInstanceId(instanceId, err)
			if err2 != nil {
				log.Println("execInstance 4.1.1 " + err.Error())
				err = err2
			}
			return taskId, err
		}

		//Перечитка, вдруг LUA поменял данные
		err = context.ReadFromLuaToContext(instanceId, point, userId)
		if err != nil {
			//log.Println(err)
			log.Println("execInstance 5511 " + err.Error())
			debug.PrintStack()
			return taskId, err
		}

		err = context.closeTask(taskId, point, userId)
		if err != nil {
			return taskId, err
		}
	}

	//Запуск подпроцесса

	if pt == "parallelgateway" || pt == "inclusivegateway" {
		return context.runParallelInclusive(req, instanceId, taskId, point, userId, pt)
	}
	if pt == "subprocess" {
		taskId, err = context.runSubProcess(req, instanceId, point, userId)
		if err != nil {
			//log.Println(err)
			log.Println("execInstance SubProcess 6 " + err.Error())
			debug.PrintStack()
			return taskId, err
		} else {
			//log.Println("vata emes!!!" + taskId)
		}

		//err = context.closeTask(taskId)
		//if err!=nil{
		//	return taskId,err
		//}

		//Выходим. Толчок выхода сделает дочерний процесс
		return taskId, nil
	}

	if pt == "endevent" {
		sql := "update bp_instances set is_finished=1 where id=?"
		if utils.GetDbDriverType() == orm.DROracle {
			sql = "update bp_instances set is_finished=1,finish_ts=systimestamp where id=?"
		}
		_, err = context.O.Raw(utils.DbBindReplace(sql), instanceId).Exec()

		context.closeTask(taskId, point, userId)
		_, err = context.O.Raw(utils.DbBindReplace("update bp_tasks set is_open=0 where instance_id=?"), instanceId).Exec()

		if os.Getenv("CRM_DEBUG_BPMS") == "1" {
			log.Println("=======================================END PROCESS `", instanceId, "`=======================================")
		}

		parent_instance_id := int64(0)
		parent_point_id := int64(0)

		err = context.O.Raw(utils.DbBindReplace(
			`select
		main.parent_id,main.parent_point_id
		 from bp_instances main
		 where main.id=?`),
			instanceId).QueryRow(&parent_instance_id, &parent_point_id)

		if err != nil {
			log.Println("execInstance 7 " + err.Error())

			debug.PrintStack()
			return taskId, err
		}

		if parent_instance_id != 0 && parent_point_id != 0 {

			if os.Getenv("CRM_DEBUG_SQL") == "1" {
				log.Println("Resume parent parent_instance_id ", parent_instance_id)
			}
			//Продолжаем родительский процесс
			//Warning! Bug!
			point, err = context.getNextPoint(req, parent_instance_id, parent_point_id, userId)
			if err != nil {
				//debug.PrintStack()
				log.Println("execInstance 8 " + err.Error())
				return taskId, err
			}
			parent_process_id, err := context.GetProcessIdByPointId(parent_point_id)
			if err != nil {
				log.Println("execInstance BindOutputVariablesFromSubProcess 9.1 " + err.Error())
				debug.PrintStack()
				return taskId, err
			}

			err = context.BindOutputVariablesFromSubProcess(instanceId)
			if err != nil {
				log.Println("execInstance BindOutputVariablesFromSubProcess 9.2 " + err.Error())
				debug.PrintStack()
				return taskId, err
			}
			//Возвращаем переменные родительского процесса на место в контекст (state) луа
			err = context.ReadContextVariableToLua(parent_instance_id, parent_process_id, userId)
			if err != nil {
				//log.Println(err)
				log.Println("execInstance BindOutputVariablesFromSubProcess 9 " + err.Error())
				debug.PrintStack()
				return taskId, err
			}

			taskId, err = context.execInstance(req, parent_instance_id, point, userId)
			if err != nil {
				//log.Println(err)
				log.Println("execInstance 9 " + err.Error())
				debug.PrintStack()
				return taskId, err
			}
			return taskId, err
		}

		//		err = closeTaskByInstancePoint(instanceId,point)
		//		if err!=nil{
		//			return 0,err
		//		}

		err = context.closeTask(taskId, point, userId)
		if err != nil {
			log.Println("closetask endevent error " + err.Error())
			return taskId, err
		}

		return taskId, nil
	}

	next_point, err := context.getNextPoint(req, instanceId, point, userId)
	if err != nil {
		//debug.PrintStack()
		log.Println("getNextPoint err " + err.Error())
		return "", err
	}
	context.closeTask(taskId, point, userId)

	taskId, err = context.execInstance(req, instanceId, next_point, userId)

	if err != nil {
		log.Println("execInstance err " + err.Error())
		//debug.PrintStack()
		return "", err
	}
	return taskId, nil
}

func (context *InstanceContext) GetInstanceIDUUIDAndPointByTask(task string) (int64, string, int64, error) {

	instanceUUID := ""
	instanceId := int64(0)
	pointId := int64(0)
	err := context.O.Raw(utils.DbBindReplace(`select i.sys$uuid, t.instance_id, t.point_id from bp_tasks t join bp_instances i on i.id=t.instance_id where t.sys$uuid=?`), task).QueryRow(&instanceUUID, &instanceId, &pointId)
	if err != nil {
		log.Println("getInstanceAndPointByTask err ", task, err.Error())
	}
	return instanceId, instanceUUID, pointId, err
}

func (context *InstanceContext) GetInstanceAndPointByTask(task string) (int64, int64, error) {

	instanceId := int64(0)
	pointId := int64(0)
	err := context.O.Raw(utils.DbBindReplace(`select t.instance_id,t.point_id from bp_tasks t where t.sys$uuid=?`), task).QueryRow(&instanceId, &pointId)
	if err != nil {
		log.Println("getInstanceAndPointByTask err ", task, err.Error())
	}
	return instanceId, pointId, err
}

func (context *InstanceContext) GetPointByTask(task string) (int64, error) {

	point := int64(0)
	err := context.O.Raw(utils.DbBindReplace(`select t.point_id from bp_tasks t
				  where t.sys$uuid=?`), task).QueryRow(&point)
	if err != nil {
		log.Println("getPointByTask err " + err.Error() + "task = " + task)
	}
	return point, err
}

func (context *InstanceContext) getPointTypeCodeByPointId(point int64) (string, error) {

	var pointTypes []string
	_, err := cached.O().Raw(utils.DbBindReplace(`select pt.code from    bp_points p
										 join bp_point_types pt on p.type_id=pt.id
										 where p.id=?`), point).QueryRows(&pointTypes)
	if err != nil {
		log.Println("getPointTypeCodeByPointId err  (1) task=", point, err)
		return "", err

	}
	if len(pointTypes) == 0 {
		log.Println("================getPointTypeByTask err  (1) point=", point, err)
		return "", errors.New("PointType not found by point ")
	}

	return pointTypes[0], err
}

func (context *InstanceContext) getPointType(point int64) (string, error) {

	var pointTypes []string
	_, err := cached.O().Raw(utils.DbBindReplace(`select pt.code from    bp_points p
										join bp_point_types pt on p.type_id=pt.id
				  where p.id=?`), point).QueryRows(&pointTypes)

	if err != nil {
		log.Println("getPointType err  (1) task=", point, err)
		return "", err

	}
	if len(pointTypes) == 0 {
		log.Println("getPointType err  (1) point=", point, err)
		return "", errors.New("getPointType not found by point")
	}

	return pointTypes[0], err
}

func (context *InstanceContext) getPointTypeAndWFStatus(point int64) (string, int64, error) {

	pointType := ""
	status := int64(0)
	err := cached.O().Raw(utils.DbBindReplace(`select pt.code,coalesce(p.wf_status_attr_value,0) from    bp_points p
										join bp_point_types pt on p.type_id=pt.id
				  where p.id=?`), point).QueryRow(&pointType, &status)

	if err != nil {
		log.Println("getPointTypeAndWFStatus err  (1) task=", point, err)
		return "", 0, err

	}

	return pointType, status, err
}

func (context *InstanceContext) getPointTypeByTask(task string) (string, error) {

	pointType := ""
	err := context.O.Raw(utils.DbBindReplace(`select pt.code from bp_tasks t,
										 bp_points p,
										 bp_point_types pt
				  where t.sys$uuid=? and p.id=t.point_id and p.type_id=pt.id`), task).QueryRow(&pointType)
	if err != nil {
		log.Println("getPointTypeByTask err " + err.Error() + "task = " + task)
	}
	return pointType, err
}

//func getPointTypeByInstance(instanceId int64) (string,error) {
//	o := orm.NewOrm()
//	o.Using("default")
//	pointType:=""
//	err := o.Raw(`select pt.code from bp_instances i,
//										 bp_points p,
//										 bp_point_types pt
//				  where i.id=? and p.id=i.point_id and p.type_id=pt.id`,instanceId).QueryRow(&pointType)
//	if err!=nil{
//		log.Println("getPointTypeByInstance err "+err.Error())
//	}
//	return pointType,err
//}

func (context *InstanceContext) getNextPoint(req *http.Request, instanceId int64, pointId int64, user_id int64) (int64, error) {

	pt, err := context.getPointTypeCodeByPointId(pointId)

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("getNextPoint starting 1: ", instanceId, pointId, user_id)
	}
	if err != nil {
		log.Println("getNextPoint err = " + err.Error())
		//debug.PrintStack()
		return int64(0), err
	}

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("getNextPoint starting 2: ", instanceId, pointId, user_id)
	}

	if pt == "exclusivegateway" {
		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("exclusivegateway processing")
		}
		return context.getNextPointFromExclusiveGateway(req, instanceId, pointId, user_id)
	}

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("getNextPoint starting 3: ", instanceId, pointId, user_id)
	}

	nextPointId := int64(0)

	type tRow struct {
		Id      int64
		PointId int64
		Cond    string
	}

	var rows []tRow

	_, err = cached.O().Raw(utils.DbBindReplace(`select sf1.id as "id",s2.point_id as "point_id",sf1.cond as "cond" from
		bp_point_sfs s1,
		bp_point_sfs s2,
		bp_points p1,
		bp_points p2,
		bp_sequence_flows sf1
			where
		s1.point_id=? and
		s1.sf_id=s2.sf_id and
		s1.is_incoming=0 and
		s2.is_incoming=1
		and s1.point_id=p1.id
		and s2.point_id=p2.id
		and p1.is_active=1
		and p2.is_active=1
		and s1.is_active=1
		and s2.is_active=1
		and sf1.id = s1.sf_id
		and sf1.is_condition=1
		and sf1.cond is not null
	`), pointId).QueryRows(&rows)

	if err != nil {
		log.Println("getNextPoint error 7 next point curr point = ", pointId, err.Error())
		log.Println(pointId)
		return 0, err
	}

	for _, v := range rows {

		b, err := context.LuaCondition(context.O, req, instanceId, pointId, user_id, v.Id, v.Cond)
		//log.Println("Run condition", b, v)
		if err != nil {
			log.Println("Run condition Error ", err)
		}
		if b {
			return v.PointId, nil
		}

	}

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("getNextPoint starting 4: ", instanceId, pointId, user_id)
	}

	var nextPointIds []int64
	_, err = cached.O().Raw(utils.DbBindReplace(`select s2.point_id from
		bp_point_sfs s1,
		bp_point_sfs s2,
		bp_points p1,
		bp_points p2,
		bp_sequence_flows sf1
			where
		s1.point_id=? and
		s1.sf_id=s2.sf_id and
		s1.is_incoming=0 and
		s2.is_incoming=1
		and s1.point_id=p1.id
		and s2.point_id=p2.id
		and p1.is_active=1
		and p2.is_active=1
		and s1.is_active=1
		and s2.is_active=1
		and sf1.id = s1.sf_id
		and (sf1.is_condition=0 or sf1.is_condition is null)

	`), pointId).QueryRows(&nextPointIds)

	if err != nil {
		log.Println("getNextPoint error net next point curr point = ", pointId, err.Error())
		log.Println(pointId)
	}
	if len(nextPointIds) == 0 {
		return int64(0), errors.New(fmt.Sprintf("%s Point deleted", pointId))
	}

	nextPointId = nextPointIds[0]

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("getNextPoint starting 5: ", instanceId, pointId, user_id, nextPointId)
	}

	return nextPointId, err
}

func (context *InstanceContext) findFirstPoint(processId int64) (int64, error) {

	var points []int64
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Print("findFirstPoint processId=", processId)
	}
	_, err := cached.O().Raw(utils.DbBindReplace(`select p.id from bp_points p
	join bp_point_types pt on pt.id = p.type_id
		where p.process_id=?
      		and p.is_active=1
      		and pt.code='startevent'

			 `), processId).QueryRows(&points)
	if err != nil {

		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("findFirstPoint. No next task found " + err.Error())
		}
		return int64(0), err
	}
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Print("findFirstPoint pointId=", points)
	}
	//log.Println(pointId)
	return points[0], nil

}

func (context *InstanceContext) terminateInstance(instanceId int64) error {

	_, err := context.O.Raw(utils.DbBindReplace("update bp_instance_tasks set is_terminated = 1 where instance_id=?"), instanceId).Exec()
	_, err = context.O.Raw(utils.DbBindReplace("update bp_instances set is_terminated = 1 where id=?"), instanceId).Exec()
	return err
}
