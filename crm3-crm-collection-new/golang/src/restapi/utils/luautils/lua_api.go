package luautils

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/stateorm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/kafka"
	lua "github.com/Shopify/go-lua"
	"github.com/clbanning/mxj"
	"github.com/go-gomail/gomail"
	strip "github.com/grokify/html-strip-tags-go"

	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"html"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"

	"path/filepath"

	"git.dar.kz/crediton-3/crm-mfo/src/pkg"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/cp1048"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/gabs"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/xmlpath"
	"github.com/Shopify/goluago"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
	"github.com/xuri/excelize/v2"
)

func LuaBebe(L *lua.State) int {
	L.PushString("HAHAHAHA")

	return 0
}

func LuaQueryEscape(L *lua.State) int {
	s1 := L.ToValue(1).(string)
	L.PushString(url.QueryEscape(s1))
	return 1
}

func LuaHasPrefix(L *lua.State) int {
	s1 := L.ToValue(1).(string)
	s2 := L.ToValue(2).(string)
	L.PushBoolean(strings.HasPrefix(s1, s2))
	return 1
}

func LuaCryptoSignPKCS1v15(L *lua.State) int {

	data, _ := L.ToString(1)
	password, _ := L.ToString(2)
	str, _ := L.ToString(3)

	res, cert, err := utils.CryptoSignPKCS1v15(data, password, str)
	if err != nil {
		L.PushString("")
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString(res)
		L.PushString(cert)
		L.PushString("")
		L.PushInteger(0)
	}
	return 4
}

func LuaDBCurrentDateTime(L *lua.State) int {

	start := time.Now()
	res := start.Format("2006-01-02 15:04:05")
	L.PushString(res)
	return 1
}

func LuaXmltoJSONString(L *lua.State) int {
	s1 := L.ToValue(1).(string)

	val, err := utils.XmltoJSONString(s1)
	if err != nil {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}
	L.PushString(val)
	L.PushString("")
	L.PushInteger(0)
	return 3
}

func LuaStripTags(L *lua.State) int {

	s, _ := L.ToString(1)
	res := strip.StripTags(s)
	res = strings.Replace(res, "&nbsp;", " ", -1)
	L.PushString(res)
	return 1
}

func LuaStrReplace(L *lua.State) int {
	s1 := L.ToValue(1).(string)
	s2 := L.ToValue(2).(string)
	s3 := L.ToValue(3).(string)
	n := L.ToValue(4).(float64)
	L.PushString(strings.Replace(s1, s2, s3, int(n)))

	return 1
}

func LuaParseHTMLTemplate(L *lua.State) int {
	s, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("LuaParseHTMLTemplate err get str 1")
		L.PushInteger(2)
		return 3
	}
	arr, err := PullTable(L, 2)
	if err != nil {
		log.Println("LuaParseHTMLTemplate PullTable", err)
		L.PushNil()
		L.PushString("LuaParseHTMLTemplate err get str 1" + err.Error())
		L.PushInteger(2)
		return 3
	}
	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("LuaParseHTMLTemplate err get str 3")
		L.PushInteger(2)
		return 3
	}
	res, err := utils.ParseHTMLTemplate(L, s, arr, int64(userId))
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	} else {
		L.PushString(res)
		L.PushString("")
		L.PushInteger(0)
		return 3
	}
	return 3
}

func LuaParseTemplate(L *lua.State) int {
	s, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("LuaParseTemplate err get str 1")
		L.PushInteger(2)
		return 3
	}
	arr, err := PullTable(L, 2)
	if err != nil {
		log.Println("LuaParseTemplate PullTable", err)
		L.PushNil()
		L.PushString("LuaParseTemplate err get str 1" + err.Error())
		L.PushInteger(2)
		return 3
	}
	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("LuaParseTemplate err get str 3")
		L.PushInteger(2)
		return 3
	}
	res, err := utils.ParseTemplate(L, s, arr, int64(userId))
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	} else {
		L.PushString(res)
		L.PushString("")
		L.PushInteger(0)
		return 3
	}
	return 3
}

func LuaSendMail2NoTLS(L *lua.State) int {
	m := gomail.NewMessage()

	//1 --String From
	//2 --Array To
	//3 --Subject
	//4 -- Content-Type
	//5 -- Body
	//6 -- Array Files
	//7 -- SMTP Host string
	//8 -- SMTP Port integer
	//9 -- login
	//10 --password

	from, ok1 := L.ToString(1)
	sTo, err1 := PullInterfaceTable(L, 2)

	var arrTo []string

	for _, v := range sTo {
		//req.Header.Set(k, v.(string))
		arrTo = append(arrTo, v.(string))
		log.Println("to added", v.(string))
		//log.Println("key=>", k, "value=>", v)
	}

	//arrTo := InterfaceStringSlice(sTo)
	subject, ok3 := L.ToString(3)
	contentType, ok4 := L.ToString(4)
	body, ok5 := L.ToString(5)
	arrFiles, err6 := PullInterfaceTable(L, 6)
	smtpHost, ok7 := L.ToString(7)
	smtpPort, ok8 := L.ToInteger(8)
	login, ok9 := L.ToString(9)
	password, ok10 := L.ToString(10)
	async := L.ToBoolean(11)

	if !ok1 {
		L.PushString("Error Bind Value from")
		L.PushInteger(3)
		return 2
	}

	if !ok3 {
		L.PushString("Error Bind Value subject")
		L.PushInteger(3)
		return 2
	}

	if !ok4 {
		L.PushString("Error Bind Value contentType")
		L.PushInteger(3)
		return 2
	}

	if !ok5 {
		L.PushString("Error Bind Value body")
		L.PushInteger(3)
		return 2
	}

	if !ok7 {
		L.PushString("Error Bind Value smtpHost")
		L.PushInteger(3)
		return 2
	}

	if !ok8 {
		L.PushString("Error Bind Value smtpPort")
		L.PushInteger(3)
		return 2
	}

	if !ok9 {
		L.PushString("Error Bind Value login")
		L.PushInteger(3)
		return 2
	}

	if !ok10 {
		L.PushString("Error Bind Value password")
		L.PushInteger(3)
		return 2
	}

	if err1 != nil {
		L.PushString(err1.Error())
		L.PushInteger(4)
		return 2
	}

	if err6 != nil {
		L.PushString(err6.Error())
		L.PushInteger(5)
		return 2
	}

	log.Print("arrTo", arrTo)
	m.SetHeader("From", from)
	m.SetHeader("To", arrTo...)

	m.SetHeader("Subject", subject)
	m.SetBody(contentType, body)
	for _, v := range arrFiles {
		m.Attach(v.(string))
		log.Println("file attached", v.(string))
	}

	d := gomail.Dialer{Host: smtpHost, Port: smtpPort, Username: login, Password: password, SSL: false, TLSConfig: nil}

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if async {
		go d.DialAndSend(m)
		L.PushString("")
		L.PushInteger(0)
		return 2
	} else {
		if err := d.DialAndSend(m); err != nil {
			L.PushString(err.Error())
			L.PushInteger(6)
			return 2
		}
	}

	L.PushString("")
	L.PushInteger(0)
	return 2
}

func LuaSendMail2(L *lua.State) int {
	m := gomail.NewMessage()

	//1 --String From
	//2 --Array To
	//3 --Subject
	//4 -- Content-Type
	//5 -- Body
	//6 -- Array Files
	//7 -- SMTP Host string
	//8 -- SMTP Port integer
	//9 -- login
	//10 --password

	from, ok1 := L.ToString(1)
	sTo, err1 := PullInterfaceTable(L, 2)

	var arrTo []string

	for _, v := range sTo {
		//req.Header.Set(k, v.(string))
		arrTo = append(arrTo, v.(string))
		log.Println("to added", v.(string))
		//log.Println("key=>", k, "value=>", v)
	}

	//arrTo := InterfaceStringSlice(sTo)
	subject, ok3 := L.ToString(3)
	contentType, ok4 := L.ToString(4)
	body, ok5 := L.ToString(5)
	arrFiles, err6 := PullInterfaceTable(L, 6)
	smtpHost, ok7 := L.ToString(7)
	smtpPort, ok8 := L.ToInteger(8)
	login, ok9 := L.ToString(9)
	password, ok10 := L.ToString(10)
	async := L.ToBoolean(11)

	if !ok1 {
		L.PushString("Error Bind Value from")
		L.PushInteger(3)
		return 2
	}

	if !ok3 {
		L.PushString("Error Bind Value subject")
		L.PushInteger(3)
		return 2
	}

	if !ok4 {
		L.PushString("Error Bind Value contentType")
		L.PushInteger(3)
		return 2
	}

	if !ok5 {
		L.PushString("Error Bind Value body")
		L.PushInteger(3)
		return 2
	}

	if !ok7 {
		L.PushString("Error Bind Value smtpHost")
		L.PushInteger(3)
		return 2
	}

	if !ok8 {
		L.PushString("Error Bind Value smtpPort")
		L.PushInteger(3)
		return 2
	}

	if !ok9 {
		L.PushString("Error Bind Value login")
		L.PushInteger(3)
		return 2
	}

	if !ok10 {
		L.PushString("Error Bind Value password")
		L.PushInteger(3)
		return 2
	}

	if err1 != nil {
		L.PushString(err1.Error())
		L.PushInteger(4)
		return 2
	}

	if err6 != nil {
		L.PushString(err6.Error())
		L.PushInteger(5)
		return 2
	}

	log.Print("arrTo", arrTo)
	m.SetHeader("From", from)
	m.SetHeader("To", arrTo...)

	m.SetHeader("Subject", subject)
	m.SetBody(contentType, body)
	for _, v := range arrFiles {
		m.Attach(v.(string))
		log.Println("file attached", v.(string))
	}

	d := gomail.NewDialer(smtpHost, smtpPort, login, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if async {
		go d.DialAndSend(m)
		L.PushString("")
		L.PushInteger(0)
		return 2
	} else {
		if err := d.DialAndSend(m); err != nil {
			L.PushString(err.Error())
			L.PushInteger(6)
			return 2
		}
	}

	L.PushString("")
	L.PushInteger(0)
	return 2
}

func LuaJoin(L *lua.State) int {
	s1, err := PullTable(L, 1)
	sep, ok := L.ToString(2)
	if !ok {
		L.PushString("")
		return 1
	}
	if err != nil {
		log.Println("LuaJoin", err)
	}
	s := InterfaceStringSlice(s1)
	L.PushString(strings.Join(s, sep))
	return 1
}

func LuaJsonToString(L *lua.State) int {
	s1, err := PullTable(L, 1)
	if err != nil {
		log.Println("error on LuaJsonToString " + err.Error())
		L.PushString("")
		return 1
	}
	//jsonData, err := json.Marshal(&s1)//315

	json := jsoniter.ConfigCompatibleWithStandardLibrary

	//jsonData, err := ffjson.Marshal(&s1)
	jsonData, err := json.Marshal(&s1)
	if err != nil {
		log.Println("LuaJsonToString1", err)
	}
	result := string(jsonData)
	L.PushString(result)
	utils.ClearInterface(&s1)
	utils.ClearInterface(&jsonData)
	return 1
}

func LuaJsonToXML(L *lua.State) int {
	s1, err := PullTable(L, 1)
	if err != nil {
		log.Println("LuaJsonToString2", err)
	}
	bytes, err := json.Marshal(s1)
	if err != nil {
		log.Println("error LuaJsonToXML " + err.Error())
	}
	m, err := mxj.NewMapJson(bytes)
	data, err := m.Xml()
	L.PushString(string(data))
	return 1
}

func LuaMkdirAll(L *lua.State) int {
	s1 := L.ToValue(1).(string)
	err := os.MkdirAll(s1, 0777)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)

	} else {
		L.PushString("")
		L.PushInteger(0)

	}
	return 2
}

