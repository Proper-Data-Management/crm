package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"

	_ "github.com/go-sql-driver/mysql"
)

func init() {

}
func NewDB() (*sql.DB, error) {
	if GetDbDriverType() == orm.DRMySQL {
		db, err := sql.Open("mysql", os.Getenv("OPENSHIFT_MYSQL_DB_USERNAME")+":"+os.Getenv("OPENSHIFT_MYSQL_DB_PASSWORD")+"@tcp("+os.Getenv("OPENSHIFT_MYSQL_DB_HOST")+":"+os.Getenv("OPENSHIFT_MYSQL_DB_PORT")+")/"+os.Getenv("OPENSHIFT_APP_NAME")+"?charset=utf8")

		if err != nil {
			return nil, err
		}
		if err = db.Ping(); err != nil {
			return nil, err
		}
		return db, nil
	} else if GetDbDriverType() == orm.DRPostgres {

		db, err := sql.Open("postgres", os.Getenv("CRM_DB_CONN_STR"))

		if err != nil {
			return nil, err
		}
		if err = db.Ping(); err != nil {
			return nil, err
		}
		return db, nil

	} else if GetDbDriverType() == orm.DROracle {

		db, err := sql.Open("oci8", os.Getenv("CRM_DB_CONN_STR"))

		if err != nil {
			return nil, err
		}
		if err = db.Ping(); err != nil {
			return nil, err
		}
		return db, nil

	}
	return nil, errors.New("Unknown Database Type")
}

//func SqlRows2JsonGorm(sqlString string) (string, error) {
//	db.Raw("SELECT name, age FROM users WHERE name = ?", 3).Scan(&result)
//}

func SqlRows2Json(sqlString string, args interface{}) (string, error) {
	db, err := NewDB()
	defer db.Close()

	//rows, err := new(gorm.DB).Raw(sqlString).Rows()
	rows, err := db.Query(sqlString, args)
	if err != nil {
		db.Close()
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		db.Close()
		return "", err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
			v = nil
			b = nil
		}
		tableData = append(tableData, entry)
		ClearInterface(&entry)
		entry = nil
	}
	rows.Close()
	rows = nil
	valuePtrs = valuePtrs[:0]
	valuePtrs = nil

	jsonData, err := json.Marshal(tableData)
	if err != nil {
		jsonData = nil
		err = nil
		tableData = nil
		db.Close()
		return "", err
	}
	//fmt.Println(string(jsonData))

	err = nil
	tableData = nil
	db.Close()

	return string(jsonData), nil
}

func SqlRows2Table(sqlString string, args ...interface{}) ([]map[string]interface{}, error) {
	db, err := NewDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(sqlString, args...)
	//rows, err := new(gorm.DB).Raw(sqlString).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			//log.Println("col1", col)
			col = strings.ToLower(col)
			//log.Println("col2", col)
			entry[col] = v
			v = nil
			b = nil
		}

		tableData = append(tableData, entry)
		ClearInterface(&entry)
		entry = nil
	}
	rows.Close()
	rows = nil
	valuePtrs = valuePtrs[:0]
	valuePtrs = nil
	if err != nil {
		return nil, err
	}

	values = nil

	//tableData = tableData[:0]

	//fmt.Println(string(jsonData))

	return tableData, nil
}

func SqlRows2TableDb(rows *sql.Rows, sqlString string, args ...interface{}) ([]map[string]interface{}, error) {

	//rows, err := new(gorm.DB).Raw(sqlString).Rows()

	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
			v = nil
			b = nil
		}
		tableData = append(tableData, entry)
		ClearInterface(&entry)
		entry = nil
	}
	rows.Close()
	rows = nil
	valuePtrs = valuePtrs[:0]
	valuePtrs = nil
	if err != nil {
		return nil, err
	}

	values = nil

	//tableData = tableData[:0]

	//fmt.Println(string(jsonData))

	return tableData, nil
}

func IsDuplicateRow(err error) bool {
	if err == nil {
		return false
	} else {
		return strings.Contains(err.Error(), "Error 1062")
	}
}

func IsTableNotExists(err error) bool {
	if err == nil {
		return false
	} else {
		return strings.Contains(err.Error(), "Error 1146")
	}
}

func IsNoRowFound(err error) bool {
	if err == nil {
		return false
	} else {
		return strings.Contains(err.Error(), "no row found")
	}
}
