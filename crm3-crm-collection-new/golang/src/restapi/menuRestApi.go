package restapi

import "net/http"
import (
	"encoding/json"
	"fmt"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

type SubMenu struct {
	Id      int64  `json:"id"`
	IdHi    int64  `json:"id_hi"`
	Title   string `json:"title"`
	TitleEn string `json:"title_en"`
	TitleRu string `json:"title_ru"`
	TitleKk string `json:"title_kk"`
	Url     string `json:"url"`
	Icon    string `json:"icon"`
	Code    string `json:"code"`
}

type Menu struct {
	Id       int64     `json:"id"`
	IdHi     int64     `json:"id_hi"`
	Title    string    `json:"title"`
	TitleEn  string    `json:"title_en"`
	TitleRu  string    `json:"title_ru"`
	TitleKk  string    `json:"title_kk"`
	Url      string    `json:"url"`
	Icon     string    `json:"icon"`
	Code     string    `json:"code"`
	CntChild int64     `json:"cnt_child"`
	Items    []SubMenu `json:"items"`
}

type Menu_v2 struct {
	BpProcessCode string    `json:"bp_process_code"`
	Id            int64     `json:"id"`
	IdHi          int64     `json:"id_hi"`
	Title         string    `json:"title"`
	TitleEn       string    `json:"title_en"`
	TitleRu       string    `json:"title_ru"`
	TitleKk       string    `json:"title_kk"`
	Url           string    `json:"url"`
	Icon          string    `json:"icon"`
	Code          string    `json:"code"`
	CntChild      int64     `json:"cnt_child"`
	Items         []SubMenu `json:"items"`
}

func MenuRestApiGetTree_v2(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type", "application/json")
	if RestCheckAuth(res, req) {
		return
	}

	o := orm.NewOrm()
	o.Using("default")
	var arr []Menu_v2
	//var subArr [] SubMenu
	_, err := o.Raw(utils.DbBindReplace(`SELECT
	case when m.by_process=1 then pr.code else null end as "bp_process_code", 
		m.id as "id",
		m.id_hi as "id_hi",
		m.title as "title", 
		case when m.by_url = 1 then  m.url else concat('#',p.url) end as "url"
	    ,coalesce(nullif(m.icon,''),p.icon) as "icon",m.code as "code"
		,(select en from translates t where t.code=m.title limit 1) as "title_en"
		,(select ru from translates t where t.code=m.title limit 1) as "title_ru"
		,(select kk from translates t where t.code=m.title limit 1) as "title_kk"
		,(select count(1) from menus mmm where mmm.id_hi=m.id and mmm.active=1) as "cnt_child"
		 FROM menus m
		 left join pages p on p.id=m.page_id
		 left join bp_processes pr on pr.id=m.bp_process_id
		 where m.id_hi is null and m.active=1
		and exists (select 1 from role_menus rm,
		user_roles ur where rm.role_id=ur.role_id and ur.user_id=? and m.id=rm.menu_id) order by m.position`),
		utils.UserId(req)).QueryRows(&arr)
	CheckPanic(err)
	for index, element := range arr {
		o.Raw(utils.DbBindReplace(`SELECT case when m.by_process=1 then pr.code else null end as bp_process_code, m.id as "id",m.id_hi as "id_hi",m.title as "title",
		case when m.by_url =1 then m.url else (select concat('#',url) from pages where id=m.page_id )end as "url",
		m.icon as "icon",m.code as "code"
			,(select en from translates t where t.code=m.title limit 1) as "title_en" 
			,(select ru from translates t where t.code=m.title limit 1) as "title_ru" 
			,(select kk from translates t where t.code=m.title limit 1) as "title_kk" 
			 FROM menus m 
			 left join bp_processes pr on pr.id=m.bp_process_id
			 where m.id_hi=? and m.active=1 
			 
			and exists (select 1 from
				 role_menus rm,
				 user_roles ur where rm.role_id=ur.role_id and ur.user_id=? and m.id=rm.menu_id) order by m.position`),
			element.Id, utils.UserId(req)).QueryRows(&arr[index].Items)
		//arr[index].Items = subArr
		//utils.ClearInterface(&subArr)
		//subArr = nil

	}

	jsonData, err := json.Marshal(arr)
	utils.ClearInterface(&arr)
	arr = nil
	CheckPanic(err)
	fmt.Fprint(res, string(jsonData))

	jsonData = jsonData[:0]
	o = nil

}

func MenuRestApiGetTree(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type", "application/json")
	if RestCheckAuth(res, req) {
		return
	}

	o := orm.NewOrm()
	o.Using("default")
	var arr []Menu
	//var subArr [] SubMenu
	_, err := o.Raw(utils.DbBindReplace(`SELECT 
	
	m.id as "id",m.id_hi as "id_hi",m.title as "title",  case when m.by_url = 1 then  m.url else concat('#',p.url) end as "url"
	    ,coalesce(nullif(m.icon,''),p.icon) as "icon",m.code as "code"
		,(select en from translates t where t.code=m.title limit 1) as "title_en"
		,(select ru from translates t where t.code=m.title limit 1) as "title_ru"
		,(select kk from translates t where t.code=m.title limit 1) as "title_kk"
		,(select count(1) from menus mmm where mmm.id_hi=m.id and mmm.active=1) as "cnt_child"
		 FROM menus m
		 left join pages p on p.id=m.page_id
		 where m.id_hi is null and m.active=1
		and exists (select 1 from role_menus rm,
		user_roles ur where rm.role_id=ur.role_id and ur.user_id=? and m.id=rm.menu_id) order by m.position`),
		utils.UserId(req)).QueryRows(&arr)
	CheckPanic(err)
	for index, element := range arr {
		o.Raw(utils.DbBindReplace(`SELECT m.id as "id",m.id_hi as "id_hi",m.title as "title",
		case when m.by_url =1 then m.url else (select concat('#',url) from pages where id=m.page_id )end as "url",
		m.icon as "icon",m.code as "code"
			,(select en from translates t where t.code=m.title limit 1) as "title_en" 
			,(select ru from translates t where t.code=m.title limit 1) as "title_ru" 
			,(select kk from translates t where t.code=m.title limit 1) as "title_kk" 
			 FROM menus m where m.id_hi=? and m.active=1 
			and exists (select 1 from
				 role_menus rm,
				 user_roles ur where rm.role_id=ur.role_id and ur.user_id=? and m.id=rm.menu_id) order by m.position`),
			element.Id, utils.UserId(req)).QueryRows(&arr[index].Items)
		//arr[index].Items = subArr
		//utils.ClearInterface(&subArr)
		//subArr = nil

	}

	jsonData, err := json.Marshal(arr)
	utils.ClearInterface(&arr)
	arr = nil
	CheckPanic(err)
	fmt.Fprint(res, string(jsonData))
	utils.ClearInterface(&jsonData)
	o = nil

}
