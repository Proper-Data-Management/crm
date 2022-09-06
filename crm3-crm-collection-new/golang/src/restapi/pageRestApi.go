package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

type View struct {
	Name        string `json:"name"`
	Templateurl string `json:"templateurl"`
}

type State struct {
	Id          int64    `json:"id"`
	Title       string   `json:"title"`
	Url         string   `json:"url"`
	Templateurl string   `json:"templateurl"`
	DbTemplate  int      `json:"db_template"`
	Controller  string   `json:"controller"`
	Name        string   `json:"name"`
	Views       []View   `json:"views"`
	Files       []string `json:"files"`
}

func PageRestApiGet(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	o := orm.NewOrm()
	o.Using("default")
	eTag := "NONE"
	o.Raw(utils.DbBindReplace("select substr(md5(max(created_at)),1,5) from pages")).QueryRow(&eTag)

	//Отключение кеширования в новой версии страниц
	versionNum := utils.GetParamValue("version_num")
	eTag = eTag + versionNum

	res.Header().Set("Etag", `"`+eTag+`"`)
	max_age := utils.GetParamValue("page-max-age")
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

	//res.Header().Set("Access-Control-Allow-Origin", "*")

	//	if RestCheckAuth(res,req){
	//		//runtime.GC()
	//		return
	//	}

	var arr []map[string]interface{}

	sql := `SELECT

	p.id,p.icon,
	p.entity_id,
	(select code from entities where id=p.entity_id) entity_code,
	(select GROUP_CONCAT(concat(pp.code,"=",pp.value)) from page_params pp where pp.page_id=p.id) _page_params,
	(select q.code from queries q where id=p.query_id) query_code,
	p.filter_set_id,
	p.query_id,
	(select fs.code from filter_sets fs where fs.id=p.filter_set_id) filter_set_code,
	p.title,p.url,p.templateurl,p.db_template,pt.controller,p.is_cache,p.code FROM pages p
	join page_types pt on p.page_type_id=pt.id
	join modules m on m.id=p.module_id
	where m.is_active=1
	`

	if utils.GetDbDriverType() == orm.DROracle {
		sql = `SELECT

		p.id,p.icon,
		p.entity_id,
		(select code from entities where id=p.entity_id) entity_code,
		(SELECT LISTAGG(pp.code||'='||pp.value,',') WITHIN GROUP (ORDER BY pp.code)  FROM page_params pp where pp.page_id=p.id) as "_page_params",
		(select q.code from queries q where id=p.query_id) query_code,
		p.filter_set_id,
		p.query_id,
		(select fs.code from filter_sets fs where fs.id=p.filter_set_id) filter_set_code,
		p.title,p.url,p.templateurl,p.db_template,pt.controller,p.is_cache,p.code FROM pages p
		join page_types pt on p.page_type_id=pt.id
		join modules m on m.id=p.module_id
		where m.is_active=1
		`

	}
	arr, err := utils.SqlRows2Table(sql)

	if err != nil {
		log.Println("PageRestApiGet Error", err)
		return
	}
	//var arrFiles []map[string]interface{}

	for index := range arr {

		// element is the element from someSlice for where we are
		var z []View
		//		log.Println("DB Template")
		//		log.Println(arr[index].DbTemplate)
		//log.Println("arr", arr[index])
		if fmt.Sprintf("%v", arr[index]["db_template"]) == "1" {
			arr[index]["templateurl"] = "../restapi/pagetemplate?id=" + fmt.Sprintf("%v", arr[index]["id"]) + "&version_num=" + versionNum

		}

		/*if arr[index]["code"] != nil && arr[index]["is_cache"] != nil && arr[index]["is_cache"].(string) == "1" && arr[index]["db_template"].(string) == "1" {
			arr[index]["templateurl"] = "/page/" + arr[index]["code"].(string)
		}*/

		z = append(z, View{Templateurl: arr[index]["templateurl"].(string), Name: "state" + fmt.Sprintf("%v", arr[index]["id"])})
		//arr[index]["name"]="state"+arr[index]["id"].(string)
		arr[index]["views"] = z
		//arrFiles := []map[string]interface{}
		//arrFiles,_ := utils.SqlRows2Table("SELECT j.url FROM `j_s_plugins` j,`page_types` pt,`pages` p,`page_type_js` ptj where p.id="+arr[index]["id"].(string)+" and p.page_type_id=pt.id and ptj.js_id=j.id and ptj.page_type_id=pt.id")
		//arrFiles,_ := utils.SqlRows2Table("Select 1 as test")

		//arrFiles,_ =  utils.SqlRows2Table("SELECT p.id,p.title,p.url,p.templateurl,p.db_template,pt.controller FROM pages p, page_types pt where p.page_type_id=pt.id")
		if arr[index]["_files"] != nil {
			arr[index]["files"] = strings.Split(arr[index]["_files"].(string), ",")
		}
		pageP := make(map[string]string)
		if arr[index]["_page_params"] != nil {
			pageParams := strings.Split(arr[index]["_page_params"].(string), ",")
			for _, v := range pageParams {
				s := strings.Split(v, "=")
				if len(s) > 0 {
					pageP[s[0]] = s[1]
				}

			}

			arr[index]["params"] = pageP
		}
		arr[index]["id"] = fmt.Sprintf("%v", arr[index]["id"])
		//arrFiles = nil
		z = nil

	}

	//jsonData, err := json.MarshalIndent(arr," "," ")

	b := new(bytes.Buffer)
	var e = json.NewEncoder(b)
	err = e.Encode(arr)

	//log.Println(string(b.Bytes()))
	//utils.ClearInterface(&arr)
	//arr = arr[:0]
	arr = arr[:0]

	if !RestCheckPanic(err, res) {
		//fmt.Fprint(res,string(jsonData))
		res.Write(b.Bytes())
		b = nil
		//fmt.Fprint(res,string(b.Bytes()))

	}

	err = nil

	//jsonData = jsonData[:0]

}