func LuaRegexpFindAllStringsAndJoin(L *lua.State) int {
	expr := L.ToValue(1).(string)
	str := L.ToValue(2).(string)
	cnt, _ := L.ToInteger(3)
	sep := L.ToValue(4).(string)

	re := regexp.MustCompile(expr)
	arr := re.FindAllString(str, cnt)
	L.PushString(strings.Join(arr, sep))
	return 1
}

func LuaBase64Encode(L *lua.State) int {
	s1 := L.ToValue(1).(string)
	b := base64.StdEncoding.EncodeToString([]byte(s1))
	L.PushString(b)
	return 1
}

func LuaParseEmailAddress(L *lua.State) int {
	s1 := L.ToValue(1).(string)

	if strings.HasPrefix(s1, "=?koi8-r") {
		s1 = strings.Replace(s1, "=?koi8-r?B?", "", -1)
		s1 = strings.Replace(s1, "?=", "", -1)

		e, err := mail.ParseAddress(s1)

		if err != nil {

			L.PushString("")
			L.PushString("")
			L.PushString(err.Error())
			L.PushInteger(1)
			return 4

		}

		b, err := base64.StdEncoding.DecodeString(e.Name)
		sr := strings.NewReader(string(b))
		tr := transform.NewReader(sr, charmap.KOI8R.NewDecoder())
		buf, err := ioutil.ReadAll(tr)
		e.Name = string(buf)
		L.PushString(e.Name)
		L.PushString(e.Address)
		L.PushString("")
		L.PushInteger(0)
		return 4

	}

	e, err := mail.ParseAddress(s1)
	if err != nil {

		L.PushString("")
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 4

	}

	//log.Println("endoding", enc)
	L.PushString(e.Name)
	L.PushString(e.Address)
	L.PushString("")
	L.PushInteger(0)
	return 4
}

func LuaBase64Decode(L *lua.State) int {
	s1 := L.ToValue(1).(string)
	b, err := base64.StdEncoding.DecodeString(s1)
	if err != nil {
		L.PushNil()
	} else {
		L.PushString(string(b))

	}
	return 1
}

func LuaStringToJson(L *lua.State) int {

	s1 := L.ToValue(1).(string)
	var face interface{}
	err := json.Unmarshal([]byte(s1), &face)
	if err != nil {
		log.Println("LuaStringToJson", err)
		json.Unmarshal([]byte("{}"), &face)

	}
	//log.Println(face)
	//log.Println(face.(type))
	Open(L)
	DeepPush(L, face)
	//L.PushLightUserData(face)
	return 1
}

func LuaSplit(L *lua.State) int {

	str := L.ToValue(1).(string)
	sep := L.ToValue(2).(string)
	arr := strings.Split(str, sep)
	DeepPush(L, arr)
	return 1
}

func (context *LuaContext) LuaEncodeQRCode(L *lua.State) int {

	str, ok := L.ToString(1)
	if !ok {
		L.PushString("")
		L.PushString("bind err 1")
		L.PushInteger(1)

		return 3

	}

	level, ok := L.ToInteger(2)

	if !ok {
		L.PushString("")
		L.PushString("bind err 2")
		L.PushInteger(1)

		return 3

	}
	size, ok := L.ToInteger(3)

	if !ok {
		L.PushString("")
		L.PushString("bind err 3")
		L.PushInteger(1)

		return 3

	}
	dir, ok := L.ToString(4)

	if !ok {
		L.PushString("")
		L.PushString("bind err 4")
		L.PushInteger(1)

		return 3

	}

	qlevel := qrcode.Medium

	if level == 0 {
		qlevel = qrcode.Low
	} else if level == 1 {
		qlevel = qrcode.Medium
	} else if level == 2 {
		qlevel = qrcode.High
	} else if level == 3 {
		qlevel = qrcode.Highest
	} else {
		qlevel = qrcode.Medium
	}

	data, err := qrcode.Encode(str, qlevel, size)
	if err != nil {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)

		return 3

	}

	uuid, err := utils.UploadRawData(context.o, dir, "qrcode.png", data)

	L.PushString(uuid)
	L.PushString("")
	L.PushInteger(0)

	return 3

}

func (context *LuaContext) LuaCopyFile(L *lua.State) int {

	src, ok := L.ToString(1)
	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind src")
		L.PushInteger(1)
		return 3

	}
	dst, ok := L.ToString(2)
	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind dst")
		L.PushInteger(2)
		return 3

	}

	sz, err := utils.CopyFile(src, dst)
	if err != nil {
		L.PushInteger(0)
		L.PushString(err.Error())
		L.PushInteger(3)
		return 3
	}

	L.PushInteger(int(sz))
	L.PushString("")
	L.PushInteger(0)
	return 3

}

func (context *LuaContext) LuaUploadRawData(L *lua.State) int {

	dir, ok := L.ToString(1)
	if !ok {
		L.PushString("")
		L.PushString("1 err")
		L.PushInteger(1)
		return 3

	}
	fileName, ok := L.ToString(2)
	if !ok {
		L.PushString("")
		L.PushString("2 err")
		L.PushInteger(2)
		return 3

	}
	dataStr, ok := L.ToString(3)

	//log.Println("dataStr = "+dataStr)
	if !ok {
		L.PushString("")
		L.PushString("3 err")
		L.PushInteger(3)
		return 3
	}
	data := []byte(dataStr)
	//log.Println(data)

	uuid, err := utils.UploadRawData(context.o, dir, fileName, data)

	if err != nil {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	} else {
		L.PushString(uuid)
		L.PushString("")
		L.PushInteger(0)
		return 3
	}

}

func (context *LuaContext) LuaDetail(L *lua.State) int {

	id, ok := L.ToInteger(1)

	if !ok {
		L.PushNil()
		L.PushString("LuaDetail error on bind ID")
		L.PushInteger(2)
		return 3
	}
	code, ok := L.ToString(2)
	if !ok {
		L.PushNil()
		L.PushString("3 err")
		L.PushInteger(2)
		return 3
	}
	userId, ok := L.ToInteger(3)
	if !ok {
		L.PushNil()
		L.PushString("4 err")
		L.PushInteger(2)
		return 3
	}

	arr, err := utils.Detail(context.o, id, code, int64(userId))

	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("arr=")
		log.Println(arr)
	}
	//var s []interface{}
	//s = append(s,arr)
	DeepPush(L, arr)

	for i := range arr {
		utils.ClearInterface(&i)
	}

	utils.ClearInterface(&arr)

	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {

		L.PushString("")
		L.PushInteger(0)
	}

	return 3
}

func LuaRuNum2Word(L *lua.State) int {

	number, _ := L.ToString(1)
	upp := L.ToBoolean(2)
	valCode, _ := L.ToString(3)
	s := utils.RuNum2Word(number, upp, valCode)
	L.PushString(s)
	return 1

}

func (context *LuaContext) LuaQuery(L *lua.State) int {

	urlStr := L.ToValue(1).(string)
	userId := L.ToValue(2).(float64)

	_, arr, err := utils.QueryByUrl(context.o, urlStr, "", int64(userId), false, "ru")
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("arr=")
		log.Println(arr)
	}
	DeepPush(L, arr)

	arr = make([]orm.Params, 0)
	//utils.ClearInterface(&arr)
	//
	//arr = nil

	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}
	return 3
}

func LuaImapProcessInBoxToMessage(L *lua.State) int {

	user := L.ToValue(1).(string)
	password := L.ToValue(2).(string)
	hostPort := L.ToValue(3).(string)
	callBack := L.ToValue(4).(string)
	lastCount := L.ToValue(5).(float64)
	lastUid := L.ToValue(6).(float64)
	user_id := L.ToValue(7).(float64)

	maxUid, err := ImapProcessInBoxToMessage(L, callBack,

		lastCount,
		lastUid,
		user, password, hostPort, user_id)

	if err != nil {
		log.Println("LuaImapProcessInBoxToMessage error on load script", err.Error())
		L.PushInteger(int(maxUid))
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}
	L.PushInteger(int(maxUid))
	L.PushString("")
	L.PushInteger(0)

	return 3
}

func (context *LuaContext) LuaQueryWithCount(L *lua.State) int {

	urlStr := L.ToValue(1).(string)
	userId := L.ToValue(2).(float64)

	allCount, arr, err := utils.QueryByUrl(context.o, urlStr, "", int64(userId), true, "ru")
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("arr=")
		log.Println(arr)
	}
	L.PushInteger(int(allCount))
	DeepPush(L, arr)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}
	return 4
}

func LuaReadFile(L *lua.State) int {

	fileName := L.ToValue(1).(string)
	arr, err := ioutil.ReadFile(fileName)
	L.PushString(string(arr))
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}
	return 3
}

func LuaTelegramNewDocumentShare(L *lua.State) int {

	token := L.ToValue(1).(string)
	chatId := L.ToValue(2).(float64)
	fileId := L.ToValue(3).(string)
	//MessageId := L.ToValue(4).(float64)

	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		L.PushString("1")
		log.Println("ERROR TELEGRAM LUA TOKEN")
		log.Println("LuaTelegramNewDocumentShare", err)
		return 1
	}

	msg := tgbotapi.NewDocumentShare(int64(chatId), fileId)
	//msg.ReplyToMessageID = int(MessageId)
	bot.Send(msg)

	L.PushString("0")
	return 1
}

func LuaTelegramNewMessage(L *lua.State) int {

	token := L.ToValue(1).(string)
	chatId := L.ToValue(2).(float64)
	MessageText := L.ToValue(3).(string)
	MessageId := L.ToValue(4).(float64)

	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		L.PushInteger(0)
		log.Println("ERROR TELEGRAM LUA TOKEN")
		log.Println("LuaTelegramNewMessage", err)
		return 1
	}

	msg := tgbotapi.NewMessage(int64(chatId), MessageText)

	/*var keyboard= tgbotapi.NewKeyboardButtonContact("test")
	  var s = []tgbotapi.KeyboardButton{}
	  s = append(s,keyboard)
	  var r = tgbotapi.NewReplyKeyboard(s)
	  msg.ReplyMarkup = r
	*/
	msg.ReplyToMessageID = int(MessageId)
	msg2, _ := bot.Send(msg)

	L.PushInteger(msg2.MessageID)
	return 1
}

