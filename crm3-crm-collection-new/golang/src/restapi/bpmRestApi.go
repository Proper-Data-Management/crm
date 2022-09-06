package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	lua "github.com/Shopify/go-lua"
	"github.com/julienschmidt/httprouter"

	"errors"
)

func BPMPublish(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	type bpmPublishRequest struct {
		ProcessId int64 `json:"processId"`
	}

	type bpmPublishResponse struct {
		Ok        bool   `json:"ok"`
		ErrorText string `json:"errorText"`
	}

	var request bpmPublishRequest
	var response bpmPublishResponse

	context := utils.BpmGenContext{}
	context.O = orm.NewOrm()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)

	response.Ok = true
	response.ErrorText = ""

	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	}

	err = context.Publish(request.ProcessId)

	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
	}

	resP, _ := json.Marshal(response)
	fmt.Fprint(res, string(resP))

}

func BPMSStart(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Set("Content-Type", "application/json;charset=UTF-8")

	type createInstanceRequest struct {
		ProcessId   int64                `json:"processId"`
		ProcessCode string               `json:"processCode"`
		Init        []luautils.NameValue `json:"init"`
	}

	type createInstanceResponse struct {
		Ok        bool                 `json:"ok"`
		Instance  string               `json:"instance"`
		Task      string               `json:"task"`
		ErrorText string               `json:"errorText"`
		Output    []luautils.NameValue `json:"output"`
	}
	var request createInstanceRequest
	var response createInstanceResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	o := orm.NewOrm()
	o.Using("default")
	context := luautils.InstanceContext{}
	context.O = o

	context.Lua = lua.NewState()
	lua.OpenLibraries(context.Lua)
	luautils.RegisterAPI(context.Lua, o)
	luautils.RegisterBPMLUaAPI(req, context.Lua, o)

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)
	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		if request.ProcessCode != "" {
			request.ProcessId, err = context.GetProcessIdByProcessCode(request.ProcessCode)
			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				//return
			}
		}

		o := orm.NewOrm()
		o.Using("default")

		instanceContext := luautils.InstanceContext{O: o}
		instanceContext.InstanceVars = make(map[int64][]luautils.NameValue, 0)
		instanceContext.InstanceTables = make(map[int64]string, 0)
		response.Output, response.Instance, response.Task, _, err = instanceContext.CreateInstance(req, request.ProcessId, utils.UserId(req), request.Init, 0)

		if err != nil {
			response.Ok = false
			response.ErrorText = err.Error()
			//return
		} else {

			response.Ok = true
			//			for _,element := range request.Init {
			//				log.Println("try to set global variable name "+element.Name)
			//				log.Println("try to set global variable value "+element.Value)
			//				err = luautils.SetGlobalVarByInstance(response.Instance,element.Name,element.Value)
			//				if err != nil{
			//					response.ErrorText = err.Error()
			//					response.Ok = false
			//				}
			//			}

		}
	}
	resP, _ := json.Marshal(response)
	fmt.Fprint(res, string(resP))
	resP = nil

}

func BPMCreateInstanceOptions(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	//res.Header().Add("Access-Control-Allow-Methods", "POST")
	//res.Header().Add("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With, Content-Type")
	//res.Header().Set("Access-Control-Allow-Origin", "*")
}

