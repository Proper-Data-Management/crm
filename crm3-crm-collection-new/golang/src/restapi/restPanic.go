package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

type TErrorResponse struct {
	ErrorText string `json:"error_text"`
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`

	Details interface{} `json:"details"`
}

func RestCheckDBPanic(err error, res http.ResponseWriter, o orm.Ormer) bool {
	if err != nil {
		//fmt.Println("error goi")
		errRes := TErrorResponse{Error: "1", ErrorText: err.Error()}
		jsonData, _ := json.Marshal(errRes)
		fmt.Fprint(res, string(jsonData))
		//o.Rollback()
		log.Println(err)
		//debug.PrintStack()
		//panic(err)
		//o = nil
		return true
	}
	return false
}

func RestCheckDBPanicDetails(err error, res http.ResponseWriter, errCode, errText string, o orm.Ormer) bool {
	if err != nil {
		//fmt.Println("error goi")
		errRes := TErrorResponse{Error: "1", ErrorCode: errCode, ErrorText: errText}
		jsonData, _ := json.Marshal(errRes)
		fmt.Fprint(res, string(jsonData))
		//o.Rollback()
		log.Println(err)
		//debug.PrintStack()
		//panic(err)
		//o = nil
		return true
	}
	return false
}

func RestCheckAuth(res http.ResponseWriter, req *http.Request) bool {
	if utils.UserId(req) == 0 {
		//fmt.Println("error goi")
		errRes := TErrorResponse{Error: "2", ErrorText: "NEED AUTH"}
		jsonData, _ := json.Marshal(errRes)
		fmt.Fprint(res, string(jsonData))
		return true
	}
	return false
}

func WriteRestCheckPanic(errText, errCode string, res http.ResponseWriter) {
	errRes := TErrorResponse{Error: errCode, ErrorText: errText}
	jsonData, _ := json.Marshal(errRes)
	fmt.Fprint(res, string(jsonData))
}

func RestCheckPanic(err error, res http.ResponseWriter) bool {
	if err != nil {
		//fmt.Println("error goi")
		errRes := TErrorResponse{Error: "1", ErrorText: err.Error()}
		jsonData, _ := json.Marshal(errRes)
		fmt.Fprint(res, string(jsonData))
		log.Println(err)
		//debug.PrintStack()
		//panic(err)
		return true
	}
	return false
}

func CheckPanic(err error) bool {
	if err != nil {
		//fmt.Println("error goi")
		log.Println(err)
		//panic(err)
		return true
	}
	return false
}