func LuaToLower(L *lua.State) int {
	str := L.ToValue(1).(string)
	L.PushString(strings.ToLower(str))
	return 1
}

func (context *LuaContext) LuaEntityValueByCode(L *lua.State) int {

	entityCode := (fmt.Sprintf("%v", L.ToValue(1)))
	entityField := (fmt.Sprintf("%v", L.ToValue(2)))
	Pk := (fmt.Sprintf("%v", L.ToValue(3)))

	//log.Println("WTF", entityCode, entityField, Pk)

	if !utils.CheckEntity(entityCode) {
		log.Println("Error on LuaEntityValueByCode CheckEntity")
		//L.PushNil()
		L.PushString("")
		return 1
	}
	value := ""
	err := context.o.Raw(utils.DbBindReplace("select "+entityField+" from "+entityCode+" where code = ?"), Pk).QueryRow(&value)
	if err != nil {
		log.Println("Error on LuaEntityValueByCode ", err)
		//L.PushNil()
		L.PushString("")
		return 1
	}
	L.PushString(value)
	return 1
}

func (context *LuaContext) LuaEntityValueById(L *lua.State) int {

	entityCode := (fmt.Sprintf("%v", L.ToValue(1)))
	entityField := (fmt.Sprintf("%v", L.ToValue(2)))
	Pk := (fmt.Sprintf("%v", L.ToValue(3)))

	//log.Println("WTF", entityCode, entityField, Pk)

	if !utils.CheckEntity(entityCode) {
		log.Println("Error on LuaEntityValueById err CheckEntity")
		//L.PushNil()
		L.PushString("")
		return 1
	}
	value := ""
	err := context.o.Raw(utils.DbBindReplace("select "+entityField+" from "+entityCode+" where id = ?"), Pk).QueryRow(&value)
	if err != nil {
		log.Println("Error on LuaEntityValueById ", err)
		L.PushString("")
		//L.PushNil()
		return 1
	}
	L.PushString(value)
	return 1
}

func LuaPathEscape(L *lua.State) int {
	res := url.PathEscape(fmt.Sprintf("%v", L.ToValue(1)))
	res = strings.Replace(res, "=", "%3D", -1)
	res = strings.Replace(res, "&", "%26", -1)
	L.PushString(res)
	return 1
}

func LuaHTMLEscapeString(L *lua.State) int {
	res := html.EscapeString(fmt.Sprintf("%v", L.ToValue(1)))

	L.PushString(res)
	return 1
}

func LuaGetParamValue(L *lua.State) int {
	res := utils.GetParamValue(fmt.Sprintf("%v", L.ToValue(1)))
	L.PushString(res)
	return 1
}

func (context *LuaContext) LuaGetUserParamValue(L *lua.State) int {
	user_id, _ := L.ToInteger(1)
	res := utils.GetUserParamValue(context.o, int64(user_id), fmt.Sprintf("%v", L.ToValue(2)))
	L.PushString(res)
	return 1
}

func LuaJSONPathParse(L *lua.State) int {
	jsonParsed, err := gabs.ParseJSON([]byte(L.ToValue(1).(string)))
	if err != nil {
		L.PushString("")
		return 1
	}

	value, ok := jsonParsed.Path(L.ToValue(2).(string)).Data().(string)
	if ok {
		L.PushString(value)
		return 1
	} else {
		L.PushString("")
		return 1
	}
	return 1

}

func (context *LuaContext) LuaDoScriptGetBool(L *lua.State) int {

	script, ok := L.ToString(1)

	if !ok {
		L.PushBoolean(false)
		L.PushString("LuaLoadScript Bind Var Error1")
		L.PushInteger(1)

		return 3
	}

	input, err := PullTable(L, 2)

	if err != nil {
		L.PushBoolean(false)
		L.PushString("LuaLoadScript Bind Var Error2")
		L.PushInteger(1)

		return 3
	}
	variable, ok := L.ToString(3)

	if !ok {
		L.PushBoolean(false)
		L.PushString("LuaLoadScript Bind Var Error3")
		L.PushInteger(1)

		return 3
	}

	DeepPush(L, input)
	L.SetGlobal("input")
	err = lua.DoString(L, script)

	L.Global(variable)
	value := L.ToBoolean(4)

	if err != nil {
		log.Println("LuaDoScript", err)
		debug.PrintStack()
		L.PushBoolean(false)
		L.PushString(err.Error())
		L.PushInteger(1)

		return 3
	} else {
		L.PushBoolean(value)
		L.PushString("")
		L.PushInteger(0)

		return 3
	}
}

func (context *LuaContext) LuaDoScriptGetTable(L *lua.State) int {

	script, ok := L.ToString(1)

	if !ok {
		L.PushNil()
		L.PushString("LuaLoadScript Bind Var Error1")
		L.PushInteger(1)

		return 3
	}

	input, err := PullTable(L, 2)

	if err != nil {
		L.PushNil()
		L.PushString("LuaLoadScript Bind Var Error2")
		L.PushInteger(1)

		return 3
	}
	variable, ok := L.ToString(3)

	if !ok {
		L.PushNil()
		L.PushString("LuaLoadScript Bind Var Error3")
		L.PushInteger(1)

		return 3
	}

	l := lua.NewState()
	lua.OpenLibraries(l)

	//loadLuas(l)

	RegisterAPI(l, context.o)
	RegisterBPMLUaAPI(nil, l, context.o)
	DeepPush(l, input)
	l.SetGlobal("input")
	err = lua.DoString(l, script)

	if err != nil {
		log.Println("LuaDoScript", err)
		debug.PrintStack()
		DeepPush(L, nil)
		L.PushString(err.Error())
		L.PushInteger(1)

		return 3
	}

	l.Global(variable)
	//value := L.ToValue(4)
	value, err := PullTable(l, 1)

	if err != nil {
		log.Println("LuaDoScript", err)
		debug.PrintStack()
		DeepPush(L, nil)
		L.PushString(err.Error())
		L.PushInteger(1)

		return 3
	} else {
		DeepPush(L, value)
		L.PushString("")
		L.PushInteger(0)

		return 3
	}
}

func (context *LuaContext) LuaDoScript(L *lua.State) int {

	script, ok := L.ToString(1)
	input, err := PullTable(L, 2)

	if !ok {
		L.PushString("LuaLoadScript Bind Var Error")
		L.PushInteger(1)
		return 2
	}

	if err != nil {
		L.PushString("LuaLoadScript Bind Var Error")
		L.PushInteger(1)
		return 2
	}

	DeepPush(L, input)
	L.SetGlobal("input")
	err = lua.DoString(L, script)

	if err != nil {
		log.Println("LuaDoScript", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	} else {
		L.PushString("")
		L.PushInteger(0)
		return 2
	}
}

func (context *LuaContext) LuaLoadScript(L *lua.State) int {

	Code, ok := L.ToString(1)

	if !ok {
		L.PushString("LuaLoadScript Bind Var Error")
		L.PushInteger(1)
		return 2
	}

	var script string
	err := context.o.Raw(utils.DbBindReplace("select script from luas where code=?"), Code).QueryRow(&script)
	if err != nil {
		log.Println("LuaLoadScript", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	}

	err = lua.DoString(L, script)

	if err != nil {
		log.Println("LuaLoadScript", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	} else {
		L.PushString("")
		L.PushInteger(0)
		return 2
	}
}

func LuaGetPasswordHash(L *lua.State) int {

	password := []byte((L.ToValue(1).(string)))
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		log.Println("LuaGetPasswordHash", err)
		debug.PrintStack()
	}
	L.PushString(string(hashedPassword))
	return 1
}

func LuaSendWSAsync(L *lua.State) int {

	wsServer, ok := L.ToString(1)
	path, ok := L.ToString(2)
	msg, ok := L.ToString(3)

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	//done := make(chan struct{})
	signal.Notify(interrupt, os.Interrupt)

	if !ok {
		L.PushString("err LuaSendWSAsync")
		L.PushInteger(2)
		return 2
	}

	//var addr = flag.String("addr", wsServer, "http service address")

	u := url.URL{Scheme: "ws", Host: wsServer, Path: path}

	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(3)
		return 2
	}

	//	err = c.WriteMessage(websocket.TextMessage,[]byte(msg))
	//	if err!=nil{
	//		L.PushString(err.Error())
	//		L.PushInteger(4)
	//		return 2
	//	}

	//	err = c.WriteMessage(websocket.PingMessage, []byte("pongpppppppppppppppppp"))
	//	if err!=nil {
	//		L.PushString(err.Error())
	//		L.PushInteger(43)
	//		return 2
	//	}
	log.Println("msg = " + msg)

	err = c.WriteMessage(websocket.BinaryMessage, []byte(msg))
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(42)
		return 2
	}

	//	ticker := time.NewTicker(time.Second)
	//	defer ticker.Stop()
	//
	//	for {
	//		select {
	//		case t := <-ticker.C:
	//			log.Println(t)
	//
	//			if err != nil {
	//				log.Println("write:", err)
	//				L.PushString("err LuaSendWSAsync")
	//				L.PushInteger(2)
	//				return 2
	//			}else{
	//				log.Println("write ok")
	//			}
	//		case <-interrupt:
	//			log.Println("interrupt")
	//		// To cleanly close a connection, a client should send a close
	//		// frame and wait for the server to close the connection.
	//			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	//			if err != nil {
	//				log.Println("write close:", err)
	//				L.PushString("err LuaSendWSAsync")
	//				L.PushInteger(2)
	//				return 2
	//			}
	//				select {
	//				case <-done:
	//				case <-time.After(time.Second):
	//				}
	//			c.Close()
	//			L.PushString("err LuaSendWSAsync")
	//			L.PushInteger(2)
	//			return 2
	//		}
	//	}

	//c.ReadJSON(&u)

	if 1 == 2 {
		L.PushString("err test LuaSendWSAsync")
		L.PushInteger(2)
		return 2
	}

	defer c.Close()
	if err != nil {
		L.PushString("test")
		L.PushInteger(1)
		return 2
	} else {
		L.PushString("")
		L.PushInteger(0)
		return 2
	}

}
func (context *LuaContext) LuaSendEmail(L *lua.State) int {

	channelCode := (L.ToValue(1).(string))
	toText := (L.ToValue(2).(string))
	toEmail := (L.ToValue(3).(string))
	subject := (L.ToValue(4).(string))
	body := (L.ToValue(5).(string))
	wait := (L.ToValue(6).(bool))

	var p []orm.Params
	_, err := context.o.Raw(utils.DbBindReplace("select * from di_chs where code=?"), channelCode).Values(&p)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		var s = utils.TypeEmail{
			FromText:     p[0]["smtp_from_text"].(string),
			FromMail:     p[0]["smtp_from"].(string),
			ToText:       toText,
			ToMail:       toEmail,
			Subject:      subject,
			Body:         body,
			AsHTML:       p[0]["smtp_html"].(string) == "1",
			SMTPLogin:    p[0]["smtp_user"].(string),
			SMTPPassword: p[0]["smtp_password"].(string),
			SMTPServer:   p[0]["smtp_host"].(string) + ":" + p[0]["smtp_port"].(string),
			IsTLS:        (p[0]["smtp_is_tls"].(string) == "1"),
		}

		log.Println(s)

		if wait {
			err := utils.SendEmail(s)
			if err != nil {
				L.PushString(err.Error())
				L.PushInteger(1)
			} else {
				L.PushString("")
				L.PushInteger(0)
			}
		} else {
			go utils.SendEmail(s)
			L.PushString("")
			L.PushInteger(0)
		}

	}
	p = nil

	return 2

}

