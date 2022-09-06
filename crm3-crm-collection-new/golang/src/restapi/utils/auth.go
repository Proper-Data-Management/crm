package utils

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"

    "git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
    "github.com/gorilla/context"
    "github.com/gorilla/sessions"
    "github.com/julienschmidt/httprouter"
    "golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("asdjkjkl39090wejiosdfklo"))

var SessionName = "CRMSessionId"

func GetStore() *sessions.CookieStore {
    store.MaxAge(0)
    return store
}

func GetSessionData(req *http.Request, code string) string {
    session, err := GetStore().Get(req, SessionName)
    if err != nil {
        log.Println("GetSessionData fail", code, err)
        return ""
    }
    if session != nil && session.Values[code] != nil {
        log.Println("GetSessionData ok")
        return fmt.Sprintf("%s", session.Values[code])
    }
    log.Println("GetSessionData empty")
    return ""
}

func SetSessionData(req *http.Request, code string, value string) {
    session, err := GetStore().Get(req, SessionName)
    if err != nil {
        log.Println("SetSessionData fail", err)
        return
    }
    session.Values[code] = value
}

func AnonymousSessionID(res http.ResponseWriter, req *http.Request) string {
    session, _ := GetStore().Get(req, SessionName)

    if session != nil && session.Values["anonymous_session_id"] == nil || session == nil ||
        session != nil && session.Values["anonymous_session_id"] == "" {
        session.Values["anonymous_session_id"] = Uuid()
        session.Save(req, res)
    }

    if session != nil && session.Values["anonymous_session_id"] != nil {
        return fmt.Sprintf("%s", session.Values["anonymous_session_id"])

    }
    return ""

}

func UserId(req *http.Request) int64 {

    //log.Println("test123")
    session, err := GetStore().Get(req, SessionName)

    /*if err!=nil{
        session = nil
        return int64(0)
    }*/

    if session != nil && session.Values["user_id"] == nil || session == nil {

        //log.Println("test123456")

        if len(req.Header["Authorization"]) > 0 && !strings.HasPrefix(req.Header["Authorization"][0], "NTLM") &&
            !strings.HasPrefix(req.Header["Authorization"][0], "Basic") {
            str := req.Header["Authorization"][0]
            //log.Println(str, "Authorization")
            arr := strings.Split(str, " ")
            if len(arr) > 1 {
                //log.Println("Profile",arr[0])
                //log.Println("Token",arr[1])
                o := orm.NewOrm()
                o.Using("default")
                user_id := int64(0)
                err = o.Raw(DbBindReplace("select at.user_id from auth_tokens at,auth_profiles ap where ap.code=? and at.profile_id=ap.id and "+
                    "at.sys$uuid = ? and at.expire_at>now() and at.reusable=1"), arr[0], arr[1]).QueryRow(&user_id)
                if err == nil {
                    return user_id
                } else {
                    log.Println("Error on get Auth Token. Token Not Found", err)
                    return int64(0)
                }

            }
        }

        if len(req.Header["Authorization"]) > 0 && strings.HasPrefix(req.Header["Authorization"][0], "Basic") {
            str := req.Header["Authorization"][0]
            //log.Println(str, "Authorization")
            arr := strings.Split(str, " ")

            if len(arr) < 1 {
                return int64(0)
            }
            //log.Println("arr Basic", arr[1])
            data, _ := base64.StdEncoding.DecodeString(arr[1])
            //log.Println("arr Basic", string(data))
            lp := strings.Split(string(data), ":")
            if len(lp) < 1 {
                return int64(0)
            }
            userName := lp[0]
            password := lp[1]

            oldPassword := ""
            loginWithoutPassword := 0
            ldapID := int64(0)
            userID := int64(0)
            o := orm.NewOrm()
            o.Using("default")
            err := o.Raw(DbBindReplace("select id,password,login_without_password,ldap_id from users where (email=? or login=?) limit 1"), userName, userName).
                QueryRow(&userID, &oldPassword, &loginWithoutPassword, &ldapID)

            if err != nil {
                log.Println("Access Denied on Query", err)
                return int64(0)
            }
            log.Println("ldap_id", ldapID)
            if ldapID != 0 {
                err := LdapAuthByLdapId(ldapID, userName, password)
                if err != nil {
                    return int64(0)
                } else {
                    return userID
                }
            }
            if loginWithoutPassword == 0 {
                oldPasswordByte1 := []byte(oldPassword)
                oldPasswordByte2 := []byte(password)
                err = bcrypt.CompareHashAndPassword(oldPasswordByte1, oldPasswordByte2)
                if err == nil {
                    log.Println("Access Ok")
                    return userID

                } else {
                    log.Println("Access Denied on compare", err)
                    return int64(0)
                }
            }

        }

        //session.Values["user_id"]=int64(0)
        //session = nil
        ClearInterface(&session)
        context.Clear(req)

        //return int64(0)
    } else {
        //session.Options.MaxAge =  3600
    }

    if len(req.Form["access_token"]) > 0 {

        log.Println("access_token", req.Form["access_token"])
        o := orm.NewOrm()
        o.Using("default")
        user_id := int64(0)
        err = o.Raw(DbBindReplace("select at.user_id from auth_tokens at where "+
            "at.sys$uuid = ? and at.expire_at>now() and at.reusable=1"), req.Form["access_token"][0]).QueryRow(&user_id)
        if err == nil {
            ClearInterface(&session)
            return user_id
        } else {
            log.Println("Error on get Auth Token. Token Not Found", err)
            return int64(0)
        }

        return 0
    }

    res := int64(0)
    if session != nil && session.Values["user_id"] != nil {
        res = session.Values["user_id"].(int64)
        ClearInterface(&session)
        return res
    }
    //utils.ClearInterface(session)
    return 0
}

