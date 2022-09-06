package restapi

import (
	"fmt"
	"net/http"
	"strconv"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/json-iterator/go"
)

type DetailSql struct {
	EntityCode            string `json:"entity_code"`
	SqlText               string `json:"sql_text"`
	SqlConditionBuildText string `json:"sql_condition_build_text"`
	Code                  string `json:"code"`
}

type detailGetErrorResponse struct {
	IsMobile bool   `json:"isMobile"`
	Lang     string `json:"lang"`
	Error    string `json:"error"`
}

func DetailRestApi(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if RestCheckAuth(res, req) {
		return
	}

	req.ParseForm()
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	o := orm.NewOrm()
	o.Using("default")

	id, err := strconv.Atoi(req.Form.Get("id"))
	code := req.Form.Get("code")
	if RestCheckDBPanic(err, res, o) {
		return
	}

	int, err := utils.Detail(o, id, code, utils.UserId(req))

	res.Header().Set("Content-Type", "application/json;charset=UTF-8")

	if err == nil {
		//b, _ := ffjson.Marshal(&int)

		
		b, _ := json.Marshal(int)

		s := string(b)
		fmt.Fprint(res, s)
		utils.ClearInterface(&b)
		for i := range int {
			utils.ClearInterface(&i)
		}
		utils.ClearInterface(&int)
	} else {

		v := detailGetErrorResponse{Error: err.Error()}
		//b, _ := ffjson.Marshal(&v)
		b, _ := json.Marshal(v)
		fmt.Fprint(res, string(b))
		utils.ClearInterface(&b)
	}

	utils.ClearInterface(&o)

}