func LuaGetHTTPListenHostPort(L *lua.State) int {
	L.PushString(os.Getenv("OPENSHIFT_GO_IP") + ":" + os.Getenv("OPENSHIFT_GO_PORT"))
	return 1
}

func LuaRegexpFindStringSubmatch(L *lua.State) int {
	//`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	re := regexp.MustCompile(L.ToValue(2).(string))
	i, _ := L.ToInteger(3)
	match := re.FindStringSubmatch(L.ToValue(1).(string))
	if len(match) > i {
		L.PushString(match[i])
	} else {
		L.PushString("")
	}
	return 1
}

func LuaRegexpCheck(L *lua.State) int {
	//`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	Re := regexp.MustCompile(L.ToValue(2).(string))
	L.PushBoolean(Re.MatchString(L.ToValue(1).(string)))
	return 1
}

func LuaXmlPathParse(L *lua.State) int {

	//path := xmlpath.MustCompile("/Envelope/Body/ttnclosedResponse/return")
	path := xmlpath.MustCompile(L.ToValue(1).(string))
	//root, err := xmlpath.Parse(strings.NewReader(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">        <soap:Header/>        <soap:Body> <m:ttnclosedResponse xmlns:m="crmnewrequest">        <m:return xmlns:xs="http://www.w3.org/2001/XMLSchema"                        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">Cancel</m:return>        <m:answer xmlns:xs="http://www.w3.org/2001/XMLSchema"                        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">Cancel</m:answer></m:ttnclosedResponse></soap:Body></soap:Envelope>`))
	root, err := xmlpath.Parse(strings.NewReader(L.ToValue(2).(string)))

	log.Println("@@@ " + L.ToValue(1).(string))
	log.Println("### " + L.ToValue(2).(string))
	if err != nil {
		log.Println("LuaXmlPathParse", err)
		debug.PrintStack()
		L.PushString("")
		return 1
	}
	if value, ok := path.String(root); ok {
		L.PushString(value)
		fmt.Println("Found:", value)
		return 1
	}
	L.PushString("")
	return 1

}

//err := exec.Command("wkhtmltopdf", htmlFilename, pdfFilename).Run()

func LuaCommand(L *lua.State) int {
	cmd := L.ToValue(1).(string)
	var param []string
	for i := 2; i < 10000000; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {
			arrInterface, err := PullTable(L, i)
			if err != nil {
				log.Println("LuaCommand 1", err)
				L.PushString(err.Error())
				L.PushInteger(1)
				return 2
			}
			arr := InterfaceSlice(arrInterface)
			for _, v := range arr {
				param = append(param, fmt.Sprintf("%v", v))
			}
		} else {
			param = append(param, fmt.Sprintf("%v", L.ToValue(i)))
			//log.Println("added value")
			//log.Println(L.ToValue(i))
		}
	}

	err := exec.Command(cmd, param...).Run()
	if err != nil {
		log.Println("LuaCommand", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}
	L.PushString("")
	L.PushInteger(0)
	return 2

}

func (context *LuaContext) LuaSqlExec(L *lua.State) int {

	sql := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []string
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlExec", err)
				L.PushString(err.Error())
				L.SetGlobal("last_error") //pop
				return 0
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				param = append(param, fmt.Sprintf("%v", v))
			}
		} else {
			//param = append(param,,L.ToValue(i).(string))
			param = append(param, fmt.Sprintf("%v", L.ToValue(i)))
			//log.Println("added value")
		}
	}

	rs, err := context.o.Raw(utils.DbBindReplace(sql), param).Exec()
	if err != nil {
		log.Println("LuaSqlExec 2", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.SetGlobal("last_error") //pop
		return 0
	}
	lastInsertId, err := rs.LastInsertId()

	L.PushString(strconv.Itoa(int(lastInsertId)))
	L.SetGlobal("sql_last_insert_id")

	if err != nil {
		log.Println("LuaSqlExec 3", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.SetGlobal("last_error") //pop
		return 0
	}

	return 0

}

func (context *LuaContext) LuaBeginTransaction(L *lua.State) int {
	err := context.o.Begin()
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}
	return 2
}

func (context *LuaContext) LuaCommitTransaction(L *lua.State) int {

	//defer context.o.Commit()
	err := context.o.Commit()
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}
	return 2
}

func (context *LuaContext) LuaRollbackTransaction(L *lua.State) int {
	err := context.o.Rollback()
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}
	return 2
}

func (context *LuaContext) LuaSqlExec3(L *lua.State) int {
	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlExec2 1", err)
				L.PushString(err.Error())
				L.PushInteger(1)
				return 2
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}
		} else {
			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				//param = append(param, []byte(L.ToValue(i).(string)))
				param = append(param, L.ToValue(i))
				//param = append(param, []byte("123"))
			}
		}
	}

	_, err := context.o.Raw(utils.DbBindReplace(sq), param).Exec()
	if err != nil {
		log.Println("LuaSqlExec2 2", err, sq)
		//debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	}
	//lastInsertId, err := rs.LastInsertId()

	//L.PushString(strconv.Itoa(int(lastInsertId)))
	//L.SetGlobal("sql_last_insert_id")

	if err != nil {
		log.Println("LuaSqlExec2 3", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}

	L.PushString("")
	//L.PushInteger(0)
	L.PushInteger(len(L.ToValue(2).(string)))
	return 2
}

func (context *LuaContext) LuaSqlExec2(L *lua.State) int {
	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlExec2 1", err)
				L.PushString(err.Error())
				L.PushInteger(1)
				return 2
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}
		} else {
			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, fmt.Sprintf("%v", L.ToValue(i)))
			}
		}
	}

	_, err := context.o.Raw(utils.DbBindReplace(sq), param).Exec()
	if err != nil {
		log.Println("LuaSqlExec2 2", err, sq)
		//debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	}
	//lastInsertId, err := rs.LastInsertId()

	//L.PushString(strconv.Itoa(int(lastInsertId)))
	//L.SetGlobal("sql_last_insert_id")

	if err != nil {
		log.Println("LuaSqlExec2 3", err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}

	L.PushString("")
	L.PushInteger(0)
	return 2
}

func (context *LuaContext) LuaSqlCall(L *lua.State) int {

	query, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("Error on LuaSqlCall. Bind Query(1 param)")
		L.PushInteger(1)
		return 3
	}

	arr, err := pullTableRec(L, 2)

	if err != nil {
		log.Println("Error on LuaSqlCall pullTableRec ", err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	}

	b, err := json.Marshal(&arr)

	if err != nil {
		log.Println("Error on LuaSqlCall Marshal ", err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	}

	var r map[string]map[string]interface{}
	err = json.Unmarshal(b, &r)

	pr, err := context.o.Raw(utils.DbBindReplace(query)).Prepare()
	defer pr.Close()
	if err != nil {
		log.Println("Error on LuaSqlCall 1 ", err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	}

	var values []interface{}
	type tresults struct {
		alias  string
		value  string
		input  bool
		output bool
	}
	var results []tresults

	for k, v := range r {

		if v["input"].(bool) && !v["output"].(bool) {
			isclob, ok := v["clob"].(bool)
			if ok && isclob && v["value"] != nil {
				values = append(values, sql.Named(k, []byte(v["value"].(string))))
			} else if v["value"] == nil {
				values = append(values, sql.Named(k, nil))
			} else {
				values = append(values, sql.Named(k, fmt.Sprintf("%v", v["value"])))
			}

		}
		if v["output"] != nil && v["output"].(bool) {
			if v["value"] == nil {
				results = append(results, tresults{alias: k, value: "", input: v["input"].(bool), output: v["output"].(bool)})
			} else {
				results = append(results, tresults{alias: k, value: fmt.Sprintf("%v", v["value"]), input: v["input"].(bool), output: v["output"].(bool)})

			}
		}
	}

	for k, v := range results {
		values = append(values, sql.Named(v.alias, sql.Out{Dest: &results[k].value, In: results[k].input}))
	}

	_, err = pr.Exec(values...)

	if err != nil {
		log.Println("Error on LuaSqlCall 2 Exec ", err)
		L.PushNil()
		L.PushString("Error on LuaSqlCall. Exec" + err.Error())
		L.PushInteger(2)
		return 3
	}

	result := make(map[string]string)
	for _, v := range results {
		result[v.alias] = v.value
	}

	DeepPush(L, result)
	L.PushString("")
	L.PushInteger(0)
	return 3
}

func (context *LuaContext) LuaSqlInsert(L *lua.State) int {
	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlInsert", err)

				L.PushInteger(0)
				L.PushString(err.Error())
				L.PushInteger(1)
				return 3
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, fmt.Sprintf("%v", v))
				}
			}
		} else {
			//param = append(param,,L.ToValue(i).(string))

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, fmt.Sprintf("%v", L.ToValue(i)))
			}

			//param = append(param, fmt.Sprintf("%v", L.ToValue(i)))
			//log.Println("added value")
		}
	}

	lastInsertId, err := utils.DbInsert(context.o, utils.DbBindReplace(sq), param...)
	if err != nil {
		log.Println("Lua SqlInsert Err " + err.Error())
		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("SQL = " + sq)
		}

		L.PushInteger(0)
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	//L.PushString(strconv.Itoa(int(lastInsertId)))
	//L.SetGlobal("sql_last_insert_id")

	if err != nil {
		log.Println("LuaSqlInsert 5", err)
		debug.PrintStack()
		L.PushInteger(0)
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	}

	L.PushInteger(int(lastInsertId))
	L.PushString("")
	L.PushInteger(0)
	return 3
}

func LuaTempDir(L *lua.State) int {
	L.PushString(os.TempDir())
	return 1
}

func LuaTempFile(L *lua.State) int {
	ext, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("err ")
		L.PushInteger(1)
		return 3
	}
	fname, err := utils.TempFile(ext)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	} else {
		L.PushString(fname)
		L.PushString("")
		L.PushInteger(0)
		return 3
	}
}

func LuaWriteFile(L *lua.State) int {
	fileName, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("Error on LuaWriteFile bind filename")
		L.PushInteger(1)
		return 2
	}
	s, ok := L.ToString(2)
	if !ok {
		L.PushString("Error on LuaWriteFile bind s")
		L.PushInteger(1)
		return 2
	}

	err := ioutil.WriteFile(fileName, []byte(s), 0666)

	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}

	L.PushString("")
	L.PushInteger(0)

	return 2
}

