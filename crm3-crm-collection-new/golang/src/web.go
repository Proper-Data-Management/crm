package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"

	"os"
	"runtime"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
	// "gopkg.in/goracle.v2/goracle"
)

//import _ "github.com/mattn/go-oci8"

//import _ "github.com/ziutek/mymysql/mysql"

func init() {

	dt := orm.DRMySQL
	dname := "mysql"
	connStr := os.Getenv("OPENSHIFT_MYSQL_DB_USERNAME") + ":" + os.Getenv("OPENSHIFT_MYSQL_DB_PASSWORD") + "@tcp(" + os.Getenv("OPENSHIFT_MYSQL_DB_HOST") + ":" + os.Getenv("OPENSHIFT_MYSQL_DB_PORT") + ")/" + os.Getenv("OPENSHIFT_APP_NAME") + "?charset=utf8"
	if os.Getenv("CRM_DB_TYPE") == "pgsql" {
		dt = orm.DRPostgres
		dname = "postgres"
		connStr = os.Getenv("CRM_DB_CONN_STR")
	} else if os.Getenv("CRM_DB_TYPE") == "oracle" {
		dt = orm.DROracle
		dname = "oracle"
		connStr = os.Getenv("CRM_DB_CONN_STR")
	}
	err := orm.RegisterDriver(dname, dt)
	if err != nil {
		panic(err)
	}

	//log.Println(connStr)

	//
	//		gorm.Open("mysql",connStr)
	//		//t := ""
	//		new(gorm.DB).Raw("select 1 s").Rows()

	err = orm.RegisterDataBase("default", dname, connStr)
	if err != nil {
		panic(err)
	} else {
		log.Printf("I-COR-00001 version %s (%d) port %s", utils.Version, utils.VersionNum, os.Getenv("OPENSHIFT_GO_PORT"))
	}

}

func main() {

	//os.Setenv("GOGC","100")
	runtime.GC()
	//log.Println(http.ListenAndServe("localhost:6060", nil))

	err := luautils.AutoLoads()
	if err != nil {
		log.Println("error", err)
	}

	restapi.HandleInit()

}