func BPMStartProcessOptions(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	//res.Header().Add("Access-Control-Allow-Methods", "POST")
	//res.Header().Add("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With, Content-Type")
	//res.Header().Set("Access-Control-Allow-Origin", "*")
}
func BPMStartProcess(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	//res.Header().Set("Access-Control-Allow-Origin", "*")
	//res.Header().Set("Content-Type", "application/json;charset=UTF-8")

	req.ParseForm()
	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	type startProcessRequest struct {
		ProcessId   int64                  `json:"processId"`
		ProcessCode string                 `json:"processCode"`
		Input       map[string]interface{} `json:"input"`
	}

	type startProcessResponse struct {
		Ok                 bool                   `json:"ok"`
		Instance           string                 `json:"instance"`
		Task               string                 `json:"task"`
		ErrorText          string                 `json:"errorText"`
		Output             map[string]interface{} `json:"output"`
		outputNameValue    []luautils.NameValue
		InstanceIsFinished bool `json:"instanceIsFinished"`
	}
	var request startProcessRequest
	var response startProcessResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)
	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		if request.ProcessCode != "" {
			request.ProcessId, err = instanceContext.GetProcessIdByProcessCode(request.ProcessCode)

			//log.Println("test555")
			//log.Println(request.ProcessId)
			if err != nil {
				err = errors.New("ERROR E-BPM-00023. Business Process " + request.ProcessCode + " Not found " + err.Error())
				log.Println(err.Error)
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				resP = nil

				return
			}
		}
		var inputNameValue []luautils.NameValue
		for k, v := range request.Input {
			inputNameValue = append(inputNameValue, luautils.NameValue{Name: k, Value: v})
		}

		o := orm.NewOrm()
		o.Using("default")

		cnt1 := 0
		err = o.Raw(utils.DbBindReplace(`
		select count(1) from
		bp_process_roles pr 
		join user_roles ur on ur.role_id=pr.role_id 
		where pr.process_id=? and 
		pr.can_run=1 and ur.user_id=?
		`), request.ProcessId, utils.UserId(req)).QueryRow(&cnt1)

		if err != nil {
			response.Ok = false
			response.ErrorText = "Error on Check Grant 1 " + err.Error()
			resP, _ := json.Marshal(response)
			fmt.Fprint(res, string(resP))
			resP = nil

			return
		}

		cnt2 := 0
		err = o.Raw(utils.DbBindReplace(`
		select count(1) from
		bp_processes
		where id=? and is_public=1
		`), request.ProcessId).QueryRow(&cnt2)

		if err != nil {

			response.Ok = false
			response.ErrorText = "Error on Check Grant 2 " + err.Error()
			resP, _ := json.Marshal(response)
			fmt.Fprint(res, string(resP))
			resP = nil

			return
		}
		//log.Println("cnt1", cnt1, request.ProcessId, utils.UserId(req))
		//log.Println("cnt2", cnt2)

		if cnt1+cnt2 == 0 {
			response.Ok = false
			response.ErrorText = "Access Denied"
			resP, _ := json.Marshal(response)
			fmt.Fprint(res, string(resP))
			resP = nil
			return
		}

		defer o.Rollback() //Если вдруг вылетит внутренний exception
		o.Begin()
		instanceContext := luautils.InstanceContext{O: o}
		instanceContext.InstanceVars = make(map[int64][]luautils.NameValue)
		instanceContext.InstanceTables = make(map[int64]string, 0)

		instanceContext.Lua = lua.NewState()
		lua.OpenLibraries(instanceContext.Lua)
		luautils.RegisterAPI(instanceContext.Lua, o)
		luautils.RegisterBPMLUaAPI(req, instanceContext.Lua, o)

		response.outputNameValue, response.Instance, response.Task, _, err = instanceContext.CreateInstance(req, request.ProcessId, utils.UserId(req), inputNameValue, 0)

		response.InstanceIsFinished = instanceContext.GetInstanceFinishedByUUID(response.Instance)
		response.Output = make(map[string]interface{})
		for _, v := range response.outputNameValue {
			response.Output[v.Name] = v.Value
		}

		if err != nil {
			log.Println("SSSSSSSS", instanceContext.ErrorJson)
			o.Rollback()
			response.Ok = false
			response.ErrorText = err.Error()
			resP, _ := json.Marshal(response)
			fmt.Fprint(res, string(resP))
			resP = nil

			return
		} else {
			o.Commit()
			response.Ok = true
			resP, _ := json.Marshal(response)
			fmt.Fprint(res, string(resP))
			resP = nil
			return
		}
	}

}

func BPMCreateInstance(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Content-Type", "application/json")

	type createInstanceRequest struct {
		ProcessId   int64                `json:"processId"`
		ProcessCode string               `json:"processCode"`
		Init        []luautils.NameValue `json:"init"`
	}

	type createInstanceResponse struct {
		Ok        bool                 `json:"ok"`
		Instance  string               `json:"instance"`
		Task      string               `json:"task"`
		ErrorText string               `json:"errorText"`
		Output    []luautils.NameValue `json:"output"`
	}
	var request createInstanceRequest
	var response createInstanceResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///
	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)
	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		if request.ProcessCode != "" {
			request.ProcessId, err = instanceContext.GetProcessIdByProcessCode(request.ProcessCode)
			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				//return
			}
		}
		o := orm.NewOrm()
		o.Using("default")
		instanceContext := luautils.InstanceContext{O: o}
		defer instanceContext.O.Rollback()
		instanceContext.O.Begin()
		response.Output, response.Instance, response.Task, _, err = instanceContext.CreateInstance(req, request.ProcessId, utils.UserId(req), request.Init, 0)
		if err != nil {
			instanceContext.O.Rollback()
		} else {
			instanceContext.O.Commit()
		}

		if err != nil {
			response.Ok = false
			response.ErrorText = err.Error()
			//return
		} else {

			response.Ok = true
			//			for _,element := range request.Init {
			//				log.Println("try to set global variable name "+element.Name)
			//				log.Println("try to set global variable value "+element.Value)
			//				err = luautils.SetGlobalVarByInstance(response.Instance,element.Name,element.Value)
			//				if err != nil{
			//					response.ErrorText = err.Error()
			//					response.Ok = false
			//				}
			//			}

		}
	}
	resP, _ := json.Marshal(response)
	fmt.Fprint(res, string(resP))
	resP = nil

}