func LuaAppendFile(L *lua.State) int {
	fileName, ok := L.ToString(1)
	if !ok {
		L.PushString("LuaAppendFile: err bind filename ")
		L.PushInteger(1)
		return 2
	}
	s, ok := L.ToString(2)
	if !ok {
		L.PushString("err ")
		L.PushInteger(1)
		return 2
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)

	defer file.Close()
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}
	_, err = file.Write([]byte(s))

	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(3)
		return 2
	}

	L.PushString("")
	L.PushInteger(0)
	return 2

}

func (context *LuaContext) LuaSQLQueryRowsExtDb(L *lua.State) int {

	extDbCode := L.ToValue(1).(string)
	query := L.ToValue(2).(string)
	o := orm.NewOrm()
	o.Using("default")

	dbtype := ""
	connstr := ""
	err := o.Raw(utils.DbBindReplace("select dbtype, connstr from extdbs where code=?"), extDbCode).QueryRow(&dbtype, &connstr)
	if err != nil {
		log.Println("LuaSQLQueryRowsExtDb", err)
		L.PushNil()
		L.PushString("ExtDb `" + extDbCode + "` not found: " + err.Error())
		L.PushInteger(1)
		return 3
	}

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 3; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSQLQueryRowsExtDb", err)
				L.PushNil()
				L.PushString(err.Error())
				L.PushInteger(1)
				return 3
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				param = append(param, fmt.Sprintf("%v", v))
			}
		} else {
			//param = append(param,,L.ToValue(i).(string))

			param = append(param, fmt.Sprintf("%v", L.ToValue(i)))
			//log.Println("added value " +fmt.Sprintf("%v",L.ToValue(i)))
		}
	}

	db, err := sql.Open(dbtype, connstr)
	defer db.Close()

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}
	if err = db.Ping(); err != nil {
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	//var p []orm.Params
	//_, err := context.o.Raw(utils.DbBindReplace(sql), param).Values(&p)
	rows, err := db.Query(utils.DbBindReplace(query), param...)

	if err != nil {
		log.Println("LuaSQLQueryRowsExtDb", err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	arrFiles, err := utils.SqlRows2TableDb(rows, query, param)

	//log.Println(query)
	if err != nil {
		log.Println("LuaSQLQueryRowsExtDb", err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		DeepPush(L, arrFiles)
		//L.PushNil()
		L.PushString("")
		L.PushInteger(0)
	}
	return 3
}

func (context *LuaContext) LuaCachedSqlQueryRows(L *lua.State) int {

	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlQueryRows", sq, err)
				L.PushNil()
				L.PushString(err.Error())
				L.PushInteger(1)
				return 3
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}

		} else {
			//param = append(param,,L.ToValue(i).(string))

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, L.ToValue(i))
			}

		}
	}

	var p []orm.Params
	_, err := cached.O().Raw(utils.DbBindReplace(sq), param).Values(&p)

	if err != nil {
		log.Println("LuaCachedSqlQueryRows", utils.DbBindReplace(sq), err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		DeepPush(L, p)
		L.PushString("")
		L.PushInteger(0)
	}
	return 3
}

func (context *LuaContext) LuaCachedSqlQueryRow(L *lua.State) int {

	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlQueryRow1", err)
				L.PushString(err.Error())
				L.PushInteger(1)
				return 2
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}

		} else {

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, L.ToValue(i))
			}
		}
	}
	var p []orm.Params
	_, err := cached.O().Raw(utils.DbBindReplace(sq), param).Values(&p)
	if err != nil {
		log.Println("LuaCachedSqlQueryRow", sq, err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	}

	if len(p) == 0 {
		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("LUA LuaCachedSqlQueryRow No data found", sq, err)
		}
		L.PushString("No data found")

		L.PushInteger(2)
		return 2
	}
	if len(p) > 1 {
		log.Println("LUA LuaCachedSqlQueryRow Too many rows", sq)
		L.PushString("Too many rows")
		L.PushInteger(3)
		return 2
	}

	for key, value := range p[0] {
		//log.Println("@@@@@@@@@"+key)
		//log.Println("@@@@@@@@@"+value.(string))
		if value != nil {
			L.PushString(value.(string))
			L.SetGlobal(key)
		}
	}

	L.PushString("")
	L.PushInteger(0)
	return 2

}

func (context *LuaContext) LuaCachedSqlQueryRow2(L *lua.State) int {

	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaCachedSqlQueryRow2 1 ", err)
				L.PushNil()
				L.PushString(err.Error())
				L.PushInteger(1)
				return 3
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}

		} else {

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, L.ToValue(i))
			}
		}
	}
	var p []orm.Params
	_, err := context.o.Raw(utils.DbBindReplace(sq), param).Values(&p)
	if err != nil {
		log.Println("LuaCachedSqlQueryRow2  2", sq, err)
		debug.PrintStack()
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	if len(p) == 0 {
		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("LUA LuaQueryRow No data found", sq, err)
		}
		L.PushNil()
		L.PushString("No data found")

		L.PushInteger(2)
		return 3
	}
	if len(p) > 1 {
		log.Println("LUA LuaQueryRow Too many rows", sq)
		L.PushNil()
		L.PushString("Too many rows")
		L.PushInteger(3)
		return 3
	}

	L.NewTable()

	for key, value := range p[0] {
		//log.Println("@@@@@@@@@"+key)
		//log.Println("@@@@@@@@@"+value.(string))
		if value != nil {
			L.PushString(value.(string))
			//L.SetGlobal(key)
			L.SetField(-2, key)
		}
	}

	L.PushString("")
	L.PushInteger(0)
	return 3

}

func (context *LuaContext) LuaSqlQueryRows(L *lua.State) int {

	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlQueryRows", sq, err)
				L.PushNil()
				L.PushString(err.Error())
				L.PushInteger(1)
				return 3
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}

		} else {
			//param = append(param,,L.ToValue(i).(string))

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, L.ToValue(i))
			}

		}
	}

	var p []orm.Params
	_, err := context.o.Raw(utils.DbBindReplace(sq), param).Values(&p)

	if err != nil {
		log.Println("LuaSqlQueryRows", utils.DbBindReplace(sq), err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		DeepPush(L, p)
		L.PushString("")
		L.PushInteger(0)
	}
	return 3
}

func (context *LuaContext) LuaSqlQueryRow(L *lua.State) int {

	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlQueryRow1", err)
				L.PushString(err.Error())
				L.PushInteger(1)
				return 2
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}

		} else {

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, L.ToValue(i))
			}
		}
	}
	var p []orm.Params
	_, err := context.o.Raw(utils.DbBindReplace(sq), param).Values(&p)
	if err != nil {
		log.Println("LuaSqlQueryRow5", sq, err)
		debug.PrintStack()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	}

	if len(p) == 0 {
		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("LUA LuaQueryRow No data found", sq, err)
		}
		L.PushString("No data found")

		L.PushInteger(2)
		return 2
	}
	if len(p) > 1 {
		log.Println("LUA LuaQueryRow Too many rows", sq)
		L.PushString("Too many rows")
		L.PushInteger(3)
		return 2
	}

	for key, value := range p[0] {
		//log.Println("@@@@@@@@@"+key)
		//log.Println("@@@@@@@@@"+value.(string))
		if value != nil {
			L.PushString(value.(string))
			L.SetGlobal(key)
		}
	}

	L.PushString("")
	L.PushInteger(0)
	return 2

}

func (context *LuaContext) LuaSqlQueryRow2(L *lua.State) int {

	sq := L.ToValue(1).(string)

	//log.Println("sql = "+sql)
	var param []interface{}
	for i := 2; i < 100; i++ {
		if L.IsNone(i) {
			break
		} else if L.IsTable(i) {

			arrInterface, err := PullTable(L, i)

			if err != nil {
				log.Println("LuaSqlQueryRow2 1", err)
				L.PushNil()
				L.PushString(err.Error())
				L.PushInteger(1)
				return 3
			}

			arr := InterfaceSlice(arrInterface)

			for _, v := range arr {
				if v == nil {
					param = append(param, sql.NullString{})

				} else {
					param = append(param, v)
				}
			}

		} else {

			if L.IsNil(i) {
				param = append(param, sql.NullString{})
			} else {
				param = append(param, L.ToValue(i))
			}
		}
	}
	var p []orm.Params
	_, err := context.o.Raw(utils.DbBindReplace(sq), param).Values(&p)
	if err != nil {
		log.Println("LuaSqlQueryRow2 2", sq, err)
		debug.PrintStack()
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	if len(p) == 0 {
		if os.Getenv("CRM_DEBUG_SQL") == "1" {
			log.Println("LUA LuaQueryRow No data found", sq, err)
		}
		L.PushNil()
		L.PushString("No data found")

		L.PushInteger(2)
		return 3
	}
	if len(p) > 1 {
		log.Println("LUA LuaQueryRow Too many rows", sq)
		L.PushNil()
		L.PushString("Too many rows")
		L.PushInteger(3)
		return 3
	}

	L.NewTable()

	for key, value := range p[0] {
		//log.Println("@@@@@@@@@"+key)
		//log.Println("@@@@@@@@@"+value.(string))
		if value != nil {
			L.PushString(value.(string))
			//L.SetGlobal(key)
			L.SetField(-2, key)
		}
	}

	L.PushString("")
	L.PushInteger(0)
	return 3

}

func LuaHttpPost2(L *lua.State) int {
	//log.Println("req = "+L.ToValue(3).(string))

	sTimeOut := os.Getenv("CRM_HTTP_TIMEOUT_MS")

	timeOut, err := strconv.Atoi(sTimeOut)
	if err != nil {
		timeOut = 30000
	}

	timeout := time.Duration(timeOut) * time.Millisecond

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	req, _ := http.NewRequest("POST", L.ToValue(1).(string), strings.NewReader(L.ToValue(3).(string)))

	arrInterface, err := PullInterfaceTable(L, 2)
	//arr := InterfaceStringSlice(arrInterface)

	for k, v := range arrInterface {
		req.Header.Set(k, v.(string))
		//log.Println("key=>", k, "value=>", v)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("LuaHttpPost21", err)
		debug.PrintStack()
		L.PushString("")
		return 1
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		L.PushString(string(response))
		//log.Println("otvet "+string(response))
		return 1
	} else {
		L.PushString("")
		return 1
	}
	L.PushString("")
	return 1

}

func LuaHttpPost(L *lua.State) int {
	//log.Println("req = "+L.ToValue(3).(string))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	timeout := time.Duration(30000 * time.Second)
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	resp, err := client.Post(L.ToValue(1).(string), L.ToValue(2).(string), strings.NewReader(L.ToValue(3).(string)))
	if err != nil {
		log.Println("LuaHttpPost1", err)
		debug.PrintStack()
		L.PushString("")
		return 1
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		L.PushString(string(response))
		//log.Println("otvet "+string(response))
		return 1
	} else {
		L.PushString("")
		return 1
	}
	L.PushString("")
	return 1

}

func LuaHttpsGet(L *lua.State) int {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	//client.
	resp, err := client.Get(L.ToValue(1).(string))
	//(resp,"ClientId=IIDFQOL0MHLNSVPFWEQ;")
	log.Println(resp)

	if resp.StatusCode != 200 {
		fmt.Println(err)
		L.PushString("")
		L.PushString(resp.Status)
		L.PushInteger(resp.StatusCode)
		return 3
	}
	if err != nil {
		fmt.Println(err)
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	response, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		L.PushString(string(response))
		L.PushString("")
		L.PushInteger(0)
		return 3
	} else {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}
	L.PushString("")
	return 1
}

func LuaHttpsGet2(L *lua.State) int {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", L.ToValue(1).(string), strings.NewReader(""))

	arrInterface, err := PullInterfaceTable(L, 2)
	//arr := InterfaceStringSlice(arrInterface)

	for k, v := range arrInterface {
		req.Header.Set(k, v.(string))
		log.Println("key=>", k, "value=>", v)
	}

	resp, err := client.Do(req)

	log.Println(resp)

	if resp.StatusCode != 200 {
		fmt.Println(err)
		L.PushString("")
		L.PushString(resp.Status)
		L.PushInteger(resp.StatusCode)
		return 3
	}
	if err != nil {
		fmt.Println(err)
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}

	response, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		L.PushString(string(response))
		L.PushString("")
		L.PushInteger(0)
		return 3
	} else {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	}
	L.PushString("")
	return 1
}

