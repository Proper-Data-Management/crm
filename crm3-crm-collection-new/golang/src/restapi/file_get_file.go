package restapi

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"github.com/julienschmidt/httprouter"
)

func GetFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	max_age := utils.GetParamValue("file-max-age")
	if max_age == "" {
		max_age = "0"
	}
	w.Header().Set("Cache-Control", "max-age="+max_age)

	path := "unix_path"
	if runtime.GOOS == "windows" {
		path = "win_path"
	}

	o := orm.NewOrm()
	o.Using("default")

	//log.Println("test")
	fullfileName := ""
	fileName := ""
	contentType := ""
	r.ParseForm()
	uuid := r.Form.Get("code")
	on_read_script := ""
	file_id := int64(0)
	access_control := int64(0)

	sql := `select coalesce(d.access_control,0), f.id, d.on_read_script,
	(select mime from file_types where ext= lower(right(f.filename,4)) ) contenttype,
	 concat(d.` + path + `,coalesce(f.filepath,f.code) ) fullfileName,f.filename 
	 from files f,dirs d where d.id=f.dir_id and f.code=?`

	if utils.GetDbDriverType() == orm.DROracle {
		sql = `select coalesce(d.access_control,0), f.id, d.on_read_script,
		(select mime from file_types where ext= lower(substr(f.filename,-4)) ) contenttype, 
		d.` + path + `||coalesce(f.filepath,f.code)  fullfileName,f.filename 
		from files f,dirs d where d.id=f.dir_id and f.code=?`

	}
	err := o.Raw(utils.DbBindReplace(sql), uuid).QueryRow(&access_control, &file_id, &on_read_script, &contentType, &fullfileName, &fileName)
	if RestCheckPanic(err, w) {
		log.Println("Error r.Form.Get(code)", uuid)
		return
	}

	if access_control != 0 {
		if !utils.DetailEntityGrantCheck(o, "files", file_id, utils.UserId(r)) {
			err = errors.New("Access Denied To File")
			RestCheckPanic(err, w)
			return
		}
	}

	w.Header().Set("Content-Type", contentType)
	if r.Form.Get("attachment") == "true" {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	}
	b, err := ioutil.ReadFile(fullfileName)

	if RestCheckPanic(err, w) {
		b = b[:0]
		return
	}
	w.Write(b)
	b = b[:0]

	x := make(map[string]interface{})

	if on_read_script != "" {
		var request ServiceRunRequest
		x["id"] = r.Form.Get("id")
		x["fileName"] = fileName
		request.Input = x
		request.UserId = utils.UserId(r)
		_, err := RunLuaServiceScript(w, r, request, on_read_script, 0)
		//log.Println("restOutput")
		//log.Println(restOutput)
		if err != nil {
			log.Println("ERROR " + err.Error())
			return
		}

	}

}