func BPMRunProcess_Deprecated(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	//	{
	//		"processCode": "beton_request",
	//	"findInstanceByUniqueVars": [{"name":"guid", "value": "dfa9dfi90sdfi-i54-0if-0fo-"}],
	//		"manualTaskPointCode": "acceptRequest",
	//	"input":[{"name":"reason", "value": "Успешно принято"}]
	//	}

	type runProcessResponse struct {
		Ok        bool   `json:"ok"`
		ErrorText string `json:"errorText"`
	}

	var request luautils.RunProcessRequest
	var response runProcessResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&request)

	//str := ""
	//err = decoder.Decode(&str)

	out, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	log.Println(string(out))

	instanceId := int64(0)

	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		for _, element := range request.FindInstanceByUniqueVars {
			//luautils.SetGlobalVarByInstance(1,element.Name,element.Value)

			log.Println("try to find Name " + element.Name)
			log.Println("try to find Value " + element.Value.(string))
			instanceId, err = instanceContext.FindInstanceIdByUniqueVar(request.ProcessCode, element.Name, element.Value)
			log.Println("found instanceId " + strconv.Itoa(int(instanceId)))

			if err != nil {
				response.Ok = true //Бывает два раза отправляют. Депрекейтед. Не страшно
				//response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				//log.Println("ERROR E-BPM-00018. Error on try to find Name request.ProcessCode= " +request.ProcessCode + " element.Name = " + element.Name + " element.Value " +  element.Value +" "+err.Error())
				log.Printf("ERROR E-BPM-00018. Error on try to find Name request.ProcessCode= %v element.Name = %v element.Value %v ERROR %v\n", request.ProcessCode, element.Name, element.Value, err.Error())
				return
			}
		}

		//		for _,element:=range request.Input{
		//			luautils.SetGlobalVarByInstance(1,element.Name,element.Value)
		//		}

		if err == nil {

			pointId, err := instanceContext.GetPointIdByAlias(request.ManualTaskPointCode)
			if err != nil {
				response.Ok = false
				response.ErrorText = "Point by Alias not found: " + err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("err on GetPointIdByAlias" + err.Error())
				return
			}

			o := orm.NewOrm()
			o.Using("default")
			err = o.Raw(utils.DbBindReplace("select sys$uuid from bp_tasks where is_open=1 and instance_id=? and point_id=? limit 1"), instanceId, pointId).QueryRow(&request.Task)
			if err != nil {
				response.Ok = true //Бывает два раза отправляют. Депрекейтед. Не страшно
				//response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("err on get task skipped " + err.Error())
				//debug.PrintStack()
				return
			}

			instanceContext := luautils.InstanceContext{O: o}

			_, err = instanceContext.ManualExecInstanceByTask(req, request.Task, utils.UserId(req), request.Input)

			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("error on ManualExecInstanceByTask " + err.Error())
				//debug.PrintStack()
				return
			} else {
				response.Ok = true
				response.ErrorText = ""
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				return

			}
		}
	}

}

