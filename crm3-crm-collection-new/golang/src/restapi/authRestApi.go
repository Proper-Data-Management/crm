package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"

	"log"
)

type LoginResponse struct {
	Result      string "json:`result`"
	RedirectURL string "json:`redirectURL`"
	AuthToken   string "json:`authToken`"
}

type LoginResponseExt struct {
	Result      string "json:`result`"
}

type LoginRequest struct {
	Login       string "json:`login`"
	Password    string "json:`password`"
	System      string "json:`system`"
	GetToken    bool   "json:`getToken`"
	Uri         string "json:`uri`"
	DeviceToken string "json:`deviceToken`"
	ICloudId    string "json:`iCloudId`"
}

const loginIncorrect = "incorrect"
const loginOk = "ok"
const loginLocked = "locked"
const loginTokenError = "tokenError"
const passwordExpired = "passwordExpired"
const loginUnknownError = "unknownError"

func LoginByToken(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	//res.Header().Add()
	req.ParseForm()
	res.Header()

	o := orm.NewOrm()
	o.Using("default")

	user_id := int64(0)
	system := ""
	err := o.Raw(utils.DbBindReplace("select user_id,system from auth_tokens where sys$uuid=?"), param.ByName("token")).QueryRow(&user_id, &system)
	if err != nil {
		if !utils.IsNoRowFound(err) {
			log.Println("error on LoginByToken " + err.Error())
		}
		http.Redirect(res, req, utils.GetDomainParamValue(req.Host, "loginpage")+"#invalidloginLdap", 301)
	}
	_, err = o.Raw(utils.DbBindReplace("delete from auth_tokens where sys$uuid=? and coalesce(reusable,0)=0"), param.ByName("token")).Exec()
	if err != nil {
		http.Redirect(res, req, utils.GetDomainParamValue(req.Host, "loginpage")+"#invalidloginLdap", 301)
	} else {

		session, _ := utils.GetStore().Get(req, utils.SessionName)

		if req.Form.Get("lang") != "" {
			//log.Println("lang = ")
			//log.Println(req.Form.Get("lang"))
			session.Values["lang"] = req.Form.Get("lang")
		}
		session.Values["user_id"] = user_id
		session.Values["system"] = system
		session.Values["uri"] = "/"
		//log.Println("user_id")
		//log.Println(session.Values["user_id"])
		session.Save(req, res)
		http.Redirect(res, req, "/static/#"+utils.GetUsersHomePage(req), 301)
	}

}

