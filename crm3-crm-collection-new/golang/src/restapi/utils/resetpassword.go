package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

type TPasswordResetResult struct {
	Error int64 `json:"error"`
}

type TPasswordResetRequest struct {
	NewPassword string `json:"new_password"`
	OldPassword string `json:"old_password"`
	UserId      string `json:"user_id"`
}

func ForceResetPasswordReq(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	var result TPasswordResetResult

	o := orm.NewOrm()
	o.Using("default")
	if GetRoleParamValue(o, UserId(req), "force_resetpassword") != "1" {
		result.Error = 2
		jsonData, _ := json.Marshal(result)
		fmt.Fprint(res, string(jsonData))
		return
	}
	var request TPasswordResetRequest
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&request)

	email := ""
	o.Raw(DbBindReplace("select email from users where id=?"), request.UserId).QueryRow(&email)
	err := ForceResetPassword(email, request.NewPassword)
	if err != nil {
		fmt.Println("ForceResetPassword:"+err.Error());
		result.Error = 1
	} else {
		result.Error = 0
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(res, string(jsonData))

}

func ResetMyPassword(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	var result TPasswordResetResult
	var request TPasswordResetRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)

	o := orm.NewOrm()
	o.Using("default")

	email, oldPassword := "", ""
	password_not_set := 0
	ldap_id := int64(0)
	o.Raw(DbBindReplace("select email,ldap_id, password,password_not_set from users where id=?"), UserId(req)).QueryRow(&email, &ldap_id, &oldPassword, &password_not_set)

	if ldap_id != 0 {
		err := LdapUserChangeMyPassword(ldap_id, email, request.OldPassword, request.NewPassword)
		if err != nil {
			result.Error = 1
		} else {
			result.Error = 0
		}
		jsonData, _ := json.Marshal(result)
		fmt.Fprint(res, string(jsonData))
		return

	} else {

		if password_not_set == 0 {
			log.Println(oldPassword)
			oldPasswordByte1 := []byte(oldPassword)
			oldPasswordByte2 := []byte(request.OldPassword)

			if GetUserParamValue(o, UserId(req), "force_resetpassword") != "1" {
				err = bcrypt.CompareHashAndPassword(oldPasswordByte1, oldPasswordByte2)
				//fmt.Println(err) // nil means it is a match
				if err != nil {
					result.Error = 1
					jsonData, _ := json.Marshal(result)
					fmt.Fprint(res, string(jsonData))
					log.Println("password incorrect")
					return
				}
			}
		}

		password := []byte(request.NewPassword)

		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

		_, err = o.Raw(DbBindReplace("update users set password_not_set=0,password=? where id=?"), string(hashedPassword), request.UserId).Exec()
		if err != nil {
			panic(err)
		}
		result.Error = 0
		jsonData, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		//checkErr(err)
		fmt.Fprint(res, string(jsonData))
	}

}