func BPMRunProcess(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	//	{
	//		"processCode": "beton_request",
	//	"findInstanceByUniqueVars": [{"name":"guid", "value": "dfa9dfi90sdfi-i54-0if-0fo-"}],
	//		"manualTaskPointCode": "acceptRequest",
	//	"input":[{"name":"reason", "value": "Успешно принято"}]
	//	}

	type runProcessResponse struct {
		Ok        bool   `json:"ok"`
		ErrorText string `json:"errorText"`
		Task      string `json:"task"`
	}

	var request luautils.RunProcessRequest
	var response runProcessResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&request)

	//str := ""
	//err = decoder.Decode(&str)

	//out, err := json.Marshal(request)
	//if err != nil {
	//	panic(err)
	//}

	//log.Println(string(out))

	instanceId := int64(0)

	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		for _, element := range request.FindInstanceByUniqueVars {
			//luautils.SetGlobalVarByInstance(1,element.Name,element.Value)

			log.Println("try to find Name " + element.Name)
			log.Println("try to find Value " + element.Value.(string))
			instanceId, err = instanceContext.FindInstanceIdByUniqueVar(request.ProcessCode, element.Name, element.Value)
			log.Println("found instanceId " + strconv.Itoa(int(instanceId)))

			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println(err)
				debug.PrintStack()
				return
			}
		}

		//		for _,element:=range request.Input{
		//			luautils.SetGlobalVarByInstance(1,element.Name,element.Value)
		//		}

		if err == nil {

			//			pointId,err:=luautils.GetPointIdByAlias(request.ManualTaskPointCode)
			//			if err!=nil{
			//				response.Ok = false
			//				response.ErrorText = "Point by Alias not found: "+err.Error()
			//				resP,_ := json.Marshal(response)
			//				fmt.Fprint(res,string(resP))
			//				log.Println(err)
			//				debug.PrintStack()
			//				return
			//			}

			o := orm.NewOrm()
			o.Using("default")
			instanceContext := luautils.InstanceContext{O: o}

			response.Task, err = instanceContext.ManualExecInstanceByTask(req, request.Task, utils.UserId(req), request.Input)
			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println(err)
				debug.PrintStack()
				return
			} else {
				response.Ok = true
				response.ErrorText = ""
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				return

			}
		}
	}

}

func BPMRunUserTaskByInstanceOptions(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	//res.Header().Add("Access-Control-Allow-Methods", "POST")
	//res.Header().Add("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With, Content-Type")
	//res.Header().Set("Access-Control-Allow-Origin", "*")
}

func BPMRunUserTaskByInstance(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	//res.Header().Set("Access-Control-Allow-Origin", "*")
	//res.Header().Set("Content-Type", "application/json")

	type runProcessResponse struct {
		Ok        bool              `json:"ok"`
		ErrorVars map[string]string `json:"errorVars"`
		ErrorText string            `json:"errorText"`
		Task      string            `json:"task"`
	}

	type RunUserTaskRequestByInstance struct {
		ProcessCode string      `json:"processCode"`
		Instance    int64       `json:"instance"`
		PointCode   string      `json:"pointCode"`
		Input       interface{} `json:"input"`
		Task        string      `json:"task"`
	}

	var request RunUserTaskRequestByInstance
	var response runProcessResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	instanceContext.Lua = lua.NewState()
	lua.OpenLibraries(instanceContext.Lua)
	luautils.RegisterAPI(instanceContext.Lua, o)
	luautils.RegisterBPMLUaAPI(req, instanceContext.Lua, o)

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&request)

	//out, err := json.Marshal(request)
	//if err != nil {
	//	panic(err)
	//}

	//log.Println(string(out))

	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		if err == nil {

			//			pointId,err:=luautils.GetPointIdByCode(request.PointCode)
			//			if err!=nil{
			//				response.Ok = false
			//				response.ErrorText = "Point by Code "+request.PointCode+" not found: "+err.Error()
			//				resP,_ := json.Marshal(response)
			//				fmt.Fprint(res,string(resP))
			//				log.Println(err)
			//				debug.PrintStack()
			//				return
			//			}

			var input = []luautils.NameValue{}
			a := request.Input.(map[string]interface{})
			for index, val := range a {
				//log.Println(index)
				//log.Println(val)
				input = append(input, luautils.NameValue{Name: index, Value: val})
			}

			pointId, err := instanceContext.GetPointByTask(request.Task)

			if err != nil {
				o.Rollback()
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("ERROR 3 " + err.Error())
				return
			}

			response.ErrorVars, err = instanceContext.CheckRqOutputIntPointVars(pointId, input)

			if err != nil {
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("ERROR 4 " + err.Error())
				return
			}

			o := orm.NewOrm()
			o.Using("default")
			instanceContext := luautils.InstanceContext{O: o}

			response.Task, err = instanceContext.ManualExecInstanceByTask(req, request.Task, utils.UserId(req), input)
			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("ERROR " + err.Error())
				debug.PrintStack()
				return
			} else {
				response.Ok = true
				response.ErrorText = ""
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				return

			}
		}
	}

}

