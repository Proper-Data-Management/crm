package restapi

import "net/http"

import (
	"fmt"
	"log"
	"net/http/pprof"
	"os"
	"runtime"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ws"
	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func handlePprof(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	switch p.ByName("pprof") {
	case "/cmdline":
		pprof.Cmdline(w, r)
	case "/profile":
		pprof.Profile(w, r)
	case "/symbol":
		pprof.Symbol(w, r)
	case "/mem":

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		fmt.Fprint(w, "Alloc ")
		fmt.Fprintln(w, mem.Alloc)

		fmt.Fprint(w, "TotalAlloc ")
		fmt.Fprintln(w, mem.TotalAlloc)

		fmt.Fprint(w, "HeapAlloc ")
		fmt.Fprintln(w, mem.HeapAlloc)

		fmt.Fprint(w, "HeapSys ")
		fmt.Fprintln(w, mem.HeapSys)

	default:
		pprof.Index(w, r)
	}

}

func HandleInit() {

	router := httprouter.New()

	router.GET("/", RedirectHome)

	router.GET("/debug/pprof/*pprof", handlePprof)

	router.ServeFiles("/static/*filepath", http.Dir("static"))

	router.GET("/restapi/test", TestRestApiGet)
	router.GET("/gc", GC)

	router.GET("/restapi/query/get", QueryRestApiGet)
	router.GET("/restapi/list/simple/get", ListRestApiGetSimpleList)
	router.POST("/restapi/update", UpdateRestApi)
	router.POST("/restapi/update_v_1_1", UpdateRestApi_v_1_1)
	router.GET("/restapi/detail", DetailRestApi)

	//Deprecated
	//router.GET("/restapi/removeall", RemoveAll)
	router.GET("/page/*filepath", DynamicPage)

	router.POST("/restapi/login", Login)
	router.GET("/restapi/loginByToken/:token", LoginByToken)

	router.POST("/restapi/extlogin", LoginExt)
	router.GET("/restapi/extloginbytoken/:token", LoginByTokenExt)
	router.GET("/restapi/extlogout", utils.LogoutExt)

	router.POST("/auth/login", utils.Login)

	router.GET("/auth/session_info", utils.GetSessionInfo)
	router.POST("/auth/resetpassword", utils.ForceResetPasswordReq)
	router.POST("/auth/resetmypassword", utils.ResetMyPassword)

	//deprecated
	router.GET("/auth/logout", utils.Logout)
	//new
	router.GET("/restapi/logout", utils.Logout)
	router.POST("/restapi/logout", utils.Logout)

	router.GET("/auth/getlanguage", utils.GetLanguage)
	router.POST("/auth/setLanguage", utils.SetLanguage)

	//router.POST("/cdr", Cdr)--DEPRECATED

	router.GET("/restapi/translates/get", TranslateRestApiGet)

	router.GET("/restapi/pages/get", PageRestApiGet)
	router.GET("/restapi/pagetemplate", PageRestApiGetPageTemplate)
	router.GET("/restapi/widgettemplate", WidgetRestApiGetWidgetTemplate)
	router.GET("/restapi/widget/:code", WidgetRestApiGetWidgetCode)

	//deprecated
	//router.POST("/restapi/generateddl", GenerateDDL)
	//router.GET("/restapi/generateallddldb", GenerateDDLAllDB)
	router.GET("/restapi/menus/tree", MenuRestApiGetTree)
	router.GET("/restapi/menus/tree_v2", MenuRestApiGetTree_v2)
	router.POST("/restapi/upload", Upload)
	router.GET("/restapi/userpic", UserPic)
	router.GET("/restapi/getfile", GetFile)

	//DEPRECATED
	//router.GET("/exportall", ExportAll)

	router.POST("/restapi/importStandartReference/", ImportStandartReferenceRestApi)
	router.POST("/restapi/importAdvancedReference/", ImportAdvancedReferenceRestApi)

	//deprecated
	//router.GET("/restapi/ExportEntity/:entity", ExportEntity)
	//router.GET("/restapi/ExportEntityByCode/:entity", ExportEntityByCode)
	//router.GET("/restapi/ExportAllEntity/:dir", ExportAllEntity)

	router.OPTIONS("/restapi/bpms/start", BPMStartProcessOptions)
	router.POST("/restapi/bpms/start", BPMStartProcess)

	//router.POST("/restapi/BPMCreateInstance/", BPMCreateInstance)
	//router.OPTIONS("/restapi/BPMCreateInstance/", BPMCreateInstanceOptions)
	router.POST("/restapi/bpms/manualExecByInstance/", BPMManualExecByInstance)
	router.GET("/restapi/bpms/ShowUserTaskForm/:task", BPMShowUserTaskForm)

	router.GET("/restapi/bpms/ShowInstanceForm/:instance", BPMShowInstanceForm)

	router.POST("/restapi/bpms/publish", BPMPublish)
	router.POST("/restapi/bpms/publish/", BPMPublish)

	router.POST("/restapi/bpms/runprocess/", BPMRunProcess_Deprecated)
	router.POST("/restapi/bpms/runUserTask", BPMRunUserTask)
	router.OPTIONS("/restapi/bpms/runUserTaskByInstance/", BPMRunUserTaskByInstanceOptions)
	router.POST("/restapi/bpms/runUserTaskByInstance/", BPMRunUserTaskByInstance)

	router.OPTIONS("/restapi/bpms/runUserTaskByInstance", BPMRunUserTaskByInstanceOptions)
	router.POST("/restapi/bpms/runUserTaskByInstance", BPMRunUserTaskByInstance)

	router.POST("/restapi/bpms/runUserTaskByTask/", BPMRunUserTaskByTask)
	//router.POST("/restapi/bpms/tableGenerate", BPMTableGenerate)
	//router.GET("/restapi/bpms/tableGenerateAll", BPMTableGenerateAll)

	//deprecated
	router.GET("/ws/ServeAccountChat", ws.ServeAccountChat)
	router.POST("/ws/ServeAccountChat", ws.ServeAccountChat)

	router.GET("/ws/ServeMsgToUser", ws.ServeMsgToUser)
	router.POST("/ws/ServeMsgToUser", ws.ServeMsgToUser)

	router.GET("/ws/ServeMsgToRoom/:room", ws.ServeMsgToRoom)
	router.POST("/ws/ServeMsgToRoom/:room", ws.ServeMsgToRoom)

	//end deprecated

	router.GET("/restapi/ws/ServeAccountChat", ws.ServeAccountChat)
	router.POST("/restapi/ws/ServeAccountChat", ws.ServeAccountChat)

	router.GET("/restapi/ws/ServeMsgToUser", ws.ServeMsgToUser)
	router.POST("/restapi/ws/ServeMsgToUser", ws.ServeMsgToUser)

	router.GET("/restapi/ws/ServeMsgToRoom/:room", ws.ServeMsgToRoom)
	router.POST("/restapi/ws/ServeMsgToRoom/:room", ws.ServeMsgToRoom)

	router.POST("/restapi/apns", ApnsRestApi)
	router.GET("/restapi/telegram/resetwebhooks", TelegramResetWebHooks)
	router.POST("/restapi/telegram/webhook", TelegramWebHook)

	router.POST("/restapi/services/run/:code", ServiceRun)
	router.OPTIONS("/restapi/services/run/:code", ServiceRunOptions)
	router.GET("/restapi/services/run/:code", ServiceRunGet)

	router.GET("/restapi/email/unsubscribe/:uuid", EmailUnSubscribe)
	router.GET("/restapi/ldap/sync/:ldap", utils.LdapSync)

	router.POST("/restapi/acccls/pub/", AccClsPublish)

	router.GET("/restapi/ntlmexample", NTLMExample)

	router.GET("/restapi/soap/:soap", SOAP)
	router.POST("/restapi/soap/:soap", SOAPDo)

	//router.POST("/restapi/acc/move/",AccMove)
	//router.GET("/restapi/acc/undo/:id",AccUndo)
	//router.GET("/restapi/acc/postbypk/",AccPostByPk)

	//flag.Parse()

	/*
		connectionCount := 2000
		l, err := net.Listen("tcp", ":8000")
		if err != nil {
			log.Fatalf("Listen: %v", err)
		}
		defer l.Close()
		l = netutil.LimitListener(l, connectionCount)
	*/

	bind := fmt.Sprintf("%s:%s", os.Getenv("OPENSHIFT_GO_IP"), os.Getenv("OPENSHIFT_GO_PORT"))
	log.Fatal(http.ListenAndServe(bind, router))

	//log.Fatal(http.ListenAndServe(bind,nil))

}

func GC(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	runtime.GC()
	fmt.Fprint(res, "GC OK")
}
func RedirectHome(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	o := orm.NewOrm()
	o.Using("default")

	if utils.UserId(req) != 0 {

		if IsMobile(req) {
			http.Redirect(res, req, "/static/mobile.html#"+utils.GetDomainParamValue(req.Host, "homepage"), 301)
		} else {
			http.Redirect(res, req, "/static/#"+utils.GetDomainParamValue(req.Host, "homepage"), 301)
		}
	} else {

		http.Redirect(res, req, utils.GetDomainParamValue(req.Host, "loginpage"), 301)
	}
}
