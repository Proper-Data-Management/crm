package restapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lock"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	lua "github.com/Shopify/go-lua"
	"github.com/julienschmidt/httprouter"
)

type ServiceRunRequest struct {
	Input              interface{} `json:"input"`
	UserId             int64
	NtlmDomain         string
	NtlmUser           string
	NtlmHost           string
	AnonymousSessionID string
}

func BindInputVariablesToService(r *http.Request, request ServiceRunRequest, state *lua.State) error {

	o := orm.NewOrm()
	o.Using("default")
	//log.Println("TEST222")
	//log.Println(request)
	r.ParseForm()

	var reqArrGet = make(map[string]string)

	for k, v := range r.Form {
		reqArrGet[k] = v[0]
	}

	var reqArrHeader = make(map[string]string)

	for k, v := range r.Header {
		reqArrHeader[k] = v[0]
	}

	var requestMap = make(map[string]interface{})
	requestMap["input"] = request.Input
	requestMap["user_id"] = request.UserId
	requestMap["anonymous_session_id"] = request.AnonymousSessionID
	requestMap["get"] = reqArrGet
	requestMap["header"] = reqArrHeader
	requestMap["host"] = r.Host

	if r.Form.Get("sys_ntlm_req") == "1" {
		requestMap["NTLM_DOMAIN"] = request.NtlmDomain
		requestMap["NTLM_USER"] = request.NtlmUser
		requestMap["NTLM_HOST"] = request.NtlmHost
	}
	requestMap["RemoteAddr"] = r.RemoteAddr
	luautils.DeepPush(state, requestMap)
	state.SetGlobal("request")

	return nil
}

func BindOutputVariablesToService(state *lua.State) (interface{}, error) {

	o := orm.NewOrm()
	o.Using("default")

	state.Global("output")
	if state.IsTable(1) {
		l, _ := luautils.PullTable(state, 1)
		//l, err := luautils.PullStringTable(state, 1)

		return l, nil
	} else {
		value, _ := state.ToString(1)
		return value, nil
	}

	return "", nil
}

func ServiceRunAddHeaders(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	type tHeaders struct {
		Header string
		Value  string
	}
	var arr []tHeaders
	o := orm.NewOrm()
	o.Using("default")

	if r.Method != "OPTIONS" && r.Method != "GET" && r.Method != "POST" && r.Method != "PUT" && r.Method != "DELETE" {
		log.Println("BAD METHOD")
		return nil
	}
	add_cond := "m_" + strings.ToLower(r.Method)
	_, err := cached.O().Raw(utils.DbBindReplace("select h.header,h.value from rest_services rs join rest_srv_hdrs h on h.service_id=rs.id where rs.code=? and 1="+add_cond), p.ByName("code")).QueryRows(&arr)
	if err != nil {
		return err
	}
	for _, v := range arr {

		//Replace Headers
		for k2, v2 := range r.Header {
			//log.Println("value1", v.Value)
			v.Value = strings.Replace(v.Value, fmt.Sprintf("{{%s}}", k2), fmt.Sprintf("%s", v2[0]), -1)
			//log.Println("tttt", fmt.Sprintf("{{%s}}", k2), fmt.Sprintf("%s", v2[0]))
			//log.Println("value2", v.Value)
		}
		w.Header().Add(v.Header, v.Value)
	}
	return nil

}
func RunLuaServiceScript(w http.ResponseWriter, r *http.Request, request ServiceRunRequest, script string, is_redirect_output int) (interface{}, error) {

	l := lua.NewState()
	lua.OpenLibraries(l)

	o := orm.NewOrm()
	o.Using("default")
	o.Begin()
	defer o.Commit()

	//loadLuas(l)

	err := luautils.RegisterAPI(l, o)
	if err != nil {
		log.Println("ERROR REGISTER API " + err.Error())
		return nil, err
	}

	luautils.RegisterBPMLUaAPI(r, l, o)

	//log.Println("Startttttt2222")

	err = BindInputVariablesToService(r, request, l)
	if err != nil {
		log.Println("error bind variables  " + err.Error())
		debug.PrintStack()
		return "", err
	}

	//Transaction

	if err := lua.DoString(l, script); err != nil {
		o.Commit()
		log.Println("RunLuaServiceScript error lua  " + err.Error())
		log.Println("RunLuaServiceScript script = " + script)
		utils.ErrorWriteUser("BPMLuaScriptError1", "nothing", request.UserId, err)
		l = nil
		return "", err
	}
	o.Commit()
	//Transaction

	output, err := BindOutputVariablesToService(l)

	if err != nil {
		log.Println("error set output var  " + err.Error())
		debug.PrintStack()
		l = nil
		return "", err
	}

	if is_redirect_output == 1 {
		//log.Println("redirect to")
		//log.Println(fmt.Sprintf("%v",output))
		http.Redirect(w, r, output.(string), 301)
	}

	//log.Println("LUA SCRIPT POINT DONE")
	//log.Println(output)
	l = nil
	return output, nil
}