func BPMRunUserTaskByTask(res http.ResponseWriter, req *http.Request, params httprouter.Params) {

	type runProcessResponse struct {
		Ok                 bool                   `json:"ok"`
		ErrorVars          map[string]string      `json:"errorVars"`
		ErrorText          string                 `json:"errorText"`
		Task               string                 `json:"task"`
		Output             map[string]interface{} `json:"output"`
		outputNameValue    []luautils.NameValue
		Instance           string `json:"instance"`
		InstanceIsFinished bool   `json:"instanceIsFinished"`
	}

	type RunUserTaskRequestByTask struct {
		Task  string      `json:"task"`
		Input interface{} `json:"input"`
	}

	var request RunUserTaskRequestByTask
	var response runProcessResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	req.ParseForm()
	decoder := json.NewDecoder(req.Body)

	o := orm.NewOrm()
	defer o.Rollback()
	o.Using("default")
	o.Begin()

	instanceContext := luautils.InstanceContext{O: o, Req: req}

	instanceContext.InstanceVars = make(map[int64][]luautils.NameValue)
	instanceContext.InstanceTables = make(map[int64]string, 0)

	instanceContext.Lua = lua.NewState()
	lua.OpenLibraries(instanceContext.Lua)
	luautils.RegisterAPI(instanceContext.Lua, o)
	luautils.RegisterBPMLUaAPI(req, instanceContext.Lua, o)

	err := decoder.Decode(&request)

	//out, err := json.Marshal(request)
	//if err != nil {
	//panic(err)
	//	o.Rollback()
	//	response.Ok = false
	//	response.ErrorText = err.Error()
	//	return
	//}

	//log.Println(string(out))

	if err != nil {
		o.Rollback()
		response.Ok = false
		response.ErrorText = err.Error()
		return
	} else {

		if err == nil {

			var input = []luautils.NameValue{}
			if request.Input != nil {
				a := request.Input.(map[string]interface{})
				for index, val := range a {
					//log.Println(index)
					//log.Println(val)
					input = append(input, luautils.NameValue{Name: index, Value: val})
				}
			}

			instanceId, instanceUUID, pointId, err := instanceContext.GetInstanceIDUUIDAndPointByTask(request.Task)

			response.Instance = instanceUUID

			if err != nil {
				o.Rollback()
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("ERROR 1 " + err.Error())
				return
			}

			response.ErrorVars, err = instanceContext.CheckRqOutputIntPointVars(pointId, input)

			if err != nil {
				o.Rollback()
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("ERROR 2 " + err.Error())
				return
			}

			response.Task, err = instanceContext.ManualExecInstanceByTask(req, request.Task, utils.UserId(req), input)

			//TODO
			response.outputNameValue = instanceContext.InstanceVars[instanceId]
			if err != nil {
				o.Rollback()
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println("ERROR " + err.Error())
				debug.PrintStack()
				return
			} else {
				if response.outputNameValue != nil {
					response.Output = make(map[string]interface{})
					for _, v := range response.outputNameValue {
						response.Output[v.Name] = v.Value
					}
				}

				response.Ok = true
				response.InstanceIsFinished = instanceContext.GetInstanceFinishedByUUID(response.Instance)
				response.ErrorText = ""
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				o.Commit()
				return

			}
		}
	}

}

func BPMRunUserTaskByTaskUUIDOptions(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	//res.Header().Add("Access-Control-Allow-Methods", "POST")
	//res.Header().Add("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With, Content-Type")
	//res.Header().Set("Access-Control-Allow-Origin", "*")

}

func BPMRunUserTask(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	//	{
	//		"processCode": "beton_request",
	//	"findInstanceByUniqueVars": [{"name":"guid", "value": "dfa9dfi90sdfi-i54-0if-0fo-"}],
	//		"manualTaskPointCode": "acceptRequest",
	//	"input":[{"name":"reason", "value": "Успешно принято"}]
	//	}

	type runProcessResponse struct {
		Ok        bool   `json:"ok"`
		ErrorText string `json:"errorText"`
		Task      string `json:"task"`
	}

	var request luautils.RunUserTaskRequest
	var response runProcessResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	instanceContext.Lua = lua.NewState()
	lua.OpenLibraries(instanceContext.Lua)
	luautils.RegisterAPI(instanceContext.Lua, o)
	luautils.RegisterBPMLUaAPI(req, instanceContext.Lua, o)

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&request)

	//str := ""
	//err = decoder.Decode(&str)

	//out, err := json.Marshal(request)
	//if err != nil {
	//	panic(err)
	//}

	//log.Println(string(out))

	instanceId := int64(0)

	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		for _, element := range request.FindInstanceByUniqueVars {
			//luautils.SetGlobalVarByInstance(1,element.Name,element.Value)

			log.Println("try to find Name " + element.Name)
			log.Println("try to find Value " + element.Value.(string))
			instanceId, err = instanceContext.FindInstanceIdByUniqueVar(request.ProcessCode, element.Name, element.Value)
			log.Println("found instanceId " + strconv.Itoa(int(instanceId)))

			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println(err)
				debug.PrintStack()
				return
			}
		}

		//		for _,element:=range request.Input{
		//			luautils.SetGlobalVarByInstance(1,element.Name,element.Value)
		//		}

		if err == nil {

			pointId, err := instanceContext.GetPointIdByCode(request.UserTaskPointCode)
			if err != nil {
				response.Ok = false
				response.ErrorText = "Point by Code " + request.UserTaskPointCode + " not found: " + err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println(err)
				debug.PrintStack()
				return
			}

			log.Println("xx instanceId=")
			log.Println(instanceId)
			log.Println("xx pointId=")
			log.Println(pointId)

			o := orm.NewOrm()
			o.Using("default")
			instanceContext := luautils.InstanceContext{O: o}

			response.Task, err = instanceContext.ManualExecInstanceByTask(req, request.Task, utils.UserId(req), request.Input)
			if err != nil {
				response.Ok = false
				response.ErrorText = err.Error()
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				log.Println(err)
				debug.PrintStack()
				return
			} else {
				response.Ok = true
				response.ErrorText = ""
				resP, _ := json.Marshal(response)
				fmt.Fprint(res, string(resP))
				return

			}
		}
	}

}

