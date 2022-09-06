package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func ImportProfileImage(w http.ResponseWriter, r *http.Request, file_id int64, uuid string) error {

	o := orm.NewOrm()
	o.Using("default")
	o.Raw(utils.DbBindReplace("update users set user_pic_file_id=? where id=?"), file_id, r.Form.Get("user_id")).Exec()

	errRes := TStandartUploadResponse{Url: utils.GetParamValue("ecm_getfile_prefix") + uuid, GuId: uuid, Result: "ok", Id: file_id}
	jsonData, _ := json.Marshal(errRes)
	fmt.Fprint(w, string(jsonData))

	return nil

}
