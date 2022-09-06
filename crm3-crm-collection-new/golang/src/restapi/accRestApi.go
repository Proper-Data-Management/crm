package restapi
import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"encoding/json"
	"fmt"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"strconv"
)


func AccUndo(res http.ResponseWriter, req *http.Request, params httprouter.Params) {

	res.Header().Add("Content-Type","application/json")
	type TAccUndoRes struct {
		Ok bool `json:"ok"`
		ErrorText string `json:"errorText"`
	}

	var response TAccUndoRes



	o := orm.NewOrm()
	o.Using("default")

	id,err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		resP, _ := json.Marshal(response)
		fmt.Fprint(res, string(resP))
		return
	}

	err = utils.AccUndo(o,int64(id))
	if err != nil {
		response.Ok = false
		response.ErrorText = err.Error()
		resP, _ := json.Marshal(response)
		fmt.Fprint(res, string(resP))
		return

	}else {
		response.Ok = true
	}

	resP, _ := json.Marshal(response)
	fmt.Fprint(res, string(resP))
}

func AccMove(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type","application/json")
	type TAccMoveRes struct {
		Ok bool `json:"ok"`
		MoveId int64 `json:"moveId"`
		ErrorText string `json:"errorText"`
	}

	var request utils.TAccMoves
	decoder := json.NewDecoder(req.Body)
	err:= decoder.Decode(&request)
	var response TAccMoveRes

	o := orm.NewOrm()
	o.Using("default")

	if err!=nil{
		response.Ok = false
		response.ErrorText = err.Error()

	}else {
		moveId,err := utils.AccMove(o,request,true,"")
		if err != nil {
			response.Ok = false
			response.ErrorText = err.Error()
		}else {
			response.Ok = true
			response.MoveId = moveId
		}
	}
	resP, _ := json.Marshal(response)
	fmt.Fprint(res, string(resP))
}


func AccClsPublish(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {


	type TAccClsPublishRequest struct {
		Id int64 `json:"id"`
	}

	type TAccClsPublishResponse struct {
		Ok bool `json:"ok"`
		ErrorText string `json:"errorText"`
	}
	var request TAccClsPublishRequest
	var response TAccClsPublishResponse


	o:=orm.NewOrm()
	o.Using("default")


	if utils.GetRoleParamValue(o, utils.UserId(req),"generate_ddl")!="1"{
		response.ErrorText = "GenerateDDL. Access denied"
		resP,_ := json.Marshal(response)
		fmt.Fprint(res,string(resP))
	}

	decoder := json.NewDecoder(req.Body)
	err:= decoder.Decode(&request)

	if err!=nil {
		response.Ok = false
		response.ErrorText = err.Error()
		resP,_ := json.Marshal(response)
		fmt.Fprint(res,string(resP))
		return
	}	else {

		err = utils.AccClsPublish(request.Id)

		response.Ok = err==nil
		if err!=nil {
			response.ErrorText = err.Error()
			resP,_ := json.Marshal(response)
			fmt.Fprint(res,string(resP))
			return
		}
		resP,_ := json.Marshal(response)
		fmt.Fprint(res,string(resP))

	}

}

/*
func AccPostByPk(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type","application/json")

	o:=orm.NewOrm()
	o.Using("default")
	move_id,err := utils.AccPostByPkOper(1,"posting.erp_income")
	if err!=nil{
		fmt.Fprintf(res,`{"error":"%v"}`,err.Error())
	}else {
		fmt.Fprintf(res, `{"id":%v}`,move_id)
	}
}

*/