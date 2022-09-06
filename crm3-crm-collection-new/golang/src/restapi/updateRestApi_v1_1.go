package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	"github.com/julienschmidt/httprouter"
)

func UpdateRestApi_v_1_1(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if RestCheckAuth(res, req) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	t := luautils.DMLRequest{}
	err := decoder.Decode(&t)
	if RestCheckPanic(err, res) {
		fmt.Println("ERROR NAH")
		return
	}

	o := orm.NewOrm()
	o.Using("default")
	o.Begin()

	//Запускаем триггеры только в конце всех обновлений
	if err == nil {
		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("TriggerSyncBefore")
		}
		_, err = luautils.TriggerSyncBefore(o, utils.UserId(req), t)
	}

	if err != nil {
		o.Rollback()
		log.Println("ERROR ON DML5_1 " + err.Error())

	}
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("DML DO")
	}

	resP, err := luautils.DML(o, utils.UserId(req), t, false)

	if err != nil {
		o.Rollback()
		log.Println("ERROR ON DML5 " + err.Error())

	}

	//Запускаем триггеры только в конце всех обновлений
	if err == nil {
		resP, err = luautils.TriggerSync(o, utils.UserId(req), t, resP)
	}

	if err != nil {
		o.Rollback()
		log.Println("ERROR ON DML6 " + err.Error())
		resP.ErrorCode = "1"
		resP.Error = 1

	}
	if err == nil {
		o.Commit()
	}

	jsonData, err := json.Marshal(resP)
	fmt.Fprint(res, string(jsonData))

	utils.ClearInterface(&resP)
	utils.ClearInterface(&jsonData)
	utils.ClearInterface(&t)

}