func LuaHttpGet(L *lua.State) int {
	resp, err := http.Get(L.ToValue(1).(string))
	if err != nil {
		L.PushString("")
		log.Println("LuaHttpGet", err)
		return 1
	}

	response, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		L.PushString(string(response))
		return 1
	} else {
		L.PushString("")
		return 1
	}
	L.PushString("")
	return 1
}

func LuaHttpGet2(L *lua.State) int {

	timeout := time.Duration(30000 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	req, _ := http.NewRequest("GET", L.ToValue(1).(string), strings.NewReader(""))

	arrInterface, err := PullInterfaceTable(L, 2)
	//arr := InterfaceStringSlice(arrInterface)

	for k, v := range arrInterface {
		req.Header.Set(k, v.(string))
		log.Println("key=>", k, "value=>", v)
	}

	resp, err := client.Do(req)

	if err != nil {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
		log.Println("LuaHttpGet", err)
		return 3
	}

	response, err := ioutil.ReadAll(resp.Body)

	//resp.Header.

	if err == nil {
		L.PushString(string(response))
		L.PushString("")
		L.PushInteger(0)
		return 3
	} else {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(2)
		return 3
	}
}

func LuaLdapCreateUser(L *lua.State) int {
	ldapId, _ := L.ToInteger(1)
	userName, _ := L.ToString(2)
	password, _ := L.ToString(3)

	err := utils.LdapCreateUser(int64(ldapId), userName, password, "")
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}

	return 2
}

func LuaUUID(L *lua.State) int {
	L.PushString(utils.Uuid())
	return 1
}

func LuaStrTrimSpace(L *lua.State) int {
	str, _ := L.ToString(1)
	L.PushString(strings.TrimSpace(str))
	return 1
}

func (context *LuaContext) LuaColumnsByQuery(L *lua.State) int {
	sql, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("Error Bind SQL")
		L.PushInteger(1)
		return 3
	}

	sql = strings.Replace(sql, "?", "-1", -1)
	columnNames, err := context.o.Raw(sql).Columns()
	if err != nil {
		log.Println("Error on LuaColumnsByQuery 2", err)
		L.PushNil()
		L.PushString(err.Error())
		L.PushInteger(1)

		return 3
	}

	DeepPush(L, columnNames)
	L.PushString("")
	L.PushInteger(0)
	return 3

}
func (context *LuaContext) LuaCheckGrantOfEntity(L *lua.State) int {

	user_id, ok := L.ToInteger(1)
	if !ok {
		L.PushString("Error Bind user_id")
		L.PushInteger(1)
		return 2
	}

	TableName, _ := L.ToString(2)
	if !ok {
		L.PushString("Error Bind tablename")
		L.PushInteger(1)
		return 2
	}
	grant, _ := L.ToString(3)
	if !ok {
		L.PushString("Error Bind grant")
		L.PushInteger(1)
		return 2
	}
	id, ok := L.ToInteger(4)
	if !ok {
		L.PushString("Error Bind Id")
		L.PushInteger(1)
		return 2
	}

	err := utils.CheckGrantOfEntity(context.o, int64(user_id), TableName, grant, int64(id))
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}

	L.PushString("")
	L.PushInteger(0)
	return 2
}

func (context *LuaContext) LuaAccPostByPkOper(L *lua.State) int {
	pk, ok := L.ToInteger(1)
	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind PK")
		L.PushInteger(1)
		return 3
	}
	operCode, _ := L.ToString(2)

	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind operCode")
		L.PushInteger(1)
		return 3
	}
	moveId, err := utils.AccPostByPkOper(context.o, int64(pk), operCode, true, "")
	if err != nil {
		L.PushInteger(0)
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	} else {
		L.PushInteger(int(moveId))
		L.PushString("")
		L.PushInteger(0)
		return 3
	}

	return 3
}

func (context *LuaContext) LuaAccPostByPkNoOper(L *lua.State) int {
	pk, ok := L.ToInteger(1)
	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind PK")
		L.PushInteger(1)
		return 3
	}
	operCode, _ := L.ToString(2)

	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind operCode")
		L.PushInteger(1)
		return 3
	}

	date, _ := L.ToString(3)

	if !ok {
		L.PushInteger(0)
		L.PushString("Error Bind Date")
		L.PushInteger(1)
		return 3
	}

	moveId, err := utils.AccPostByPkOper(context.o, int64(pk), operCode, false, date)
	if err != nil {
		L.PushInteger(0)
		L.PushString(err.Error())
		L.PushInteger(1)
		return 3
	} else {
		L.PushInteger(int(moveId))
		L.PushString("")
		L.PushInteger(0)
		return 3
	}

	return 3
}

func (context *LuaContext) LuaAccUndoByPkNoOper(L *lua.State) int {
	entityId, ok := L.ToInteger(1)
	if !ok {
		L.PushString("Error Bind PK")
		L.PushInteger(1)
		return 2
	}

	pk, ok := L.ToInteger(2)
	if !ok {
		L.PushString("Error Bind PK")
		L.PushInteger(1)
		return 2
	}

	err := utils.AccUndoByEntityIdPk(context.o, int64(entityId), int64(pk))
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
		return 2
	} else {
		L.PushString("")
		L.PushInteger(0)
		return 2
	}

	return 2
}

func LuaForceResetPassword(L *lua.State) int {
	log.Println("ForceResetPassword")
	email, _ := L.ToString(1)
	password, _ := L.ToString(2)
	err := utils.ForceResetPassword(email, password)

	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(1)
	} else {
		L.PushString("")
		L.PushInteger(0)
	}

	return 1
}

//----------kafka {
func LuaWriteLog(L *lua.State) int {
	str, _ := L.ToString(1)
	fmt.Println(str)
	return 0
}

func LuaKafkaReaderList(L *lua.State) int {
	arr := kafka.KafkaReaderList()
	L.PushString("")
	L.PushInteger(0)
	DeepPush(L, arr)
	return 3
}
func LuaKafkaReaderClose(L *lua.State) int {
	readerID, _ := L.ToString(1)
	kafka.KafkaReaderClose(readerID)
	L.PushString("")
	L.PushInteger(0)
	return 2
}
func LuaKafkaReaderCreate(L *lua.State) int {
	readerID, _ := L.ToString(1)
	host, _ := L.ToString(2)
	topic, _ := L.ToString(3)
	offset, _ := L.ToInteger(4)
	cb, _ := L.ToString(5)
	confString, _ := L.ToString(6)

	go kafka.KafkaReaderCreate(readerID, host, topic, int64(offset), func(offset int64, data string) {
		L.Global(cb)
		L.PushString(confString)
		L.PushInteger(int(offset))
		L.PushString(data)
		L.Call(3, 0)
	})
	L.PushString("")
	L.PushInteger(0)
	return 2
}
func LuaKafkaWriteMessage(l *lua.State) int {
	broker := lua.CheckString(l, 1)
	topic := lua.CheckString(l, 2)
	msg := lua.CheckString(l, 3)
	err := kafka.KafkaWriteMessage(broker, topic, msg)
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 2
	}
	l.PushString("")
	l.PushInteger(0)
	return 2
}

//-----------}kafka

func LuaClearCache(L *lua.State) int {
	cached.ClearCache()
	return 0
}

func LuaVersion(L *lua.State) int {
	L.PushString(utils.Version)
	return 1
}

func LuaLoadString(L *lua.State) int {
	str, ok := L.ToString(1)
	if !ok {
		L.PushString("Error on Bind 1 param")
		L.PushInteger(1)
		return 2
	}

	err := lua.LoadString(L, str)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	} else {
		L.PushString("")
		L.PushInteger(0)
		return 2
	}

}

func LuaLoadStringNamed(L *lua.State) int {
	str := lua.CheckString(L, 1)
	name := lua.CheckString(L, 2)
	err := lua.LoadBuffer(L, str, name, "")
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	} else {
		L.PushString("")
		L.PushInteger(0)
		return 2
	}

}

func LuaMqSend(L *lua.State) int {

	ConnStr, ok := L.ToString(1)
	if !ok {
		L.PushString("error on bind ConnStr")
		L.PushInteger(1)
		return 2
	}
	Queue, ok := L.ToString(2)
	if !ok {
		L.PushString("error on bind Queue")
		L.PushInteger(1)
		return 2
	}
	Exchange, ok := L.ToString(3)
	if !ok {
		L.PushString("error on bind Exchange")
		L.PushInteger(1)
		return 2
	}

	routingKey, ok := L.ToString(4)
	if !ok {
		L.PushString("error on bind routingKey")
		L.PushInteger(1)
		return 2
	}
	Body, ok := L.ToString(5)
	if !ok {
		L.PushString("error on bind Body")
		L.PushInteger(1)
		return 2
	}

	durable := L.ToBoolean(6)

	async := L.ToBoolean(7)

	if async {
		go TestSend(ConnStr, Queue, Exchange, routingKey, Body, durable)
		L.PushString("")
		L.PushInteger(0)
		return 2
	}

	err := TestSend(ConnStr, Queue, Exchange, routingKey, Body, durable)
	if err != nil {
		L.PushString(err.Error())
		L.PushInteger(2)
		return 2
	}
	L.PushString("")
	L.PushInteger(0)
	return 2
}