func CheckGrantToPage(userId int64, pageId string) bool {

	o := orm.NewOrm()
	o.Using("default")

	if utils.GetParamValue("page_check_grant") == "0" {
		return true
	}
	if utils.GetRoleParamValue(o, userId, "is_admin") == "1" {
		return true
	}

	cnt := 0
	db, _ := utils.NewDB()
	defer db.Close()
	err := db.QueryRow(utils.DbBindReplace("select count(1) cnt from pages p,role_pages rp,user_roles ur where ur.user_id=? and ur.role_id=rp.role_id and p.id=rp.page_id and p.id=?"), userId, pageId).Scan(&cnt)
	CheckPanic(err)
	//log.Println(cnt)
	db.Close()

	return cnt > 0
}

func CheckGrantToPageCode(userId int64, pageCode string) bool {

	o := orm.NewOrm()
	o.Using("default")

	if utils.GetParamValue("page_check_grant") == "0" {
		return true
	}
	if utils.GetRoleParamValue(o, userId, "is_admin") == "1" {
		return true
	}

	db, _ := utils.NewDB()
	defer db.Close()
	cnt := 0
	err := db.QueryRow(utils.DbBindReplace("select count(1) cnt from pages p,role_pages rp,user_roles ur where ur.user_id=? and ur.role_id=rp.role_id and p.id=rp.page_id and p.code=?"), userId, pageCode).Scan(&cnt)
	CheckPanic(err)
	//log.Println(cnt)
	db.Close()
	return cnt > 0
}

func PageRestApiGetPageTemplate(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	/*
		if RestCheckAuth(res,req){
			return
		}
	*/
	o := orm.NewOrm()
	o.Using("default")
	eTag := "NONE"
	o.Raw(utils.DbBindReplace("select substr(md5(max(updated_at)),1,5) from pages")).QueryRow(&eTag)
	versionNum := utils.GetParamValue("version_num")
	eTag = eTag + versionNum
	res.Header().Set("Etag", `"`+eTag+`"`)
	max_age := utils.GetParamValue("page-max-age")
	if max_age == "" {
		max_age = "0"
	}
	res.Header().Set("Cache-Control", "max-age="+max_age)

	if match := req.Header.Get("If-None-Match"); match != "" {

		//log.Println("match",match)
		if strings.Contains(match, eTag) {
			res.WriteHeader(http.StatusNotModified)
			return
		}
	}
	if 1 == 0 {
		return
	}
	req.ParseForm()
	s := ""
	db, _ := utils.NewDB()
	defer db.Close()
	if req.Form.Get("id") != "" {

		//err1 := db.QueryRow("SELECT p.template FROM pages p where p.id=?", req.Form.Get("id")).Scan(&s)

		o := orm.NewOrm()
		o.Using("default")
		err1 := o.Raw(utils.DbBindReplace("SELECT p.template FROM pages p where p.id=?"), req.Form.Get("id")).QueryRow(&s)
		/*tmp, err := GetCustomTemplatePageById(req.Form.Get("id"), utils.UserId(req))
		if err == nil {
			s = tmp
		} else {
			//log.Println("suk emes "+err.Error())
		}*/
		if err1 != nil {
			log.Println("Error on get page", err1.Error())
			err1 = db.QueryRow(utils.DbBindReplace("SELECT p.template FROM pages p where p.code=?"), "404").Scan(&s)
			//RestCheckPanic(err1, res)

		}
		/*
			if !CheckGrantToPage(utils.UserId(req), req.Form.Get("id")) {
				err1 = db.QueryRow("SELECT p.template FROM pages p where p.code=?", "403").Scan(&s)
				RestCheckPanic(err1, res)
			}
		*/

		fmt.Fprint(res, s)
	} else if req.Form.Get("code") != "" {
		err2 := db.QueryRow(utils.DbBindReplace("SELECT p.template FROM pages p where p.code=?"), req.Form.Get("code")).Scan(&s)

		if err2 != nil {
			err2 := db.QueryRow(utils.DbBindReplace("SELECT p.template FROM pages p where p.code=?"), "404").Scan(&s)
			RestCheckPanic(err2, res)
			db.Close()
			fmt.Fprint(res, s)
			return
		}
		if !CheckGrantToPageCode(utils.UserId(req), req.Form.Get("code")) {
			err1 := db.QueryRow("SELECT p.template FROM pages p where p.code=?", "403").Scan(&s)
			RestCheckPanic(err1, res)
			db.Close()
			fmt.Fprint(res, s)
			return
		}

		tmp, err := utils.GetCustomTemplatePageByCode(req.Form.Get("code"), utils.UserId(req))
		if err == nil {
			s = tmp
		}

		db.Close()
		fmt.Fprint(res, s)
	}

}