func System(req *http.Request) string {
    session, _ := GetStore().Get(req, SessionName)
    if session.Values["system"] == nil {
        session.Values["system"] = ""
    } else {
        //session.Options.MaxAge =  3600
    }
    return session.Values["system"].(string)
}

type SessionInfo struct {
    UserId int64 `json:"user_id"`
}

func GetSessionInfo(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
    var s SessionInfo
    s.UserId = UserId(req)
    jsonData, _ := json.Marshal(s)
    //checkErr(err)
    fmt.Fprint(res, string(jsonData))
}
func Login(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

    req.ParseForm()

    DoLoginLog(UserId(req), 1)
    o := orm.NewOrm()
    o.Using("default")
    session, _ := GetStore().Get(req, SessionName)
    //session.Options.MaxAge =  3600
    user_id := int64(0)

    oldPassword := ""
    loginWithoutPassword := 0
    ldap_id := int64(0)
    err := o.Raw(DbBindReplace("select id,`password`,`login_without_password`,`ldap_id` from users where (email=? or login=?) limit 1"), req.PostForm.Get("email"), req.PostForm.Get("email")).QueryRow(&user_id, &oldPassword, &loginWithoutPassword, &ldap_id)

    log.Println("ldap_id", ldap_id)
    if ldap_id != 0 {
        err := LdapAuthByLdapId(ldap_id, req.PostForm.Get("email"), req.PostForm.Get("password"))
        if err != nil {
            http.Redirect(res, req, GetDomainParamValue(req.Host, "loginpage")+"#invalidloginLdap", 301)
            return
        }
    }
    //log.Println(oldPassword)

    if loginWithoutPassword == 0 {
        oldPasswordByte1 := []byte(oldPassword)
        oldPasswordByte2 := []byte(req.PostForm.Get("password"))
        err = bcrypt.CompareHashAndPassword(oldPasswordByte1, oldPasswordByte2)
    }

    if err != nil {
        fmt.Fprint(res, err)
        http.Redirect(res, req, GetDomainParamValue(req.Host, "loginpage")+"#invalidlogin", 301)
    } else {
        session.Values["user_id"] = user_id
        session.Save(req, res)
        http.Redirect(res, req, GetUsersHomePage(req), 301)
    }
}

func GetUsersHomePage(req *http.Request) string {
    o := orm.NewOrm()
    o.Using("default")
    url := ""
    err := o.Raw(DbBindReplace(`select p.url url from pages p,roles r,user_roles ur
where ur.role_id=r.id and p.id=r.home_page_id and ur.user_id=? limit 1`), UserId(req)).QueryRow(&url)
    if err != nil {
        log.Println("cannot get home page url " + err.Error())
        log.Println("using homepage " + err.Error())
        return GetDomainParamValue(req.Host, "homepage")
    } else {
        log.Println("GetUsersHomePage =" + url)
        return url
    }

}

