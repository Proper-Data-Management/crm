package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "git.dar.kz/crediton-3/crm-mfo/src/lib/go-oci8"
)

func getDSN() string {
	var dsn string
	if len(os.Args) > 1 {
		dsn = os.Args[1]
		if dsn != "" {
			return dsn
		}
	}
	dsn = os.Getenv("CRM_DB_CONN_STR")
	if dsn != "" {
		return dsn
	}
	fmt.Fprintln(os.Stderr, `Please specifiy connection parameter in GO_OCI8_CONNECT_STRING environment variable,
or as the first argument! (The format is user/name@host:port/sid)`)
	return "scott/tiger@XE"
}

func main() {
	os.Setenv("NLS_LANG", "")

	db, err := sql.Open("oci8", getDSN())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	db.Exec("drop table blob_example")

	_, err = db.Exec("create table blob_example(id varchar2(256) not null primary key, data blob)")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Over 4000bytes
	b := []byte(strings.Repeat("qwe12asd34", 7000))

	_, err = db.Exec("insert into blob_example(id, data) values(:1, :2)", "001", b)
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("select * from blob_example")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var data []byte
		rows.Scan(&id, &data)
		if string(b) != string(data) {
			panic("BLOB doesn't work correctly")
		}
		fmt.Println(id, string(data))
	}
}