func ServiceRunOptions(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//var input []luautils.NameValue

	///log.Println("options", w.Header())
	//w.Header().Set("Content-Type", "application/json")

	//w.Header().

	ServiceRunAddHeaders(w, r, p)

}

func ServiceRun(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	//var input []luautils.NameValue

	ServiceRunAddHeaders(w, r, p)

	r.ParseForm()

	var request ServiceRunRequest

	if r.Form.Get("sys_ntlm_req") == "1" {
		NtlmDomain, NtlmUser, NtlmHost, err1 := utils.NTLMProcess(w, r, p)
		if err1 != nil {
			return
		}

		request.NtlmDomain = NtlmDomain
		request.NtlmUser = NtlmUser
		request.NtlmHost = NtlmHost
	}

	//w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	//decoder := json.NewDecoder(r.Body)

	type serviceRunResponse struct {
		Output     interface{} `json:"output"`
		ResultText string      `json:"resultText"`
		ResultCode string      `json:"resultCode"`
		HasError   bool        `json:"hasError"`
	}

	var response serviceRunResponse

	bodyBytes, err2 := ioutil.ReadAll(r.Body)
	//bodyString := ""
	if err2 == nil {
		bodyString := string(bodyBytes)
		err2 = json.Unmarshal([]byte(bodyString), &request.Input)

		if os.Getenv("CRM_VERBOSE_RESTAPI") == "1" {
			log.Println("restapi body", p, bodyString)
		}
	}

	//err := decoder.Decode(&request.Input)
	//if err != nil {
	//	RestCheckPanic(err, w)
	//}

	request.UserId = utils.UserId(r)
	request.AnonymousSessionID = utils.AnonymousSessionID(w, r)

	//log.Println(request.Input)
	o := orm.NewOrm()
	o.Using("default")
	body := ""
	is_redirect_output := 0
	cut_status_info := 0
	res_rc := ""
	mime := ""
	no_parallel := 0
	err := cached.O().Raw(utils.DbBindReplace(`select res_rc.code as res_rc_code, res_rc.mime, r.body,is_redirect_output,cut_status_info,no_parallel
from rest_services r
join rest_content_types res_rc on res_rc.id=r.res_content_type_id
where r.code=?
 and (( (only_auth=1 and ?<>0 or only_auth=0) and coalesce(is_public,0)=0 ) or is_public=1 or exists (select 1 from role_rest_services rs,user_roles ur
where rs.rest_service_id=r.id and
ur.role_id=rs.role_id and ur.user_id=?) )`), p.ByName("code"), request.UserId, request.UserId).QueryRow(&res_rc, &mime, &body, &is_redirect_output, &cut_status_info, &no_parallel)
	o = nil
	if err != nil {
		response.HasError = true
		response.ResultCode = "2"
		response.ResultText = "Service not found or Access Denied. " + err.Error()
		response.Output = ""
		jsonData, _ := json.Marshal(response)
		fmt.Fprint(w, string(jsonData))
		return
	}

	//log.Println("body="+body)

	r.ParseForm()
	if r.Form.Get("async") == "1" {
		if no_parallel == 1 {
			key := "rest_srv_lock:" + p.ByName("code")
			lockobj := lock.Global()
			lock, ok := lockobj.Acquire(key)
			if !ok {
				response.HasError = true
				response.ResultCode = "7"
				response.ResultText = "Service already started"
				response.Output = ""
				jsonData, _ := json.Marshal(response)
				fmt.Fprint(w, string(jsonData))
				return
			}

			go func() {
				defer lockobj.Release(key, lock)
				RunLuaServiceScript(w, r, request, body, is_redirect_output)
			}()
		} else {
			go RunLuaServiceScript(w, r, request, body, is_redirect_output)
		}

		response.HasError = false
		response.ResultCode = "0"
		response.ResultText = "ASYNC OK"
		response.Output = "ASYNC OK"
		jsonData, _ := json.Marshal(response)
		fmt.Fprint(w, string(jsonData))
		jsonData = jsonData[:0]
		utils.ClearInterface(&response)
		return
	} else {
		if no_parallel == 1 {
			key := "rest_srv_lock:" + p.ByName("code")
			lockobj := lock.Global()
			lock, ok := lockobj.Acquire(key)
			if !ok {
				response.HasError = true
				response.ResultCode = "7"
				response.ResultText = "Service already started"
				response.Output = ""
				jsonData, _ := json.Marshal(response)
				fmt.Fprint(w, string(jsonData))
				return
			}
			defer lockobj.Release(key, lock)
		}
	}

	output, err := RunLuaServiceScript(w, r, request, body, is_redirect_output)
	if err != nil {
		response.HasError = true
		response.ResultCode = "1"
		response.ResultText = err.Error()
		response.Output = output
		jsonData, _ := json.Marshal(response)

		fmt.Fprint(w, string(jsonData))
		jsonData = jsonData[:0]
		utils.ClearInterface(&response)
		return

	}
	response.HasError = false
	response.ResultCode = "0"
	response.ResultText = "Service Ok"
	response.Output = output
	w.Header().Set("Content-type", mime)

	if cut_status_info == 1 {
		jsonData, _ := json.Marshal(response.Output)
		fmt.Fprint(w, string(jsonData))
		utils.ClearInterface(&output)
		return
	} else {

		if res_rc == "json" {
			if reflect.TypeOf(output).Kind() == reflect.Map || reflect.TypeOf(output).Kind() == reflect.Slice {
				jsonData, err := json.Marshal(response)
				if err != nil {
					log.Printf("error = %v", err)
				}
				fmt.Fprint(w, string(jsonData))
				utils.ClearInterface(&output)
				utils.ClearInterface(&jsonData)
				return
			} else {
				fmt.Fprint(w, output)
				utils.ClearInterface(&output)
				return
			}
		} else if res_rc != "json" {
			fmt.Fprint(w, output)
			utils.ClearInterface(&output)
			return
		}

	}

}

