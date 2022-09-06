package restapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

type WidgetSql struct {
	Sqltext string `json:"sqltext"`
	Code    string `json:"code"`
}

func WidgetRestApiGetWidgetTemplate(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	//	if RestCheckAuth(res,req){
	//		return
	//	}

	req.ParseForm()
	o := orm.NewOrm()
	o.Using("default")

	eTag := "NONE"
	o.Raw(utils.DbBindReplace("select substr(md5(concat(max(updated_at),'')),1,5) from widgets")).QueryRow(&eTag)
	versionNum := utils.GetParamValue("version_num")
	eTag = eTag + versionNum

	res.Header().Set("Etag", `"`+eTag+`"`)
	max_age := utils.GetParamValue("widget-max-age")
	if max_age == "" {
		max_age = "0"
	}
	res.Header().Set("Cache-Control", "max-age="+max_age)

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, eTag) {
			res.WriteHeader(http.StatusNotModified)
			return
		}
	}

	s := ""
	err := o.Raw(utils.DbBindReplace("SELECT p.template FROM widgets p where p.id=?"), req.Form.Get("id")).QueryRow(&s)
	RestCheckDBPanic(err, res, o)
	fmt.Fprint(res, s)
	o = nil
	err = nil

}

func WidgetRestApiGetWidgetCode(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	//	if RestCheckAuth(res,req){
	//		return
	//	}

	req.ParseForm()
	o := orm.NewOrm()
	o.Using("default")

	eTag := "NONE"
	o.Raw(utils.DbBindReplace("select substr(md5(max(updated_at)),1,5) from widgets")).QueryRow(&eTag)

	versionNum := utils.GetParamValue("version_num")
	eTag = eTag + versionNum

	res.Header().Set("Etag", `"`+eTag+`"`)
	max_age := utils.GetParamValue("widget-max-age")
	if max_age == "" {
		max_age = "0"
	}
	res.Header().Set("Cache-Control", "max-age="+max_age)

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, eTag) {
			res.WriteHeader(http.StatusNotModified)
			return
		}
	}

	s := ""
	id := 0
	wc := ""
	err := o.Raw(utils.DbBindReplace("SELECT code,id,p.template FROM widgets p where p.code=?"), param.ByName("code")).QueryRow(&wc, &id, &s)

	if RestCheckDBPanic(err, res, o) {
		return
	}
	s = strings.Replace(s, "__WIDGET_URL__", "#/settings/widgetdetails/"+strconv.Itoa(id)+"&version_num="+versionNum, -1)
	s = strings.Replace(s, "__WIDGET_CODE__", wc, -1)

	fmt.Fprint(res, s)
	o = nil
	err = nil

}