func LuaMqReceive(L *lua.State) int {

	ConnStr, ok := L.ToString(1)
	if !ok {
		L.PushString("error on bind ConnStr")
		L.PushInteger(1)
		return 2
	}
	Queue, ok := L.ToString(2)
	if !ok {
		L.PushString("error on bind Queue")
		L.PushInteger(1)
		return 2
	}
	Exchange, ok := L.ToString(3)
	if !ok {
		L.PushString("error on bind Exchange")
		L.PushInteger(1)
		return 2
	}
	RoutingKey, ok := L.ToString(4)
	if !ok {
		L.PushString("error on bind RoutingKey")
		L.PushInteger(1)
		return 2
	}

	durable := L.ToBoolean(5)
	if !ok {
		L.PushString("error on bind durable")
		L.PushInteger(1)
		return 2
	}

	autoDelete := L.ToBoolean(6)
	if !ok {
		L.PushString("error on bind autodelete")
		L.PushInteger(1)
		return 2
	}

	autoAck := L.ToBoolean(7)
	if !ok {
		L.PushString("error on bind auto-ack")
		L.PushInteger(1)
		return 2
	}

	prefetchCount, ok := L.ToInteger(8)
	if !ok {
		L.PushString("error on bind prefetch-count")
		L.PushInteger(1)
		return 2
	}

	cb, _ := L.ToString(9)

	go TestReceive(L, ConnStr, Queue, Exchange, RoutingKey, durable, autoDelete, autoAck, prefetchCount, cb)

	L.PushString("")
	L.PushInteger(0)
	return 2
}

func LuaVersionNum(L *lua.State) int {
	L.PushInteger(utils.VersionNum)
	return 1
}

func (context *LuaContext) LuaBPMNPublish(L *lua.State) int {

	processId, ok := L.ToInteger(1)
	if !ok {

		L.PushString("LuaBPMNPublish. Error on Bind processId ")
		L.PushInteger(1)
		return 2
	}

	bpmGencontext := utils.BpmGenContext{}
	bpmGencontext.O = context.o
	err := bpmGencontext.Publish(int64(processId))

	if err != nil {

		L.PushString("BPMNPublish. Error on Publish " + err.Error())
		L.PushInteger(2)
		return 2
	}

	L.PushString("")
	L.PushInteger(0)
	return 2

}

func LuaTimeParseFormat(L *lua.State) int {
	input, ok := L.ToString(1)
	if !ok {
		L.PushString("")
		return 1
	}
	inputLayout, ok := L.ToString(2)
	if !ok {
		L.PushString("")
		return 1
	}
	outputLayout, ok := L.ToString(3)
	if !ok {
		L.PushString("")
		return 1
	}
	t, _ := time.Parse(inputLayout, input)
	fmt.Println(t) // 2017-08-31 00:00:00 +0000 UTC
	result := t.Format(outputLayout)

	L.PushString(result)
	return 1
}

func LuaTimeParseUnix(L *lua.State) int {
	input, ok := L.ToString(1)
	if !ok {
		L.PushInteger(0)
		return 1
	}
	inputLayout, ok := L.ToString(2)
	if !ok {
		L.PushInteger(0)
		return 1
	}

	t, _ := time.Parse(inputLayout, input)
	fmt.Println(t) // 2017-08-31 00:00:00 +0000 UTC

	L.PushInteger(int(t.Unix()))
	return 1
}

func LuaSleep(L *lua.State) int {
	interval, ok := L.ToInteger(1)
	if ok {
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	return 0
}

func LuaHostname(L *lua.State) int {
	name, err := os.Hostname()
	if err == nil {
		L.PushString(name)
		L.PushString("")
		L.PushInteger(0)
	} else {
		L.PushString("")
		L.PushString(err.Error())
		L.PushInteger(1)
	}
	return 3
}

//-----read dir{
func IoutilReadDir(path string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}
func LuaIoutilReadDir(l *lua.State) int {
	path := lua.CheckString(l, 1)
	files, err := IoutilReadDir(path)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
	}
	DeepPush(l, files)
	l.PushString("")
	l.PushInteger(0)

	return 3
}

func LuaFilepathGlob(l *lua.State) int {
	path := lua.CheckString(l, 1)
	files, err := filepath.Glob(path)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
	}
	DeepPush(l, files)
	l.PushString("")
	l.PushInteger(0)
	return 3
}

//-----}read dir
func LuaFromCP1048(l *lua.State) int {
	s := lua.CheckString(l, 1)
	dec := cp1048.CodePage1048.NewDecoder()
	r, err := dec.Bytes([]byte(s))
	l.PushString(string(r))
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString("")
	l.PushInteger(0)
	return 3
}
func LuaToCP1048(l *lua.State) int {
	s := lua.CheckString(l, 1)
	enc := cp1048.CodePage1048.NewEncoder()
	r, err := enc.Bytes([]byte(s))
	l.PushString(string(r))
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString("")
	l.PushInteger(0)
	return 3
}

func (context *LuaContext) CSVRead(L *lua.State) int {

	str, ok := L.ToString(1)
	if !ok {
		L.PushNil()
		L.PushString("Error on CSVRead. Bind str(1 param)")
		L.PushInteger(1)
		return 3
	}

	comma, ok := L.ToString(2)
	if !ok {
		L.PushNil()
		L.PushString("Error on CSVRead. Bind comma(2 param)")
		L.PushInteger(1)
		return 3
	}

	r := csv.NewReader(strings.NewReader(str))
	if len([]rune(comma)) == 0 {
		L.PushNil()
		L.PushString("Error on CSVRead 2. Bind comma(2 param)")
		L.PushInteger(1)
		return 3
	}
	r.Comma = []rune(comma)[0]

	records, err := r.ReadAll()
	if err != nil {
		L.PushNil()
		L.PushString("Error on CSVRead." + err.Error())
		L.PushInteger(1)
		return 3
	}

	DeepPush(L, records)
	L.PushString("")
	L.PushInteger(0)
	return 3

}

func (context *LuaContext) LuaFileContent(L *lua.State) int {

	file_idStr := fmt.Sprintf("%v", L.ToValue(1))

	file_id, err := strconv.Atoi(file_idStr)

	if err != nil {

		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContent Invalid FileId - " + err.Error())
		L.PushInteger(1)
		return 5
	}

	user_idStr := fmt.Sprintf("%v", L.ToValue(2))

	user_id, err := strconv.Atoi(user_idStr)

	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContent Invalid UserId - " + err.Error())
		L.PushInteger(1)
		return 5
	}

	path := "unix_path"
	if runtime.GOOS == "windows" {
		path = "win_path"
	}

	fullfileName := ""
	fileName := ""
	contentType := ""
	on_read_script := ""
	access_control := int64(0)

	sql := `select coalesce(d.access_control,0), f.id, d.on_read_script,
	(select mime from file_types where ext= lower(right(f.filename,4)) ) contenttype,
	 concat(d.` + path + `,coalesce(nullif(f.filepath,''),f.code) ) fullfileName,f.filename 
	 from files f,dirs d where d.id=f.dir_id and f.id=?`

	if utils.GetDbDriverType() == orm.DROracle {
		sql = `select coalesce(d.access_control,0), f.id, d.on_read_script,
		(select mime from file_types where ext= lower(substr(f.filename,-4)) ) contenttype, 
		d.` + path + `||coalesce(f.filepath,f.code)  fullfileName,f.filename 
		from files f,dirs d where d.id=f.dir_id and f.id=?`

	}
	err = context.o.Raw(utils.DbBindReplace(sql), file_id).QueryRow(&access_control, &file_id, &on_read_script, &contentType, &fullfileName, &fileName)
	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContent" + err.Error())
		L.PushInteger(1)
		return 5
	}

	if access_control != 0 {
		if !utils.DetailEntityGrantCheck(context.o, "files", int64(file_id), int64(user_id)) {
			err = errors.New("Access Denied To File")
			L.PushNil()
			L.PushString("")
			L.PushString("")
			L.PushString("Error on LuaFileContent " + err.Error())
			L.PushInteger(1)
			return 5
		}
	}

	b, err := ioutil.ReadFile(fullfileName)

	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContent ReadFile " + err.Error())
		L.PushInteger(1)
		return 5
	}

	L.PushString(fileName)
	L.PushString(contentType)
	L.PushString(string(b))
	L.PushString("")
	L.PushInteger(0)
	return 5

}

func (context *LuaContext) LuaFileContentNew(L *lua.State) int {

	file_idStr := fmt.Sprintf("%v", L.ToValue(1))

	file_id, err := strconv.Atoi(file_idStr)

	if err != nil {

		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContentNew Invalid FileId - " + err.Error())
		L.PushInteger(1)
		return 5
	}

	user_idStr := fmt.Sprintf("%v", L.ToValue(2))

	user_id, err := strconv.Atoi(user_idStr)

	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContentNew Invalid UserId - " + err.Error())
		L.PushInteger(1)
		return 5
	}

	path := "unix_path"
	if runtime.GOOS == "windows" {
		path = "win_path"
	}

	fullfileName := ""
	fileName := ""
	contentType := ""
	on_read_script := ""
	access_control := int64(0)

	sql := `select coalesce(d.access_control,0), f.id, d.on_read_script,
	(select mime from file_types where ext= lower(right(f.filename,4)) ) contenttype,
	 concat(d.` + path + `,coalesce(nullif(f.filepath,''),f.code) ) fullfileName,f.filename 
	 from files f,dirs d where d.id=f.dir_id and f.id=?`

	if utils.GetDbDriverType() == orm.DROracle {
		sql = `select coalesce(d.access_control,0), f.id, d.on_read_script,
		(select mime from file_types where ext= lower(substr(f.filename,-4)) ) contenttype, 
		d.` + path + `||coalesce(f.filepath,f.code)  fullfileName,f.filename 
		from files f,dirs d where d.id=f.dir_id and f.id=?`

	}
	err = context.o.Raw(utils.DbBindReplace(sql), file_id).QueryRow(&access_control, &file_id, &on_read_script, &contentType, &fullfileName, &fileName)
	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContentNew" + err.Error())
		L.PushInteger(1)
		return 5
	}

	if access_control != 0 {
		if !utils.DetailEntityGrantCheck(context.o, "files", int64(file_id), int64(user_id)) {
			err = errors.New("Access Denied To File")
			L.PushNil()
			L.PushString("")
			L.PushString("")
			L.PushString("Error on LuaFileContentNew " + err.Error())
			L.PushInteger(1)
			return 5
		}
	}
  
  f, err := os.Open(fullfileName)
	if err != nil {
    L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContentNew open file " + err.Error())
		L.PushInteger(1)
		return 5
	}
  defer f.Close()
  decoder := charmap.Windows1251.NewDecoder()
	reader := decoder.Reader(f)
  b, err := ioutil.ReadAll(reader)
	if err != nil {
    L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on LuaFileContentNew ReadAll" + err.Error())
		L.PushInteger(1)
		return 5
	}

	L.PushString(fileName)
	L.PushString(contentType)
	L.PushString(string(b))
	L.PushString("")
	L.PushInteger(0)
	return 5

}