func Login(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	var request LoginRequest
	err := decoder.Decode(&request)
	if err != nil {
		panic(err)
	}
	var result LoginResponse
	result.Result = loginUnknownError + "XXX"
	req.ParseForm()
	utils.DoLoginLog(utils.UserId(req), 1)
	o := orm.NewOrm()
	o.Using("default")
	session, _ := utils.GetStore().Get(req, utils.SessionName)
	//session.Options.MaxAge =	3600
	user_id := int64(0)
	oldPassword := ""
	loginWithoutPassword := 0
	isPassExpired := 0
	isLocked := 0

	if request.ICloudId != "" {
		o.Raw(utils.DbBindReplace("insert into users (email,login_without_password) values (?,?)"), request.ICloudId, 1).Exec()
		request.Login = request.ICloudId
	}
	ldap_id := int64(0)
	err = o.Raw(utils.DbBindReplace("select ldap_id,id,password,login_without_password,is_locked,case when pass_expired_at<now() then 1 else 0 end pass_expired  from users where (email=? or login=?)"), request.Login, request.Login).QueryRow(&ldap_id, &user_id, &oldPassword, &loginWithoutPassword, &isLocked, &isPassExpired)
	log.Println("error on Login", err)
	//log.Println("ldap_id",ldap_id)
	if ldap_id != 0 {
		err := utils.LdapAuthByLdapId(ldap_id, request.Login, request.Password)
		if err != nil {
			log.Println(err)
			result.Result = loginIncorrect
			jsonData, _ := json.Marshal(result)
			fmt.Fprint(res, string(jsonData))
			return
			//http.Redirect(res, req, utils.GetDomainParamValue(req.Host, "loginpage")+"#invalidloginLdap", 301)
		}

		if !request.GetToken {
			result.Result = loginOk
			session.Values["user_id"] = user_id
			session.Values["system"] = request.System
			session.Values["uri"] = request.Uri

			_, err2 := o.Raw(utils.DbBindReplace("update users set last_channel = ? where id=?"), request.System, user_id).Exec()
			if err2 != nil {
				log.Println("Warning, Error on set last_channel")
			}

			session.Save(req, res)
			result.RedirectURL = utils.GetUsersHomePage(req)
			jsonData, _ := json.Marshal(result)
			fmt.Fprint(res, string(jsonData))
			return
		} else {
			rs, err := o.Raw(utils.DbBindReplace("insert into auth_tokens (user_id,system,reusable,expire_at) values (?,?,1,date_add(now(),interval 45 day))"), user_id, request.System).Exec()
			if err != nil {
				result.Result = loginUnknownError
			}
			lid, err := rs.LastInsertId()
			err = o.Raw(utils.DbBindReplace("select sys$uuid from auth_tokens where id=?"), lid).QueryRow(&result.AuthToken)
			if err != nil {
				result.Result = loginUnknownError
			} else {
				result.Result = loginOk
			}
			jsonData, _ := json.Marshal(result)
			fmt.Fprint(res, string(jsonData))
			return
		}
	}

	if isPassExpired == 1 {
		result.Result = passwordExpired
		jsonData, _ := json.Marshal(result)
		fmt.Fprint(res, string(jsonData))
		return
	}

	if isLocked == 1 {
		result.Result = loginLocked
		jsonData, _ := json.Marshal(result)
		fmt.Fprint(res, string(jsonData))
		return
	}

	if loginWithoutPassword == 0 {
		oldPasswordByte1 := []byte(oldPassword)
		oldPasswordByte2 := []byte(request.Password)
		err = bcrypt.CompareHashAndPassword(oldPasswordByte1, oldPasswordByte2)
	}

	//log.Println("request.DeviceToken" + request.DeviceToken)
	//log.Println("request.System" + request.System)
	if err != nil {
		result.Result = loginIncorrect
	} else {
		if request.DeviceToken != "" {
			_, err2 := o.Raw(utils.DbBindReplace("update users set device_token=? where id=?"), request.DeviceToken, user_id).Exec()
			if err2 != nil {
				log.Println("Warning, Error on set device_token")
			}
		}

		if request.System != "" {
			_, err2 := o.Raw(utils.DbBindReplace("update users set last_channel=? where id=?"), request.System, user_id).Exec()
			if err2 != nil {
				log.Println("Warning, Error on set last_channel")
			}
		}

		result.Result = loginOk

		if !request.GetToken {
			session.Values["user_id"] = user_id
			session.Values["system"] = request.System
			session.Values["uri"] = request.Uri
			session.Save(req, res)
			result.RedirectURL = utils.GetUsersHomePage(req)
		} else {
			rs, err := o.Raw(utils.DbBindReplace("insert into auth_tokens (user_id,system,reusable,expire_at,profile_id) values (?,?,1,date_add(now(),interval 45 day),(select id from auth_profiles where code=?))"), user_id, request.System, utils.GetParamValue("default_oauth_profile")).Exec()

			//rs,err := o.Raw("insert into auth_tokens (user_id,`system`) values (?,?)",user_id,request.System).Exec()
			if err != nil {
				result.Result = loginUnknownError
			}
			lid, err := rs.LastInsertId()
			err = o.Raw(utils.DbBindReplace("select sys$uuid from auth_tokens where id=?"), lid).QueryRow(&result.AuthToken)
			if err != nil {
				result.Result = loginUnknownError
			} else {
				result.Result = loginOk
			}

		}
	}

	jsonData, _ := json.Marshal(result)
	fmt.Fprint(res, string(jsonData))
}