func BPMManualExecByInstance(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	type execInstanceRequest struct {
		InstanceId int64                `json:"instanceId"`
		Task       string               `json:"task"`
		PointId    int64                `json:"pointId"`
		Input      []luautils.NameValue `json:"input"`
	}

	type execInstanceResponse struct {
		Ok        bool   `json:"ok"`
		ErrorText string `json:"errorText"`
		Task      string `json:"task"`
	}
	var request execInstanceRequest
	var response execInstanceResponse

	////Warning need uncomment///
	//if RestCheckAuth(res, req) {
	//	return
	//}
	////Warning need uncomment///

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)
	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		//return
	} else {

		o := orm.NewOrm()
		o.Using("default")
		instanceContext := luautils.InstanceContext{O: o}

		instanceContext.Lua = lua.NewState()
		lua.OpenLibraries(instanceContext.Lua)
		luautils.RegisterAPI(instanceContext.Lua, o)
		luautils.RegisterBPMLUaAPI(req, instanceContext.Lua, o)

		response.Task, err = instanceContext.ManualExecInstanceByTask(req, request.Task, utils.UserId(req), nil)
		if err != nil {
			response.Ok = false
			response.ErrorText = err.Error()
			debug.PrintStack()
			//return
		} else {
			response.Ok = true
		}
	}
	resP, _ := json.Marshal(response)
	fmt.Fprint(res, string(resP))

}

func getProcessFormByTask(task string, currentUserId int64) (int64, string, int64, int64, string, error) {
	o := orm.NewOrm()
	o.Using("default")
	form := ""
	instance_id := int64(0)
	processId := int64(0)
	processTitle := ""
	form_id := int64(0)
	err := o.Raw(utils.DbBindReplace("select pr.id, pr.title, f.id, t.instance_id,f.template from bp_processes pr,bp_process_forms f,bp_tasks t,bp_points po where pr.id=f.process_id and t.sys$uuid=? and t.point_id=po.id and f.process_id=po.process_id limit 1"), task).QueryRow(&processId, &processTitle, &form_id, &instance_id, &form)

	if err != nil {
		return 0, "", 0, 0, "", err
	} else {

		//if userId != currentUserId{
		//	return instance_id,"AnotherUserTask",nil
		//}
		return processId, processTitle, form_id, instance_id, form, nil
	}

}

func getFormByOpenTask(task string, currentUserId int64, nocheck string) (int64, string, int64, string, error) {
	o := orm.NewOrm()
	o.Using("default")
	form := ""
	instance_id := int64(0)
	processId := int64(0)
	processTitle := ""
	userId := int64(0)
	err := o.Raw(utils.DbBindReplace(`select pr.id, pr.title, user_id,t.instance_id,coalesce(pp.form, p.form) from 
bp_points p
join bp_processes pr on pr.id=p.process_id 
join bp_tasks t on t.point_id=p.id
join bp_point_types pt  on pt.id=p.type_id
left join bp_instances i on i.id=t.instance_id 
left join bp_points pp on pp.id = i.finish_point_id
   where  t.sys$uuid=? and  
    ( (pt.code='usertask' 
    
    and
    (
		( exists (select 1 from bp_task_actors where task_id=t.id and user_id=? )
			or ?='1'
		)
    or pr.is_public = 1
    or exists (select 1 from bp_instances i where i.id=t.instance_id and i.finish_point_id is not null)
    )
    ) 
  )`), task, currentUserId, nocheck).QueryRow(&processId, &processTitle, &userId, &instance_id, &form)
	if err != nil {

		return 0, "", 0, "User Task Not Found", err
	} else {
		return processId, processTitle, instance_id, form, nil
	}

}