func (context *LuaContext) XlsxRead(L *lua.State) int {

	file_idStr := fmt.Sprintf("%v", L.ToValue(1))

	file_id, err := strconv.Atoi(file_idStr)

	if err != nil {

		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on XlsxRead Invalid FileId - " + err.Error())
		L.PushInteger(1)
		return 5
	}

	user_idStr := fmt.Sprintf("%v", L.ToValue(2))

	user_id, err := strconv.Atoi(user_idStr)

	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on XlsxRead Invalid UserId - " + err.Error())
		L.PushInteger(1)
		return 5
	}

	path := "unix_path"
	if runtime.GOOS == "windows" {
		path = "win_path"
	}

	fullfileName := ""
	fileName := ""
	contentType := ""
	on_read_script := ""
	access_control := int64(0)

	sql := `select coalesce(d.access_control,0), f.id, d.on_read_script,
	(select mime from file_types where ext= lower(right(f.filename,4)) ) contenttype,
	 concat(d.` + path + `,coalesce(nullif(f.filepath,''),f.code) ) fullfileName,f.filename 
	 from files f,dirs d where d.id=f.dir_id and f.id=?`

	if utils.GetDbDriverType() == orm.DROracle {
		sql = `select coalesce(d.access_control,0), f.id, d.on_read_script,
		(select mime from file_types where ext= lower(substr(f.filename,-4)) ) contenttype, 
		d.` + path + `||coalesce(f.filepath,f.code)  fullfileName,f.filename 
		from files f,dirs d where d.id=f.dir_id and f.id=?`

	}
	err = context.o.Raw(utils.DbBindReplace(sql), file_id).QueryRow(&access_control, &file_id, &on_read_script, &contentType, &fullfileName, &fileName)
	if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on XlsxRead" + err.Error())
		L.PushInteger(1)
		return 5
	}

	if access_control != 0 {
		if !utils.DetailEntityGrantCheck(context.o, "files", int64(file_id), int64(user_id)) {
			err = errors.New("Access Denied To File")
			L.PushNil()
			L.PushString("")
			L.PushString("")
			L.PushString("Error on XlsxRead " + err.Error())
			L.PushInteger(1)
			return 5
		}
	}

  f, err := excelize.OpenFile(fullfileName)
  if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on XlsxRead OpenFile " + err.Error())
		L.PushInteger(1)
		return 5
	}
  
  firstSheet := f.WorkBook.Sheets.Sheet[0].Name
  
  results := make([][]string, 0, 64)
  rows, err := f.GetRows(firstSheet)
  if err != nil {
		L.PushNil()
		L.PushString("")
		L.PushString("")
		L.PushString("Error on XlsxRead GetRows" + err.Error())
		L.PushInteger(1)
		return 5
	}
  for _, row := range rows {
      results = append(results, row)
  }
  DeepPush(L, results)
	L.PushString("")
	L.PushInteger(0)
	return 3
}

func (context *LuaContext) AccClsPublish(L *lua.State) int {

	accClsId, ok := L.ToInteger(1)

	if !ok {
		L.PushString("Error on AccClsPublish. Bind accClsId(1 param)")
		L.PushInteger(1)
		return 2
	}

	err := utils.AccClsPublish(int64(accClsId))
	if err != nil {
		L.PushString("Error on AccClsPublish. " + err.Error())
		L.PushInteger(2)
		return 2
	}

	L.PushString("")
	L.PushInteger(0)
	return 2

}

func RegisterAPI(l *lua.State, o orm.Ormer) error {

	lua.OpenLibraries(l)
	//loadLuas(l)
	Open(l)

	luaContext := LuaContext{}
	luaContext.o = o
	//-------------
	l.Register("ToCP1048", LuaToCP1048)
	l.Register("FromCP1048", LuaFromCP1048)
	//-------------
	//read dir{
	l.Register("FilepathGlob", LuaFilepathGlob)
	l.Register("IoutilReadDir", LuaIoutilReadDir)
	//}read dir
	//-------kafka{
	l.Register("KafFkaReaderList", LuaKafkaReaderList)
	l.Register("KafkaReaderClose", LuaKafkaReaderClose)
	l.Register("KafkaReaderCreate", LuaKafkaReaderCreate)
	l.Register("WriteLog", LuaWriteLog)
	l.Register("KafkaWriteMessage", LuaKafkaWriteMessage)
	//-----------}kafka
	l.Register("httpPost", LuaHttpPost)
	l.Register("httpPost2", LuaHttpPost2)
	l.Register("httpGet", LuaHttpGet)
	l.Register("HttpGet2", LuaHttpGet2)
	l.Register("HttpsGet", LuaHttpsGet)
	l.Register("HttpsGet2", LuaHttpsGet2)
	l.Register("beBe", LuaBebe)

	l.Register("BeginTransaction", luaContext.LuaBeginTransaction)
	l.Register("CommitTransaction", luaContext.LuaCommitTransaction)
	l.Register("RollBackTransaction", luaContext.LuaRollbackTransaction)

	l.Register("CachedSqlQueryRow", luaContext.LuaCachedSqlQueryRow)
	l.Register("CachedSqlQueryRow2", luaContext.LuaCachedSqlQueryRow2)
	l.Register("CachedSqlQueryRows", luaContext.LuaCachedSqlQueryRows)

	l.Register("SqlQueryRow", luaContext.LuaSqlQueryRow)
	l.Register("SqlQueryRow2", luaContext.LuaSqlQueryRow2)
	l.Register("SqlQueryRows", luaContext.LuaSqlQueryRows)
	l.Register("SqlExec", luaContext.LuaSqlExec)
	l.Register("SqlExec2", luaContext.LuaSqlExec2)
	l.Register("SqlExec3", luaContext.LuaSqlExec3)
	l.Register("SqlInsert", luaContext.LuaSqlInsert)
	l.Register("LoadScript", luaContext.LuaLoadScript)
	l.Register("SendEmail", luaContext.LuaSendEmail)

	l.Register("XmlPathParse", LuaXmlPathParse)
	l.Register("GetPasswordHash", LuaGetPasswordHash)
	l.Register("JSONPathParse", LuaJSONPathParse)
	l.Register("GetUserParamValue", luaContext.LuaGetUserParamValue)
	l.Register("GetParamValue", LuaGetParamValue)
	l.Register("HTMLEscapeString", LuaHTMLEscapeString)
	l.Register("PathEscape", LuaPathEscape)
	l.Register("TelegramNewMessage", LuaTelegramNewMessage)
	l.Register("TelegramNewDocumentShare", LuaTelegramNewDocumentShare)
	l.Register("ReadFile", LuaReadFile)
	l.Register("WriteFile", LuaWriteFile)
	l.Register("AppendFile", LuaAppendFile)
	l.Register("StrReplace", LuaStrReplace)
	l.Register("HasPrefix", LuaHasPrefix)

	l.Register("JsonToString", LuaJsonToString)
	l.Register("StringToJson", LuaStringToJson)
	l.Register("QueryEscape", LuaQueryEscape)
	l.Register("Split", LuaSplit)
	l.Register("Command", LuaCommand)
	l.Register("RegexpCheck", LuaRegexpCheck)
	l.Register("Query", luaContext.LuaQuery)
	l.Register("QueryWithCount", luaContext.LuaQueryWithCount)
	l.Register("Detail", luaContext.LuaDetail)
	l.Register("ParseTemplate", LuaParseTemplate)
	l.Register("ParseHTMLTemplate", LuaParseHTMLTemplate)
	l.Register("UploadRawData", luaContext.LuaUploadRawData)

	l.Register("TempFile", LuaTempFile)
	l.Register("SendWSAsync", LuaSendWSAsync)

	l.Register("EncodeQRCode", luaContext.LuaEncodeQRCode)
	l.Register("StrTrimSpace", LuaStrTrimSpace)
	l.Register("RuNum2Word", LuaRuNum2Word)
	l.Register("LdapCreateUser", LuaLdapCreateUser)
	l.Register("ForceResetPassword", LuaForceResetPassword)
	l.Register("ToLower", LuaToLower)
	l.Register("TempDir", LuaTempDir)
	l.Register("UUID", LuaUUID)
	l.Register("GetHTTPListenHostPort", LuaGetHTTPListenHostPort)
	l.Register("Base64Decode", LuaBase64Decode)
	l.Register("Base64Encode", LuaBase64Encode)
	l.Register("MkdirAll", LuaMkdirAll)
	l.Register("JsonToXML", LuaJsonToXML)
	l.Register("Join", LuaJoin)
	l.Register("AccPostByPkOper", luaContext.LuaAccPostByPkOper)
	l.Register("AccPostByPkNoOper", luaContext.LuaAccPostByPkNoOper)
	l.Register("AccUndoByPkNoOper", luaContext.LuaAccUndoByPkNoOper)
	l.Register("RegexpFindAllStringsAndJoin", LuaRegexpFindAllStringsAndJoin)
	l.Register("EntityValueById", luaContext.LuaEntityValueById)
	l.Register("EntityValueByCode", luaContext.LuaEntityValueByCode)
	l.Register("DoScript", luaContext.LuaDoScript)
	l.Register("DoScriptGetBool", luaContext.LuaDoScriptGetBool)
	l.Register("DoScriptGetTable", luaContext.LuaDoScriptGetTable)
	l.Register("StripTags", LuaStripTags)
	l.Register("SendMail2", LuaSendMail2)
	l.Register("SendMail2NoTLS", LuaSendMail2NoTLS)
	l.Register("XmltoJSONString", LuaXmltoJSONString)
	l.Register("DBCurrentDateTime", LuaDBCurrentDateTime)
	l.Register("ImapProcessInBoxToMessage", LuaImapProcessInBoxToMessage)
	l.Register("SqlQueryRowsExtDb", luaContext.LuaSQLQueryRowsExtDb)
	l.Register("ParseEmailAddress", LuaParseEmailAddress)
	l.Register("CheckGrantOfEntity", luaContext.LuaCheckGrantOfEntity)
	l.Register("RegexpFindStringSubmatch", LuaRegexpFindStringSubmatch)
	l.Register("CryptoSignPKCS1v15", LuaCryptoSignPKCS1v15)
	l.Register("ColumnsByQuery", luaContext.LuaColumnsByQuery)
	l.Register("ClearCache", LuaClearCache)
	l.Register("Version", LuaVersion)
	l.Register("VersionNum", LuaVersionNum)
	l.Register("BPMNPublish", luaContext.LuaBPMNPublish)
	l.Register("MqSend", LuaMqSend)
	l.Register("MqReceive", LuaMqReceive)
	l.Register("TimeParseFormat", LuaTimeParseFormat)
	l.Register("TimeParseUnix", LuaTimeParseUnix)

	l.Register("Hostname", LuaHostname)
	l.Register("LoadString", LuaLoadString)
	l.Register("LoadStringNamed", LuaLoadStringNamed)
	l.Register("Sleep", LuaSleep)
	l.Register("SqlCall", luaContext.LuaSqlCall)
	l.Register("CSVRead", luaContext.CSVRead)
	l.Register("AccClsPublish", luaContext.AccClsPublish)
	l.Register("FileContent", luaContext.LuaFileContent)
	l.Register("FileContentNew", luaContext.LuaFileContentNew)
	l.Register("XlsxRead", luaContext.XlsxRead)
	l.Register("CopyFile", luaContext.LuaCopyFile)

	goluago.Open(l)
	stateorm.Open(l, o)
	pkg.Open(l)
	return nil
}