func LoginByTokenExt(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	req.ParseForm()
	res.Header().Add("Content-Type", "application/json")

	o := orm.NewOrm()
	o.Using("default")

	var result LoginResponseExt
	user_id := int64(0)
	system := ""
	err := o.Raw(utils.DbBindReplace("select user_id,system from auth_tokens where sys$uuid = ?"), param.ByName("token")).QueryRow(&user_id, &system)
	if err != nil {
		if !utils.IsNoRowFound(err) {
			log.Println("error on LoginByToken " + err.Error())
		}
		result.Result = loginTokenError
		res.WriteHeader(401)
	} else {
		_, err = o.Raw(utils.DbBindReplace("delete from auth_tokens where sys$uuid = ? and coalesce(reusable,0) = 0"), param.ByName("token")).Exec()
		if err != nil {
			result.Result = loginTokenError
			res.WriteHeader(401)
		} else {
			session, _ := utils.GetStore().Get(req, utils.SessionName)
			if req.Form.Get("lang") != "" {
				session.Values["lang"] = req.Form.Get("lang")
			}
			session.Values["user_id"] = user_id
			session.Values["system"] = system
			session.Values["uri"] = "/"
			session.Save(req, res)
			result.Result = loginOk
		}
	}

	jsonData, _ := json.Marshal(result)
	fmt.Fprint(res, string(jsonData))
}

func LoginExt(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	res.Header().Add("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	var request LoginRequest
	err := decoder.Decode(&request)
	if err != nil {
		panic(err)
	}
	var result LoginResponseExt
	req.ParseForm()
	utils.DoLoginLog(utils.UserId(req), 1)
	o := orm.NewOrm()
	o.Using("default")
	session, _ := utils.GetStore().Get(req, utils.SessionName)
	//session.Options.MaxAge =	3600
	user_id := int64(0)
	oldPassword := ""
	loginWithoutPassword := 0
	isPassExpired := 0
	isLocked := 0

	if request.ICloudId != "" {
		o.Raw(utils.DbBindReplace("insert into users (email,login_without_password) values (?,?)"), request.ICloudId, 1).Exec()
		request.Login = request.ICloudId
	}
	ldap_id := int64(0)
	err = o.Raw(utils.DbBindReplace("select ldap_id,id,password,login_without_password, is_locked, case when pass_expired_at<now() then 1 else 0 end pass_expired  from users where (email=? or login=?)"), request.Login, request.Login).QueryRow(&ldap_id, &user_id, &oldPassword, &loginWithoutPassword, &isLocked, &isPassExpired)

	if isPassExpired == 1 {
		result.Result = passwordExpired
		res.WriteHeader(401)
		jsonData, _ := json.Marshal(result)
		fmt.Fprint(res, string(jsonData))
		return
	}

	if isLocked == 1 {
		result.Result = loginLocked
		res.WriteHeader(423)
		jsonData, _ := json.Marshal(result)
		fmt.Fprint(res, string(jsonData))
		return
	}

	if loginWithoutPassword == 0 {
		oldPasswordByte1 := []byte(oldPassword)
		oldPasswordByte2 := []byte(request.Password)
		err = bcrypt.CompareHashAndPassword(oldPasswordByte1, oldPasswordByte2)
	}

	if err != nil {
		log.Println("error on Login", err)
		result.Result = loginIncorrect
		res.WriteHeader(401)
	} else {
		if request.DeviceToken != "" {
			_, err2 := o.Raw(utils.DbBindReplace("update users set device_token = ? where id=?"), request.DeviceToken, user_id).Exec()
			if err2 != nil {
				log.Println("Warning, Error on set device_token")
			}
		}

		if request.System != "" {
			_, err2 := o.Raw(utils.DbBindReplace("update users set last_channel = ? where id=?"), request.System, user_id).Exec()
			if err2 != nil {
				log.Println("Warning, Error on set last_channel")
			}
		}

		result.Result = loginOk
		session.Values["user_id"] = user_id
		session.Values["system"] = request.System
		session.Values["uri"] = request.Uri
		session.Save(req, res)

	}

	jsonData, _ := json.Marshal(result)
	fmt.Fprint(res, string(jsonData))
}
