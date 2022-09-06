package utils

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

type DetailSql struct {
	EntityCode string `json:"entity_code"`
	SqlText    string `json:"sql_text"`
	Code       string `json:"code"`
}

func Detail(o orm.Ormer, id int, code string, user_id int64) (map[string]interface{}, error) {

	//defer ClearInterface(&o)
	if !DetailGrantCheck(o, code, int64(id), user_id) {
		return nil, errors.New("DetailGrantCheck")
	}

	//var result map[string]interface{}
	result := make(map[string]interface{})

	extra := []orm.Params{}

	ws := []DetailSql{}

	//EntityCode            string `json:"entity_code"`
	//SqlText               string `json:"sql_text"`
	//SqlConditionBuildText string `json:"sql_condition_build_text"`
	//Code                  string `json:"code"`

	_, err := o.Raw(DbBindReplace(`select (select code from entities where id=w.entity_id) as "entity_code", 
	ws.sql_text as "sql_text",
	ws.code as "code"
	from details w, detail_queries ws where ws.detail_id=w.id and w.code=?`), code).QueryRows(&ws)
	if err != nil {
		return nil, err
	}
	//CheckPanic(err)

	//fmt.Fprint(res,&ws)
	i := 0
	//fmt.Print(len(ws))

	for _, element := range ws {

		i++

		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("DEBUG Detail element", element)
		}

		element.SqlText = strings.Replace(element.SqlText, ":user_id", strconv.Itoa(int(user_id)), -1)
		//element.SqlText = strings.Replace(element.SqlText, ":id", strconv.Itoa(id), -1)

		splitSQL := strings.Split(element.SqlText, "?")
		var ids []int
		maps := []orm.Params{}

		for j := 2; j <= len(splitSQL); j++ {
			ids = append(ids, id)
		}
		sql := DbBindReplace(element.SqlText)
		_, err := o.Raw(DbBindReplace(sql), ids).Values(&maps)

		//checkErr(err)
		if err != nil {
			log.Println("Err Detail Sql=" + element.SqlText + " " + err.Error())
			//RestCheckDBPanic(err,res,o)
			return nil, err
		} else {
			//log.Println("sql ok"+element.SqlText)
		}

		if len(maps) == 0 {
			result[element.Code] = []orm.Params{}
		} else {
			result[element.Code] = maps
		}

		_, err = o.Raw(DbBindReplace(`select dq.code detail_query_code,e.code entity_code from details d
		join detail_queries dq on dq.detail_id=d.id
		join entities e on e.id=dq.entity_id		
		where d.code = ?`), code).Values(&extra)

		if len(extra) == 0 {
			result["_extra"] = []orm.Params{}
		} else {
			result["_extra"] = extra
		}

		ClearInterface(&ids)

		for m := range maps {
			ClearInterface(&m)
		}
		for m := range extra {
			ClearInterface(&m)
		}

		ClearInterface(&maps)
		ClearInterface(&extra)

	}

	for m := range ws {
		ClearInterface(&m)
	}
	ClearInterface(&ws)

	return result, nil
	//	fmt.Fprintln(res,"}")

}
