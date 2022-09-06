package restapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

type TStandartUploadResponse struct {
	GuId              string      `json:"guid"`
	Url               string      `json:"url"`
	Id                int64       `json:"id"`
	Result            string      `json:"result"`
	FileName          string      `json:"filename"`
	RestServiceOutput interface{} `json:"restServiceOutput"`
}

func Upload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	w.Header().Set("Content-Type", "application/json")

	r.ParseForm()
	r.ParseMultipartForm(32 << 20)

	if r.ContentLength > 2000000000 {
		fmt.Fprint(w, "{\"result\":\"TOO LARGE FILE\" }")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}

	//log.Println("Content-Transfer-Encoding=" + handler.Header.Get("Content-Transfer-Encoding"))

	if r.Form.Get("dir") == "" {
		log.Panic("DIR is Empty")
	}
	o := orm.NewOrm()
	o.Using("default")

	restBody := ""
	origFileName := os.TempDir() + "/" + handler.Filename

	var file_id = int64(0)
	uuid := r.Form.Get("dir") + "-" + utils.Uuid()
	//if (r.Form.Get("dir")!="") {

	cntAllow := 0
	err = o.Raw(utils.DbBindReplace(`select count(1) from dirs
	where 
	(allowed_ftg_id is null
	or
	allowed_ftg_id in 
	
	(select 1 from file_type_groups ftg
	join file_type_group_types ftgt on ftg.id =ftgt.group_id
	join file_types f on f.id=ftgt.file_type_id
	where f.ext = lower(?)	
	)
	)
	and code=?
	`), filepath.Ext(handler.Filename), r.Form.Get("dir")).QueryRow(&cntAllow)

	if RestCheckPanic(err, w) {
		return
	}

	if cntAllow == 0 {
		fmt.Fprint(w, "{\"result\":\"Access Denied\" }")
		return

	}

	dir := ""
	pth := "unix_path"
	if runtime.GOOS == "windows" {
		pth = "win_path"
	}

	on_upload_script := ""
	filename_as_uuid := 0
	path_expr := ""
	err = o.Raw(utils.DbBindReplace("select path_expr, coalesce(filename_as_uuid,0), on_upload_script, "+pth+",on_upload_script restBody from dirs where code=?"), r.Form.Get("dir")).QueryRow(&path_expr, &filename_as_uuid, &on_upload_script, &dir, &restBody)

	if RestCheckPanic(err, w) {
		return
	}

	baseName := path.Base(strings.Replace(origFileName, "\\", "/", -1))

	fileName := dir + uuid
	filepath := uuid
	path_expr_value := ""
	if path_expr != "" {

		err = o.Raw(utils.DbBindReplace("select "+path_expr+" from dirs main where code=?"), r.Form.Get("dir")).QueryRow(&path_expr_value)
		if RestCheckPanic(err, w) {
			return
		}
		path_expr_value = path_expr_value + "/"
	}

	if filename_as_uuid == 1 {
		fileName = dir + path_expr_value + uuid + "/" + baseName
		filepath = uuid + "/" + baseName
		os.MkdirAll(dir+path_expr_value+uuid, os.ModePerm)
	} else {
		filepath = path_expr_value
		fileName = dir + filepath + uuid
		os.MkdirAll(dir+path_expr_value, os.ModePerm)
	}

	//log.Println("uuid=" , uuid)

	file_id, err = utils.DbInsert(o, "insert into files (filepath,dir_id,code,title,filename) values (?,(select id from dirs where code=?),?,?,?)", filepath, r.Form.Get("dir"), uuid, "TEST123", baseName)

	if RestCheckPanic(err, w) {
		return
	}

	//}

	defer file.Close()

	if handler.Header.Get("Content-Transfer-Encoding") == "base64" {
		base64File := base64.NewDecoder(base64.StdEncoding, file)

		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		//log.Println(fileName)
		if err != nil {
			fmt.Println(err)
			fmt.Fprint(w, "{\"result\": \""+err.Error()+"\"}")
			return
		}

		io.Copy(f, base64File)

	} else {

		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		if err != nil {
			fmt.Println(err)
			fmt.Fprint(w, "{\"result\": \""+err.Error()+"\"}")
			return
		}
		io.Copy(f, file)
	}

	if on_upload_script != "" {

		x := make(map[string]interface{})

		var request ServiceRunRequest
		request.UserId = utils.UserId(r)
		x["id"] = r.Form.Get("id")
		x["fileName"] = fileName
		x["uuid"] = uuid
		x["origFileName"] = origFileName
		x["fileURL"] = utils.GetParamValue("ecm_getfile_prefix") + uuid + "&attachment=true"
		request.Input = x
		log.Println("restBody=" + restBody)
		restOutput, err := RunLuaServiceScript(w, r, request, restBody, 0)
		if err != nil {
			fmt.Println(err)
			fmt.Fprint(w, "{\"result\": \""+err.Error()+"\"}")
			return
		}
		errRes := TStandartUploadResponse{Url: utils.GetParamValue("ecm_getfile_prefix") + uuid, GuId: uuid, Result: "ok", Id: file_id, RestServiceOutput: restOutput}
		jsonData, _ := json.Marshal(errRes)
		fmt.Fprint(w, string(jsonData))

	} else {
		errRes := TStandartUploadResponse{Url: utils.GetParamValue("ecm_getfile_prefix") + uuid, GuId: uuid, FileName: baseName, Result: "ok", Id: file_id}
		jsonData, _ := json.Marshal(errRes)
		fmt.Fprint(w, string(jsonData))
		return
	}
}
