package restapi

import (
	"net/http"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

func DynamicPage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	code := "/" + strings.TrimLeft(req.RequestURI, "/page")

	o := orm.NewOrm()
	o.Using("default")
	template := ""
	//log.Println(code)
	//log.Println(req.RequestURI)

	o.Raw(utils.DbBindReplace("select template from pages where url=? and page_type_id=(select id from page_types where code='dynamic')"), code).QueryRow(&template)
	res.Write([]byte(template))

}
