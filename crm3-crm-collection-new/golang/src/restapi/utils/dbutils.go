package utils

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func DBCreateSEQTriggerByTableName(o orm.Ormer, tableName string) error {
	sql := `CREATE TRIGGER before_insert_uuid_` + tableName + `
	BEFORE INSERT ON ` + tableName + `
	FOR EACH ROW
	  IF new.sys$uuid IS NULL
	  THEN
		SET new.sys$uuid = uuid_v4();
	  END IF;`
	if GetDbDriverType() != orm.DROracle {
		_, err := o.Raw(sql).Exec()
		if err != nil {
			log.Println("==============DBCreateSEQTriggerByTableName ERROR 1==============" + sql)
			return err
		}

	} else {

		sequence_exists := 0
		err := o.Raw(DbBindReplace(`select count(1) as sequence_exists from user_sequences where sequence_name = 'SEQ_'||upper(?)`), tableName).QueryRow(&sequence_exists)

		if err != nil {
			log.Println("==============DBCreateSEQTriggerByTableName ERROR 2==============" + sql)
			return err
		}

		if sequence_exists == 0 {
			sql = "create sequence SEQ_" + tableName
			_, err = o.Raw(sql).Exec()
			if err != nil {
				log.Println("==============DBCreateSEQTriggerByTableName ERROR 3==============" + sql)
				return err
			}
		}

		sql = `CREATE OR REPLACE TRIGGER TRG_` + tableName + `_SEQ
		BEFORE INSERT ON ` + tableName + `
		FOR EACH ROW
		DECLARE
		  BEGIN		
			IF :new.sys$uuid IS NULL
			THEN
				:new.sys$uuid := uuid_v4();
			END IF;
			IF :new.id IS NULL
			THEN
				:new.id := SEQ_` + tableName + `.nextval();
			END IF;			
		  end TRG_` + tableName + `_SEQ;	
		  `
		_, err = o.Raw(sql).Exec()
		if err != nil {
			log.Println("==============DBCreateSEQTriggerByTableName ERROR 4==============" + sql)
			return err
		}

		sql = `select 'alter table '||?||' add constraint PK_'||?||' primary key (ID)'  
		res from dual e 
		where 
	not exists(
	  SELECT 1
	  FROM user_indexes i
	  WHERE uniqueness = 'UNIQUE'
	  and i.INDEX_NAME=upper('PK_'||?)
	  and i.TABLE_name = upper(?) ) `
		sqlPk := ""
		err = o.Raw(DbBindReplace(sql), tableName, tableName, tableName, tableName).QueryRow(&sqlPk)
		log.Println("==============DBCreateSEQTriggerByTableName DEBUG 5==============", err)
		if err == nil {
			_, err = o.Raw(sqlPk).Exec()
			if err != nil {
				log.Println("==============DBCreateSEQTriggerByTableName ERROR 5==============" + sql)
				return err

			}
		}

	}
	return nil
}

func DbBindReplace(sql string) string {

	if os.Getenv("CRM_DB_TYPE") == "oracle" {
		i := 0
		for {
			if !strings.Contains(sql, "?") {
				break
			}
			i++
			sql = strings.Replace(sql, "?", ":"+strconv.Itoa(i), 1)
		}
		sql = strings.Replace(sql, " limit 1", " and rownum = 1", -1)

	}

	return sql
}

func GetDbDriverType() orm.DriverType {
	if os.Getenv("CRM_DB_TYPE") == "pgsql" {
		return orm.DRPostgres
	}
	if os.Getenv("CRM_DB_TYPE") == "mysql" {
		return orm.DRMySQL
	}

	if os.Getenv("CRM_DB_TYPE") == "oracle" {
		return orm.DROracle
	}

	return orm.DRMySQL
}

func GetDbStringDelimiter() string {
	if GetDbDriverType() == orm.DRPostgres {
		return "\""
	} else if GetDbDriverType() == orm.DRMySQL {
		return "`"
	} else if GetDbDriverType() == orm.DROracle {
		return ""
	}
	return "`"
}

func DbInsert(o orm.Ormer, query string, args ...interface{}) (int64, error) {
	id := int64(0)
	if GetDbDriverType() == orm.DRPostgres {
		err := o.Raw(query+" returning id", args).QueryRow(&id)
		return id, err
	} else if GetDbDriverType() == orm.DRMySQL {
		rs, err := o.Raw(query, args).Exec()
		if err != nil {
			return 0, err
		} else {
			id, err = rs.LastInsertId()
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	} else if GetDbDriverType() == orm.DROracle {

		pr, err := o.Raw(DbBindReplace(
			"begin " + query + " returning id into :result; end;")).Prepare()

		defer pr.Close()
		if err != nil {
			log.Println("Error on DbInsert 1 ", err)
			return 0, err
		}
		args = append(args, sql.Named("result", sql.Out{Dest: &id}))
		_, err = pr.Exec(args...)

		if err != nil {
			log.Println("Error on DbInsert 2 ", err)
			return 0, err
		} else {
			return id, nil
		}
	}
	return 0, nil
}