func BPMShowUserTaskForm(res http.ResponseWriter, req *http.Request, prm httprouter.Params) {

	type tResponseStep struct {
		IsRefused int    `json:"isRefused"`
		Step      int    `json:"step"`
		StepTitle string `json:"stepTitle"`
	}
	type tResponse struct {
		formId           int64           `json:"formId"`
		form             string          `json:"form"`
		CurrentStep      int             `json:"currentStep"`
		RefuseFinalStep  int             `json:"refuseFinalStep"`
		HideEndEventForm int             `json:"hideEndEventForm"`
		IsOpen           int             `json:"isOpen"`
		InstanceId       int64           `json:"instanceId"`
		PointCode        string          `json:"pointCode"`
		ProcessTitle     string          `json:"processTitle"`
		Vars             interface{}     `json:"vars"`
		Steps            []tResponseStep `json:"steps"`
		Ok               bool            `json:"ok"`
		ProcessId        int64           `json:"processId"`
	}
	req.ParseForm()
	var response tResponse
	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}

	task := prm.ByName("task")
	nocheck := req.Form.Get("nocheck")
	if nocheck == "" {
		nocheck = "0"
	}
	instanceContext.O.Raw(utils.DbBindReplace("select p.step,t.is_open,coalesce(p.refuse_final_step,0),coalesce(p.hide_endevent_form,0) from bp_tasks t,bp_points p where t.point_id=p.id and t.sys$uuid=?"), task).QueryRow(&response.CurrentStep, &response.IsOpen, &response.RefuseFinalStep, &response.HideEndEventForm)

	sql_get_steps := `
	select

	coalesce( (select 1 from bp_tasks t2,bp_points po2
	where t2.point_id=po2.id and t2.instance_id=i.id and coalesce(po2.refuse_final_step,0)=1 and t2.is_open=0 and po.step=po2.step limit 1),0) is_refused,

	coalesce( (select po2.title from bp_tasks t2,bp_points po2
	where t2.point_id=po2.id and t2.instance_id=i.id and coalesce(po2.refuse_final_step,0)=1 and t2.is_open=0 and po.step=po2.step limit 1),po.title) step_title,
	po.step step
	 from bp_points po,bp_instances i,bp_processes pr
	where coalesce(po.refuse_final_step,0)<>1 and po.process_id=i.process_id and pr.id=i.process_id and i.id=? and coalesce(po.title,'')<>'' and coalesce(po.step,0)<>0 order by abs(po.step) `

	//	if err!=nil{
	//		log.Println("error BPMShowUserTaskFormByInstance" +err.Error())
	//		fmt.Fprint(res,"error BPMShowUserTaskFormByInstance"+err.Error())
	//	}
	err := errors.New("")
	response.ProcessId, response.ProcessTitle, response.InstanceId, response.form, err = getFormByOpenTask(task, utils.UserId(req), nocheck)
	if err != nil {

		log.Println("error getFormByOpenTask", err.Error())

		response.ProcessId, response.ProcessTitle, response.formId, response.InstanceId, response.form, err = getProcessFormByTask(task, utils.UserId(req))

		_, err := o.Raw(utils.DbBindReplace(sql_get_steps), response.InstanceId).QueryRows(&response.Steps)
		//log.Println("test")

		if err != nil {
			log.Println("error BPMShowUserTaskFormByInstance" + err.Error())
			response.Ok = false
			b, err := json.Marshal(response)
			if err != nil {
				log.Println("error BPMShowUserTaskFormByInstance" + err.Error())
				response.Ok = false
			}
			fmt.Fprint(res, string(b))
			return
		} else {

			response.Ok = true

			if req.Form.Get("showform") != "" {
				fmt.Fprint(res, response.form)
				return
			}

			response.Vars, err = instanceContext.BindVariablesByFormToVar(response.InstanceId, response.formId)

			_, err := cached.O().Raw(utils.DbBindReplace(sql_get_steps), response.InstanceId).QueryRows(&response.Steps)
			//log.Println("error 222",err)
			//response.Steps,err =

			b, err := json.Marshal(response)
			if err != nil {
				log.Println("error BPMShowUserTaskFormByInstance" + err.Error())
				response.Ok = false
				//fmt.Fprint(res,"error BPMShowUserTaskFormByInstance"+err.Error())
			}

			fmt.Fprint(res, string(b))
			return

		}

		//fmt.Fprint(res,"error BPMShowUserTaskFormByInstance"+err.Error())
	} else {

		response.Ok = true

		if req.Form.Get("showform") != "" {
			fmt.Fprint(res, response.form)
			return
		}

		point, err := instanceContext.GetPointByTask(task)
		v, err := instanceContext.BindInputGlobalVariablesToVar(response.InstanceId, point)
		response.Vars = v

		_, err = o.Raw(utils.DbBindReplace(sql_get_steps), response.InstanceId).QueryRows(&response.Steps)
		//log.Println("error 222", err)

		b, err := json.Marshal(response)
		if err != nil {
			log.Println("error BPMShowUserTaskFormByInstance" + err.Error())
			response.Ok = false
			//fmt.Fprint(res,"error BPMShowUserTaskFormByInstance"+err.Error())
		}

		fmt.Fprint(res, string(b))
		return
	}

}

