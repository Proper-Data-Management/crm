package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

//Надо удалить и перенести в restservices/run. Уязвимость. Hacking
func ListRestApiGetSimpleList(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	userID := utils.UserId(req)
	if userID == 0 {
		return
	}
	if RestCheckAuth(res, req) {
		return
	}

	delim := utils.GetDbStringDelimiter()

	o := orm.NewOrm()
	o.Using("default")
	req.ParseForm()

	// Проверка SQL-инъекции
	if !utils.CheckEntity(req.Form.Get("code")) {
		return
	}
	lang := utils.GetLanguage2(req)
	notEmpty := utils.TranslateTo("Not Empty", lang)
	empty := utils.TranslateTo("Empty", lang)
	iAm := utils.TranslateTo("I'Am", lang)

	entityId := int64(0)
	lookup_expr := ""

	err := o.Raw(utils.DbBindReplace("select id,lookup_expr from entities where code=?"), req.Form.Get("code")).QueryRow(&entityId, &lookup_expr)
	if lookup_expr == "" {
		lookup_expr = "main.title"
	}

	pref := ""
	var params []string

	if req.Form.Get("code") == "users" {
		pref = "select -3 id ,? title, ? title_short union all "
		if utils.GetDbDriverType() == orm.DROracle {
			pref = "select -3 id ,? title, ? title_short from dual union all "
		}
		params = append(params, "<i>"+iAm+"</i>", iAm)
	}

	pref2 := ""
	if req.Form.Get("contains") == "" {
		pref2 = " select -2 id ,? title ,? title_short union all  select -1 id ,? title ,? title_short union all "
		if utils.GetDbDriverType() == orm.DROracle {
			pref2 = " select -2 id ,? title ,? title_short from dual union all  select -1 id ,? title ,? title_short from dual union all "
		}
		params = append(params, "<i>"+notEmpty+"</i>", notEmpty)
		params = append(params, "<i>"+empty+"</i>", empty)
	}

	contains := "%" + req.Form.Get("contains") + "%"
	//fmt.Println("contains",contains)
	containsArr := strings.Split(contains, " ")
	strings.Join(containsArr, "or")
	cond_lookup := "1 = 1 "

	cond_lookup = utils.QueryGetViewQueryLimit(o, entityId, userID)
	for i := range containsArr {
		cond_lookup = lookup_expr + " like ? " + cond_lookup
		containsArr[i] = "%" + containsArr[i] + "%"
	}

	var result []orm.Params

	sql := pref + pref2 + " select main.id," + lookup_expr + " as title,substr(" + lookup_expr + ",1,85) title_short from " + delim + req.Form.Get("code") + delim + " main where " + cond_lookup + " limit 100"

	if utils.GetDbDriverType() == orm.DROracle {
		sql = pref + pref2 + " select main.id," + lookup_expr + " as title,substr(" + lookup_expr + ",1,85) title_short from " + delim + req.Form.Get("code") + delim + " main where " + cond_lookup + " and rownum <= 100"

	}

	_, err = o.Raw(utils.DbBindReplace(sql), params, containsArr).Values(&result)
	//sql,err := utils.SqlRows2Json(pref +"select -2 id ,? title ,? title_short union all  select -1 id ,? title ,? title_short union all select id,title,substr(title,1,30) title_short from `"+req.Form.Get("code")+"`",params)
	if err != nil {
		fmt.Println("ListRestApiGetSimpleList SQL = ", sql)
		RestCheckPanic(err, res)
		return
	}
	//fmt.Println("sql",sql)
	b, err := json.Marshal(result)
	fmt.Fprint(res, string(b))
	utils.ClearInterface(&b)
	return

}