func GetLanguage2(req *http.Request) string {
    session, _ := GetStore().Get(req, SessionName)
    //log.Println(GetLanguage2, session.Values["lang"])
    if (session.Values["lang"] == nil) || (session.Values["lang"] == "") {
        return "ru"
    }

    return session.Values["lang"].(string)
}

func GetLanguage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
    session, _ := GetStore().Get(req, SessionName)

    //req.ParseForm();

    fmt.Fprint(res, "{\"lang\":\""+session.Values["lang"].(string)+"\"}")
}

func SetLanguage(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

    type tLang struct {
        Lang string `json:"lang"`
    }
    decoder := json.NewDecoder(req.Body)
    var t tLang
    decoder.Decode(&t)
    session, err := GetStore().Get(req, SessionName)
    if err != nil {
        log.Println("SetLanguage Error", err.Error())
    }
    session.Values["lang"] = t.Lang

    session.Save(req, res)

    fmt.Fprint(res, `{"result":true}`)

}

func Logout(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

    user_id := UserId(req)
    DoLoginLog(user_id, 2)
    session, _ := GetStore().Get(req, SessionName)
    session.Values["user_id"] = int64(0)
    session.Values["anonymous_session_id"] = nil

    err := session.Save(req, res)
    if err != nil {
        fmt.Fprint(res, err.Error())
        return
    }

    req.ParseForm()

    if GetDomainParamValue(req.Host, "loginpage") != "" {
        fmt.Fprint(res, `<html><head><meta http-equiv="refresh" content="0; url=`+GetDomainParamValue(req.Host, "loginpage")+`?angularjs_redirecturi=`+req.Form.Get("angularjs_redirecturi")+`" /></head></html>`)
        ChangeUserStatus(user_id, "OFFLINE")
    } else {
        fmt.Fprint(res, `<html><head></head><body><h1>Param loginpage not set</h1></body></html>`)
    }

}

func LogoutExt(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

    DoLoginLog(UserId(req), 2)
    session, _ := GetStore().Get(req, SessionName)
    session.Values["user_id"] = int64(0)
    session.Values["anonymous_session_id"] = nil

    err := session.Save(req, res)
    if err != nil {
        res.WriteHeader(401)
        fmt.Fprint(res, err.Error())
        return
    }

}

func DoLoginLog(user_id int64, login_type int64) {

    //var userLog models.LoginLogs

    //userLog.UserId = user_id
    //userLog.LoginType = login_type

    o := orm.NewOrm()
    o.Using("default")
    o.Raw(DbBindReplace("insert into login_logs (user_id,login_type) values (?,?)"), user_id, login_type).Exec()
    o = nil

}

func ForceResetPassword(email, newPassword string) error {

    log.Println("resetting " + email + " - to " + newPassword)
    o := orm.NewOrm()
    o.Using("default")
    ldap_id := int64(0)
    o.Raw(DbBindReplace("select ldap_id from users where email=?"), email).QueryRow(&ldap_id)

    if ldap_id != 0 {
        err := LdapUserForceResetPassword(ldap_id, email, newPassword)
        return err

    } else {
        password := []byte(newPassword)
        hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
        _, err = o.Raw(DbBindReplace("update users set password_not_set=0,password=? where email=?"), string(hashedPassword), email).Exec()
        return err
    }
}

func ChangeUserStatus(user_id int64, status_code string) {
    o := orm.NewOrm()
    o.Using("default")
    o.Raw(DbBindReplace("UPDATE clc_user_status_hs SET active = 0, closed_at = NOW() WHERE active = 1 AND user_id = ?"), user_id).Exec()
    status_id := int64(0)
    o.Raw(DbBindReplace("SELECT id AS status_id FROM clc_user_statuses WHERE code = ?"), status_code).QueryRow(&status_id)
    o.Raw(DbBindReplace("INSERT INTO clc_user_status_hs (user_id, status_id, created_by, active) VALUES (?, ?, 1, 1)"), user_id, status_id).Exec()
}