func BPMShowInstanceForm(res http.ResponseWriter, req *http.Request, prm httprouter.Params) {

	type tResponseStep struct {
		IsRefused int    `json:"isRefused"`
		Step      int    `json:"step"`
		StepTitle string `json:"stepTitle"`
	}
	type tResponse struct {
		formId           int64           `json:"formId"`
		form             string          `json:"form"`
		CurrentStep      int             `json:"currentStep"`
		RefuseFinalStep  int             `json:"refuseFinalStep"`
		HideEndEventForm int             `json:"hideEndEventForm"`
		IsOpen           int             `json:"isOpen"`
		InstanceId       int64           `json:"instanceId"`
		PointCode        string          `json:"pointCode"`
		ProcessTitle     string          `json:"processTitle"`
		Vars             interface{}     `json:"vars"`
		Steps            []tResponseStep `json:"steps"`
		Ok               bool            `json:"ok"`
		ProcessId        int64           `json:"processId"`
	}
	req.ParseForm()
	var response tResponse
	o := orm.NewOrm()
	o.Using("default")
	instanceContext := luautils.InstanceContext{O: o}
	instanceContext.InstanceTables = make(map[int64]string)
	instanceContext.InstanceVars = make(map[int64][]luautils.NameValue)
	instanceContext.Lua = lua.NewState()
	lua.OpenLibraries(instanceContext.Lua)
	luautils.RegisterAPI(instanceContext.Lua, o)
	luautils.RegisterBPMLUaAPI(nil, instanceContext.Lua, o)	

	form := ""

	instance := prm.ByName("instance")

	pointId := int64(0)
	processId := int64(0)
	instanceContext.O.Raw(utils.DbBindReplace("select p.process_id,i.id, p.id, p.form, p.step, coalesce(p.refuse_final_step,0),coalesce(p.hide_endevent_form,0) from bp_instances i join bp_points p on p.id=i.finish_point_id where i.sys$uuid=?"), instance).QueryRow(&processId, &response.InstanceId, &pointId, &form, &response.CurrentStep, &response.RefuseFinalStep, &response.HideEndEventForm)

	err := errors.New("")

	if req.Form.Get("showform") != "" {
		fmt.Fprint(res, form)
		return
	}

	response.IsOpen = 1
	response.Ok = true

	tableName, err := instanceContext.GetProcessTableByProcessId(processId)

	if err != nil {
		log.Println("error BPMShowInstanceFormByInstance 1" + err.Error())
		response.Ok = false
	}

	err = instanceContext.ReadAllVariablesFromDB(response.InstanceId, processId, tableName, int64(0))
	if err != nil {
		log.Println("error BPMShowInstanceFormByInstance 2" + err.Error())
		response.Ok = false
	}
	v, err := instanceContext.BindGlobalOuputVariablesToVar(response.InstanceId, processId)
	response.Vars = v
	b, err := json.Marshal(response)
	if err != nil {
		log.Println("error BPMShowInstanceFormByInstance 3" + err.Error())
		response.Ok = false
	}

	fmt.Fprint(res, string(b))
	return
}

func BPMTableGenerate(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	type TBPMTableGenerateRequest struct {
		BpId int64 `json:"bpId"`
	}

	type TDDLGenerateResponse struct {
		Ok        bool   `json:"ok"`
		ErrorText string `json:"errorText"`
	}
	var request TBPMTableGenerateRequest
	var response TDDLGenerateResponse

	context := utils.BpmGenContext{}
	context.O = orm.NewOrm()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)

	if err != nil {
		response.Ok = false
	} else {
		err = context.BpmTableGenerate(request.BpId)
		response.Ok = err == nil
		if err != nil {
			response.ErrorText = err.Error()
		}
		resP, _ := json.Marshal(response)
		fmt.Fprint(res, string(resP))

	}

}
