package utils

import (
	"errors"
	"fmt"
	"log"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

func accClsGetCodeById(accClsId int64) (string, error) {
	o := orm.NewOrm()
	o.Using("default")
	result := ""
	err := o.Raw(DbBindReplace("select code from acc_cls where id=?"), accClsId).QueryRow(&result)
	return result, err
}

func accClsCreateFKIndexes(accClsId int64) error {

	accClsCode, err := accClsGetCodeById(accClsId)
	if err != nil {
		err = errors.New(fmt.Sprintf("ERROR E-COR-00001. Error on accclsutils.accClsCreateFKIndexes accClsId=%v error = %v", accClsId, err.Error()))
		log.Println(err.Error())
		return err
	}
	accClsTableName := "acc$" + accClsCode

	sql :=
		`select concat (
'ALTER TABLE ',
?,
' ADD CONSTRAINT ',
concat('acc$',e.code,'_',ea.code, '_', e2.code, '_fk'),' FOREIGN KEY (', ea.code,') REFERENCES ',e2.code,' (id) ' )
res from acc_cls_attrs ea,data_types dt, acc_cls e,entities e2
where e.id=ea.acc_cls_id
and ea.data_type_id=dt.id and dt.code='reference'
and ea.entity_link_id = e2.id
and e.code=?
and not exists(
SELECT *
FROM information_schema.TABLE_CONSTRAINTS
WHERE constraint_type = 'FOREIGN KEY' and table_schema=database()
and constraint_name=concat('acc$',e.code,'_',ea.code,'_', e2.code, '_fk') COLLATE utf8_unicode_ci
)`

	if GetDbDriverType() == orm.DROracle {
		sql = `select 
		'ALTER TABLE '||
		?||
		' ADD CONSTRAINT '||
		'FK_ACC$'||e.code||'_'||ea.code||'_' || e2.code||' FOREIGN KEY ('|| ea.code||') REFERENCES '||e2.code||' (id) ' 
		as "res" from acc_cls_attrs ea,data_types dt, acc_cls e,entities e2
		where e.id=ea.acc_cls_id
		and ea.data_type_id=dt.id and dt.code='reference'
		and ea.entity_link_id = e2.id
		and e.code=?
		and not exists(
		SELECT *
		FROM user_constraints
		WHERE 
		 constraint_name=upper('FK_acc$'||e.code||'_'||ea.code||'_'||e2.code) and table_name = upper(e.code)
		 )`
	}

	o := orm.NewOrm()
	o.Using("default")
	log.Println("accClsCreateFKIndexes sql=" + sql)
	type addFieldsRows struct {
		Res string `json:"res"`
	}
	var ws = []addFieldsRows{}
	//log.Println(sql)
	_, err = o.Raw(DbBindReplace(sql), accClsTableName, accClsCode).QueryRows(&ws)

	if err != nil {
		err = errors.New(fmt.Sprintf("ERROR E-COR-00002. Error on accclsutils.accClsCreateFKIndexes accClsId=%v error = %v", accClsId, err.Error()))
		log.Println(err.Error())
		return err
	}
	for _, element := range ws {
		if element.Res == "" {
			log.Println("CONTINUE")
			continue
		}
		sql := element.Res
		log.Println("#################" + sql)
		_, err := o.Raw(DbBindReplace(sql)).Exec()
		if err != nil {
			err = errors.New(fmt.Sprintf("ERROR E-COR-00003. Error on accclsutils.accClsCreateFKIndexes accClsId=%v error = %v", accClsId, err.Error()))
			log.Println(err.Error())
			return err
		}
	}
	return err

}

func accClsAlterAddFields(accClsId int64) error {

	accClsCode, err := accClsGetCodeById(accClsId)
	if err != nil {

		return nil
	}
	accClsTableName := "acc$" + accClsCode

	sql := `
select
      CASE
      WHEN (dt.code = "reference")
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon)
      WHEN (coalesce(ea.len,0) > 0)
        THEN
          CONCAT(ea.code, ' ', dt.db_data_type, '(', ea.len, ')')
      WHEN (coalesce(ea.len,0) = 0)
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon) END res
from acc_cls_attrs ea,acc_cls e,data_types dt where e.id=ea.acc_cls_id
                                                    and dt.id=ea.data_type_id
                                                    and e.id=?
and not exists
(select 1 from information_schema.columns i where i.table_schema=database()
  and i.table_name=concat('acc$',e.code) COLLATE utf8_unicode_ci and i.column_name=ea.code  COLLATE utf8_unicode_ci
)
	`

	if GetDbDriverType() == orm.DROracle {

		sql = `
		select
		CASE
		WHEN (dt.code = 'reference')
		  THEN
			ea.code ||' ' || dt.ora_db_data_type || ' ' || dt.ora_addon
		WHEN (coalesce(ea.len,0) > 0)
		  THEN
			ea.code || ' ' || dt.ora_db_data_type || '(' || ea.len || ' char)'
		WHEN (coalesce(ea.len,0) = 0)
		  THEN
			ea.code || ' ' || dt.db_data_type || ' ' || dt.ora_addon END res
  from acc_cls_attrs ea,acc_cls e,data_types dt where e.id=ea.acc_cls_id
													  and dt.id=ea.data_type_id
													  and e.id=?
  and not exists
  (select 1 from user_tab_cols i where 
	 i.table_name=upper('acc$'||e.code) and i.column_name=upper(ea.code)
	 )`

	}

	o := orm.NewOrm()
	o.Using("default")
	type addFieldsRows struct {
		Res string `json:"res"`
	}
	var ws = []addFieldsRows{}
	//log.Println(sql)
	_, err = o.Raw(DbBindReplace(sql), accClsId).QueryRows(&ws)

	if err != nil {
		log.Println("accClsAlterAddFields " + err.Error())
		return err
	}
	for _, element := range ws {
		if element.Res == "" {
			//log.Println("CONTINUE")
			continue
		}
		sql := "alter table " + accClsTableName + " add " + element.Res
		//log.Println("@@@@@@@@@@"+sql)
		_, err := o.Raw(DbBindReplace(sql)).Exec()
		if err != nil {
			log.Println("error accClsAlterAddFields", err.Error())
			return err
		}
	}
	return err

}

func accClsAddUnique(accClsId int64) error {

	o := orm.NewOrm()
	o.Using("default")

	accClsCode, err := accClsGetCodeById(accClsId)
	if err != nil {
		return nil
	}
	accClsTableName := "acc$" + accClsCode

	fields := ""
	err = o.Raw(DbBindReplace("select group_concat(code) from acc_cls_attrs where acc_cls_id=?"), accClsId).QueryRow(&fields)
	if err != nil {
		return nil
	}
	sql := "CREATE UNIQUE INDEX " + accClsTableName + "_uq_index ON " + accClsTableName + " (" + fields + ")"
	_, err = o.Raw(DbBindReplace(sql)).Exec()

	return err

}
func accClsNewFields(accClsId int64) (string, error) {

	o := orm.NewOrm()
	o.Using("default")

	sql := `
select
  GROUP_CONCAT(
      CASE
      WHEN (dt.code = "reference")
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon)
      WHEN (ea.len > 0)
        THEN
          CONCAT(ea.code, ' ', dt.db_data_type, '(', ea.len, ')')
      WHEN (ea.len = 0)
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon) END
  )
from acc_cls_attrs ea,acc_cls e,data_types dt where e.id=ea.acc_cls_id
                                                    and dt.id=ea.data_type_id
                                                    and e.id=?
`

	if GetDbDriverType() == orm.DROracle {

		sql = `select
		LISTAGG(
		  CASE
		  WHEN (dt.code = 'reference')
			THEN
			ea.code || ' ' || dt.ora_db_data_type || ' ' || dt.ora_addon
		  WHEN (ea.len > 0)
			THEN
			ea.code || ' ' || dt.ora_db_data_type || '(' || ea.len || ')'
		  WHEN (ea.len = 0)
			THEN
			ea.code ||  ' ' || dt.ora_db_data_type ||  ' ' || dt.ora_addon END
		,',')
		from acc_cls_attrs ea,acc_cls e,data_types dt where e.id=ea.acc_cls_id
								  and dt.id=ea.data_type_id
								  and e.id=?`

	}

	res := ""
	err := o.Raw(DbBindReplace(sql), accClsId).QueryRow(&res)

	return res, err

}

func AccClsCreateTable(accClsId int64) error {

	o := orm.NewOrm()
	o.Using("default")

	accClsCode, err := accClsGetCodeById(accClsId)
	if err != nil {
		return nil
	}
	accClsTableName := "acc$" + accClsCode
	accClsHstTableName := "ach$" + accClsCode
	accClsMvTableName := "acm$" + accClsCode

	if tableExists(accClsTableName) {
		return nil
	}

	newFld, err := accClsNewFields(accClsId)
	if err != nil {
		err = errors.New(fmt.Sprintf("ERROR E-COR-00004. Error on accclsutils.AccClsCreateTable.accClsNewFields accClsId=%v error = %v", accClsId, err.Error()))
		log.Println(err.Error())
		return err
	}

	sql := "create table " + accClsTableName + " (`id` int(10) unsigned NOT NULL AUTO_INCREMENT,sys$uuid varchar(36),value double NOT NULL DEFAULT 0, " + newFld + ",PRIMARY KEY (`id`))"

	if GetDbDriverType() == orm.DROracle {
		sql = "create table " + accClsTableName + " (id integer,sys$uuid varchar2(36),value number (18,2) DEFAULT 0  NOT NULL, " + newFld + ")"

	}
	_, err = o.Raw(DbBindReplace(sql)).Exec()
	if err != nil {
		log.Println("==============SQL ERROR 0==============" + sql)
		return err
	}

	sql = "create table " + accClsHstTableName + " (`id` int(10) unsigned NOT NULL AUTO_INCREMENT,sys$uuid varchar(36),move_id int(10) unsigned NOT NULL,acc_id int(10) unsigned NOT NULL,bal_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP , created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP , value double NOT NULL DEFAULT 0,PRIMARY KEY (`id`))"
	if GetDbDriverType() == orm.DROracle {
		sql = "create table " + accClsHstTableName + ` (id integer  NOT NULL,sys$uuid varchar2(36),move_id integer NOT NULL,acc_id integer NOT NULL,bal_at date DEFAULT SYSDATE  NOT NULL, created_at date DEFAULT sysdate  NOT NULL, 
			value number  DEFAULT 0 NOT NULL)`
	}

	_, err = o.Raw(DbBindReplace(sql)).Exec()
	if err != nil {
		log.Println("==============SQL ERROR 1 ==============" + sql)
		return err
	}

	sql = "create table " + accClsMvTableName + " (`id` int(10) unsigned NOT NULL AUTO_INCREMENT,sys$uuid varchar(36),move_id int(10) unsigned NOT NULL,acc_id int(10) unsigned NOT NULL,  move_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP , created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP , value double NOT NULL DEFAULT 0,PRIMARY KEY (`id`))"
	if GetDbDriverType() == orm.DROracle {
		sql = "create table " + accClsMvTableName + ` (id integer NOT NULL,sys$uuid varchar2(36),move_id integer NOT NULL,acc_id integer NOT NULL,  move_at date  DEFAULT SYSDATE  NOT NULL, created_at date DEFAULT sysdate  NOT NULL, 
			value number  DEFAULT 0 NOT NULL)`
	}

	_, err = o.Raw(DbBindReplace(sql)).Exec()
	if err != nil {
		log.Println("==============SQL ERROR 3 ==============" + sql)
		return err
	}

	sql = "CREATE INDEX " + accClsMvTableName + "_move_id_index ON " + accClsMvTableName + " (move_id)"
	if GetDbDriverType() == orm.DROracle {
		sql = "CREATE INDEX IE_" + accClsMvTableName + "_move_id ON " + accClsMvTableName + " (move_id)"
	}
	_, err = o.Raw(DbBindReplace(sql)).Exec()
	if err != nil {
		log.Println("==============SQL ERROR 4 ==============" + sql)
		return err
	}

	err = DBCreateSEQTriggerByTableName(o, accClsTableName)

	if err != nil {
		log.Println("==============SQL ERROR 5 ==============" + sql)
		return err
	}

	err = DBCreateSEQTriggerByTableName(o, accClsHstTableName)

	if err != nil {
		log.Println("==============SQL ERROR 6 ==============" + sql)
		return err
	}
	err = DBCreateSEQTriggerByTableName(o, accClsMvTableName)

	if err != nil {
		log.Println("==============SQL ERROR 7 ==============" + sql)
		return err
	}

	err = accClsAddUnique(accClsId)

	if err != nil {
		log.Println("==============SQL ERROR 8 ==============" + sql)
		return err
	}
	return nil
}

func AccClsPublish(id int64) error {

	log.Println("id = ")
	log.Println(id)
	err := AccClsCreateTable(id)
	if err != nil {
		return err
	}
	log.Println("AccClsPublish AccClsCreateTable ok")

	err = accClsAlterAddFields(id)
	if err != nil {
		return err
	}
	log.Println("AccClsPublish accClsAlterAddFields ok")

	accClsCreateFKIndexes(id)
	if err != nil {
		return err
	}
	log.Println("AccClsPublish accClsCreateFKIndexes ok")

	return nil
}