func ServiceRunGet(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	//var input []luautils.NameValue

	ServiceRunAddHeaders(w, r, p)

	r.ParseForm()

	var request ServiceRunRequest

	//w.Header().Set("Content-Type", "application/json")
	///w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Form.Get("sys_ntlm_req") == "1" {
		NtlmDomain, NtlmUser, NtlmHost, err1 := utils.NTLMProcess(w, r, p)
		if err1 != nil {
			return
		}

		request.NtlmDomain = NtlmDomain
		request.NtlmUser = NtlmUser
		request.NtlmHost = NtlmHost
	}

	//w.Header().Set("Content-Type","application/json")
	type serviceRunResponse struct {
		Output     interface{} `json:"output"`
		ResultText string      `json:"resultText"`
		ResultCode string      `json:"resultCode"`
		HasError   bool        `json:"hasError"`
	}

	request.Input = ""
	var response serviceRunResponse

	request.UserId = utils.UserId(r)
	request.AnonymousSessionID = utils.AnonymousSessionID(w, r)

	//log.Println(request.Input)
	o := orm.NewOrm()
	o.Using("default")
	body := ""
	is_redirect_output := 0
	cut_status_info := 0
	res_rc := ""
	mime := "application/json"
	no_parallel := 0
	err := cached.O().Raw(utils.DbBindReplace("select res_rc.code as res_rc_code, res_rc.mime, r.body,is_redirect_output,cut_status_info,r.no_parallel from rest_services r,rest_content_types res_rc where r.code=? and res_rc.id=r.res_content_type_id and ( is_public=1 or exists (select 1 from role_rest_services rs,user_roles ur where rs.rest_service_id=r.id and ur.role_id=rs.role_id and ur.user_id=?) )  "), p.ByName("code"), request.UserId).QueryRow(&res_rc, &mime, &body, &is_redirect_output, &cut_status_info, &no_parallel)

	if no_parallel == 1 {
		key := "rest_srv_lock:" + p.ByName("code")
		lockobj := lock.Global()
		lock, ok := lockobj.Acquire(key)
		if !ok {
			response.HasError = true
			response.ResultCode = "7"
			response.ResultText = "Service already started"
			response.Output = ""
			jsonData, _ := json.Marshal(response)
			fmt.Fprint(w, string(jsonData))
			return
		}
		defer lockobj.Release(key, lock)
	}

	o = nil
	if err != nil {
		response.HasError = true
		response.ResultCode = "2"
		response.ResultText = "Service not found or Access Denied. " + err.Error()
		response.Output = ""
		jsonData, _ := json.Marshal(response)
		fmt.Fprint(w, string(jsonData))
		return
	}

	w.Header().Set("Content-type", mime)

	//log.Println("body="+body)
	output, err := RunLuaServiceScript(w, r, request, body, is_redirect_output)
	if err != nil {
		response.HasError = true
		response.ResultCode = "1"
		response.ResultText = err.Error()
		response.Output = output
		jsonData, _ := json.Marshal(response)
		fmt.Fprint(w, string(jsonData))
		return

	}
	//	response.HasError = false
	//	response.ResultCode =  "0"
	//	response.ResultText = "Service Ok"
	//	response.Output = output

	//log.Println(reflect.TypeOf(output).Kind())

	if res_rc == "json" {
		if reflect.TypeOf(output).Kind() == reflect.Map || reflect.TypeOf(output).Kind() == reflect.Slice {
			jsonData, _ := json.Marshal(output)

			fmt.Fprint(w, string(jsonData))
		} else {
			fmt.Fprint(w, output)
		}

	} else if res_rc != "json" {
		fmt.Fprint(w, output)
	}

	//	jsonData, err := json.Marshal(response)
	//	fmt.Fprint(w,string(jsonData))

}
