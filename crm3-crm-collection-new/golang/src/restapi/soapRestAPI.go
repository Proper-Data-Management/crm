package restapi

import (
	"bytes"
	"log"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	lua "github.com/Shopify/go-lua"
	"github.com/julienschmidt/httprouter"
)

func BindInputToSOAP(w http.ResponseWriter, r *http.Request, input string, state *lua.State) error {

	var reqArrGet = make(map[string]string)

	for k, v := range r.Form {
		reqArrGet[k] = v[0]
	}

	var reqArrHeader = make(map[string]string)

	for k, v := range r.Header {
		reqArrHeader[k] = v[0]
	}

	var requestMap = make(map[string]interface{})
	requestMap["input"] = input
	requestMap["user_id"] = utils.UserId(r)
	requestMap["anonymous_session_id"] = utils.AnonymousSessionID(w, r)
	requestMap["get"] = reqArrGet
	requestMap["header"] = reqArrHeader
	requestMap["host"] = r.Host

	requestMap["RemoteAddr"] = r.RemoteAddr
	luautils.DeepPush(state, requestMap)
	state.SetGlobal("request")

	return nil
}

func SOAPWSDL(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	o := orm.NewOrm()
	o.Using("default")

	o.Begin()
	defer o.Rollback()

	wsdl := ""
	err := o.Raw("select wsdl from soap where code=?", params.ByName("soap")).QueryRow(&wsdl)
	if err != nil {
		//w.Header[""]
		log.Println("SOAP Error", err)
		return
	}

	w.Header().Add("Content-Type", "text/xml")
	w.Write([]byte(wsdl))

}
func SOAP(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	r.ParseForm()

	_, filled := r.Form["wsdl"]

	if filled {
		SOAPWSDL(w, r, params)
	}

}

func SOAPDo(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	o := orm.NewOrm()
	o.Using("default")

	o.Begin()
	defer o.Commit()

	script := ""

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	err := o.Raw("select script from soap where code=?", params.ByName("soap")).QueryRow(&script)
	if err != nil {
		//w.Header[""]
		log.Println("SOAP Error", err)
		return
	}

	l := lua.NewState()
	lua.OpenLibraries(l)

	err = luautils.RegisterAPI(l, o)
	if err != nil {
		log.Println("SOAP Error", err)
		return
	}

	//err = luautils.RegisterBPMLUaAPI(r, l, o)
	//if err != nil {
	//	log.Println("SOAP Error", err)
	//	return
	//}

	BindInputToSOAP(w, r, body, l)

	if err := lua.DoString(l, script); err != nil {
		o.Commit()
		return
	}

	l.Global("output")
	value, _ := l.ToString(1)

	w.Header().Add("Content-Type", "text/xml")

	w.Write([]byte(value))

	o.Commit()

	//w.Write([]byte(wsdl))

}
