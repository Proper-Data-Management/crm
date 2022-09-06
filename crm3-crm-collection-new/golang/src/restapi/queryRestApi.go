package restapi

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"

	"encoding/csv"

	jsoniter "github.com/json-iterator/go"
)

type queryGetResponse struct {
	IsMobile         bool         `json:"isMobile"`
	Lang             string       `json:"lang"`
	EntityCode       string       `json:"entityCode"`
	PageCount        int          `json:"pageCount"`
	AllCount         int          `json:"allCount"`
	Title            string       `json:"title"`
	Error            string       `json:"error"`
	GetSelectedTitle string       `json:"getSelectedTitle"`
	GetSelectedRow   interface{}  `json:"getSelectedRow"`
	Items            []orm.Params `json:"items"`
	NeedFilter       int          `json:"needFilter"`
}

func QueryRestApiGet(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	//return

	o2 := orm.NewOrm()
	o2.Using("default")
	//log.Println("By row2")

	//var arr2 [] orm.ParamsList

	err := req.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	limitFrom := "1"
	limitTo := 5

	json := jsoniter.ConfigCompatibleWithStandardLibrary

	iPerPage, _ := strconv.ParseInt(req.Form.Get("perpage"), 10, 32)
	iPage, _ := strconv.ParseInt(req.Form.Get("page"), 10, 32)
	code := req.Form.Get("code")

	if code == "" {
		return
	}

	limitFrom = strconv.Itoa(int(iPerPage * (iPage - 1)))
	//if limitFrom==0{

	//}
	//fmt.Println(limitFrom)
	limitTo, err = strconv.Atoi(req.Form.Get("perpage"))
	if err != nil {
		limitTo = 0
	}
	pageCount := 0

	sql := ""
	title := ""
	isPublic := ""
	connStr := ""
	extdb_id := ""
	entityId := int64(0)
	respO := queryGetResponse{}

	//respO.Lang = utils.GetLanguage2(req)

	err = o2.Raw(utils.DbBindReplace(
		`select main.need_filter, main.entity_id,(select code from entities where id=main.entity_id) entity_code, 
	main.extdb_id, main.sql_text,main.title,main.is_public,(select connstr from extdbs where id=main.extdb_id) extdb_connstr 
	from queries main where code=?`), code).QueryRow(&respO.NeedFilter, &entityId, &respO.EntityCode, &extdb_id, &sql, &title, &isPublic, &connStr)
	if RestCheckDBPanic(err, res, o2) {

		log.Println("QueryRestApiError1", sql, code, err)
		//utils.ErrorWrite("QueryRestApiError1", "", err)
		err = nil
		o2 = nil
		return
	}

	if isPublic != "1" {
		if RestCheckAuth(res, req) {
			o2 = nil
			return
		}
	}

	o := orm.NewOrm()
	extdb_id = "ext" + extdb_id
	if connStr != "" {
		//orm.

		_, err := orm.GetDB(extdb_id)
		if err != nil {

			err = orm.RegisterDataBase(extdb_id, "mysql", connStr)
			if RestCheckDBPanic(err, res, o) {

				log.Println(sql)
				utils.ErrorWrite("QueryRestApiError2", "", err)
				err = nil
				o = nil
				return
			}
		}
		err = o.Using(extdb_id)
		log.Println("using external db " + extdb_id + " " + connStr)
		if RestCheckDBPanic(err, res, o) {

			log.Println(sql)
			utils.ErrorWrite("QueryRestApiError3", "", err)
			err = nil
			o = nil
			return
		}

	} else {
		o.Using("default")
	}

	origSql := sql
	lang := utils.GetLanguage2(req)
	foundFlt, sql, filterArray, formArray, err := utils.QueryFilterBuild(o, entityId, sql, req.Form, utils.UserId(req), lang)

	//log.Println("filterArray", filterArray)
	//log.Println("formArray", filterArray)
	//log.Println("sql", sql)
	filledFormArray := 0
	for _, v := range formArray {
		if v != "" {
			filledFormArray++
		}
	}
	if !foundFlt && filledFormArray == 0 && len(filterArray) == 0 && respO.NeedFilter == 1 {
		respO.Error = "1"
		jsonData, _ := json.Marshal(respO)
		_, err = fmt.Fprint(res, string(jsonData))
		return
	}

	sql = strings.Replace(sql, ":domain", "'"+utils.SQLInjectTruncate(req.Host)+"'", -1)

	if req.Form.Get("getTitleById") != "" {
		sql1 := "select main.title from " + respO.EntityCode + " main where id=?"
		o.Raw(utils.DbBindReplace(sql1), req.Form.Get("getTitleById")).QueryRow(&respO.GetSelectedTitle)
		//		if RestCheckDBPanic(err ,res ,o ) {
		//			log.Println("getTitleById " + sql1)
		//			utils.ErrorWrite("QueryRestApiError4","",err)
		//			return
		//		}
	}

	if RestCheckDBPanic(err, res, o) {
		log.Println(sql)
		utils.ErrorWrite("QueryRestApiError5", "", err)
		return
	}

	arr := []orm.Params{}

	//log.Println("sql2 = "+sql)
	if limitTo == 0 { //No pagination

		pr, err := o.Raw(utils.DbBindReplace(sql)).Prepare()

		if RestCheckDBPanic(err, res, o) {
			log.Println("sql err 1 = " + sql)
			utils.ErrorWrite("QueryRestApiError", "", err)
			return
		}

		defer pr.Close()

		if RestCheckDBPanic(err, res, o) {
			log.Println("sql err 2 = " + sql)
			utils.ErrorWrite("QueryRestApiError", "", err)
			return
		}
		//allArray := append(filterArray, formArray)
		_, err = pr.Values(&arr, filterArray, formArray)

		if RestCheckDBPanic(err, res, o) {
			log.Println("sql err 2 = " + sql)
			utils.ErrorWrite("QueryRestApiError", "", err)
			return
		}

	} else {

		if strings.Contains(sql, "SQL_CALC_FOUND_ROWS") {

			respO.AllCount = 1000000
			pageCount = 1000

		} else {
			cntSQL := "SELECT count(1) FROM (" + sql + ") alldata"
			err := o.Raw(utils.DbBindReplace(cntSQL), filterArray, formArray).QueryRow(&respO.AllCount)
			if RestCheckDBPanic(err, res, o) {
				log.Println("sql = " + cntSQL)
				log.Println("QueryRestApiError6 = " + err.Error())
				//utils.ErrorWrite("QueryRestApiError6","",err)

				return
			}

			pageCount = int(math.Ceil(float64(respO.AllCount))/float64(limitTo)) + 1

		}

		o := orm.NewOrm()
		o.Using("default")
		if utils.GetDbDriverType() == orm.DRMySQL || utils.GetDbDriverType() == orm.DRPostgres {
			_, err = o.Raw(utils.DbBindReplace(sql+" limit ? offset ?"), filterArray, formArray, limitTo, limitFrom).Values(&arr)
		} else if utils.GetDbDriverType() == orm.DROracle {
			_, err = o.Raw(utils.DbBindReplace(sql+" OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"), filterArray, formArray, limitFrom,
				limitTo).Values(&arr)
		}
		if RestCheckDBPanic(err, res, o) {

			log.Println(sql)
			log.Println(limitFrom)
			log.Println(limitTo)
			utils.ErrorWrite("QueryRestApiError31", "", err)
			err = nil
			o = nil
			return
		}

	}

	respO.IsMobile = IsMobile(req)

	//log.Println("By row3", req.Form.Get("getRowById"))

	if req.Form.Get("getRowById") != "" && req.Form.Get("getRowById") != "null" && req.Form.Get("getRowById") != "undefined" {
		rowId, _ := strconv.Atoi(req.Form.Get("getRowById"))
		//log.Println("By row", rowId)
		sql2 := origSql
		sql2 = strings.Replace(sql2, "%filter%", "where main.id=?", -1)
		sql2 = strings.Replace(sql2, "%order%", "", -1)
		sql2 = strings.Replace(sql2, ":user_id", fmt.Sprintf("%v", utils.UserId(req)), -1)
		
		//log.Println("getRowById",sql2, rowId, formArray)
		if utils.DetailEntityGrantCheck(o, respO.EntityCode, int64(rowId), utils.UserId(req)) {
			oo := []orm.Params{}
			_, err = o.Raw(utils.DbBindReplace(sql2), rowId, formArray).Values(&oo)
			if err != nil {
				log.Println("error! detail check ", err.Error(), code)
			}

			if len(oo) == 1 {
				respO.GetSelectedRow = oo[0]
				//log.Println("By row len", len(oo), oo[0])
			}
		}

	}

	//	if req.Form.Get("getRowById")!="" {
	//		for _,v := range arr {
	//			if v["id"] == req.Form.Get("getRowById"){
	//				respO.GetSelectedRow = v
	//			}
	//		}
	//	}

	respO.Items = arr
	respO.Error = "0"
	respO.Title = title
	respO.PageCount = pageCount
	jsonData, err := json.Marshal(respO)

	if !RestCheckDBPanic(err, res, o) {

		if req.Form.Get("format") == "csv" {
			res.Header().Add("Content-Type", "text/csv")
			res.Header().Add("Content-Disposition", "inline; filename=\""+code+".csv\"")

			w := csv.NewWriter(res)
			w.Comma = '\t'

			h := respO.Items[0]
			fields := []string{}
			for key, _ := range h {
				fields = append(fields, key)
			}

			i := 0
			for _, v := range respO.Items {
				i++
				row := []string{}

				for _, value := range fields {
					if v[value] != nil {
						row = append(row, v[value].(string))
					} else {
						row = append(row, "")
					}
				}
				if i == 1 {
					err := w.Write(fields)
					if err != nil {
						log.Println("err " + err.Error())
					}
				}
				err := w.Write(row)
				if err != nil {
					log.Println("err " + err.Error())
				}

				w.Flush()

			}
		} else {
			_, err = fmt.Fprint(res, string(jsonData))
			if err != nil {
				log.Println("err " + err.Error())
			}

		}
	}
	//arr = nil
	formArray = nil
	filterArray = nil
	err = nil
	//jsonData = nil
	utils.ClearInterface(&jsonData)

	for i := range arr {
		utils.ClearInterface(&i)
	}

	utils.ClearInterface(&arr)
	o = nil

	respO.Items = nil
	for i := range respO.Items {
		utils.ClearInterface(&i)
	}
	utils.ClearInterface(&respO.Items)
	utils.ClearInterface(&respO.GetSelectedRow)
	utils.ClearInterface(&respO)
	//respO = queryGetResponse{}

	for i := range filterArray {
		utils.ClearInterface(&i)
	}
	utils.ClearInterface(&filterArray)

	for i := range formArray {
		utils.ClearInterface(&i)
	}
	utils.ClearInterface(&formArray)

}
