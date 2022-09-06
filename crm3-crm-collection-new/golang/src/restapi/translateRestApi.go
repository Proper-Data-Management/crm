package restapi

import (
	"net/http"

	"runtime"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"

	"bytes"
	"log"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

type translateGetResponse struct {
	ru string
	kk string
	en string
}

type translates struct {
	Id   int64  `gorm:"save_associations:false"`
	Code string `gorm:"save_associations:false"`
	En   string `gorm:"save_associations:false"`
	Ru   string `gorm:"save_associations:false"`
	Kk   string `gorm:"save_associations:false"`
}

func TranslateRestApiGet(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	//o := orm.NewOrm()
	//o.Using("default")

	o := orm.NewOrm()
	o.Using("default")
	lang := utils.GetLanguage2(req)

	eTag := "NONE"
	err := o.Raw(utils.DbBindReplace("select substr(md5(max(concat(updated_at,''))),1,5) from translates")).QueryRow(&eTag)
	versionNum := utils.GetParamValue("version_num")
	eTag = eTag + versionNum

	if err != nil {
		log.Println("translate1", err)
		return
	}
	eTag = eTag + lang
	res.Header().Set("Etag", `"`+eTag+`"`)
	res.Header().Set("Cache-Control", "max-age=86400")

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, eTag) {
			res.WriteHeader(http.StatusNotModified)
			return
		}
	}

	var kk bytes.Buffer
	var ru bytes.Buffer
	var en bytes.Buffer
	var all bytes.Buffer

	//o, _:= gorm.Open("mysql", os.Getenv("OPENSHIFT_MYSQL_DB_USERNAME")+":"+os.Getenv("OPENSHIFT_MYSQL_DB_PASSWORD")+"@tcp("+os.Getenv("OPENSHIFT_MYSQL_DB_HOST")+":"+os.Getenv("OPENSHIFT_MYSQL_DB_PORT")+")/"+os.Getenv("OPENSHIFT_APP_NAME")+"?charset=utf8")

	//var arr translates
	var arr []orm.Params

	_, err = o.Raw(utils.DbBindReplace(`SELECT code as "code",en as "en",ru as "ru",kk as "kk" FROM translates`)).Values(&arr)
	if err != nil {
		log.Println("translate2", err)
		return
	}
	utils.ClearInterface(&o)
	o = nil

	en.WriteString("\"en\":{")
	ru.WriteString("\"ru\":{")
	kk.WriteString("\"kk\":{")

	for k, v := range arr {

		if v["en"] == nil {
			v["en"] = ""
		}
		if v["ru"] == nil {
			v["ru"] = ""
		}
		if v["kk"] == nil {
			v["kk"] = ""
		}
		en.WriteString(strconv.Quote(v["code"].(string)))
		en.WriteString(":")
		en.WriteString(strconv.Quote(v["en"].(string)))

		ru.WriteString(strconv.Quote(v["code"].(string)))
		ru.WriteString(":")
		ru.WriteString(strconv.Quote(v["ru"].(string)))

		kk.WriteString(strconv.Quote(v["code"].(string)))
		kk.WriteString(":")
		kk.WriteString(strconv.Quote(v["kk"].(string)))

		if k < len(arr)-1 {
			en.WriteString(",")
			ru.WriteString(",")
			kk.WriteString(",")
		}

	}

	all.WriteString("{")
	all.Write(en.Bytes())
	all.WriteString("},")
	all.Write(ru.Bytes())
	all.WriteString("},")
	all.Write(kk.Bytes())
	all.WriteString("},\"lang\":")
	all.WriteString(strconv.Quote(utils.GetLanguage2(req)))
	all.WriteString("}")

	res.Write(all.Bytes())
	utils.ClearInterface(&arr)

	all.Reset()
	//	utils.ClearInterface(&kk)
	//	utils.ClearInterface(&ru)
	//	utils.ClearInterface(&en)
	kk.Reset()
	ru.Reset()
	en.Reset()

	//
	//	kk = ""
	//	ru = ""
	//	en = ""
	//	all = ""

	for k := range arr {
		for x := range arr[k] {
			delete(arr[k], x)
		}
	}

	arr = arr[:0]
	arr = nil

	runtime.GC()

}
