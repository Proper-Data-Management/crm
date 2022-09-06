package utils

import (
	"log"
	"os"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

var openshift_db = os.Getenv("OPENSHIFT_APP_NAME")

type BpmGenContext struct {
	Name string
	O    orm.Ormer
}

func (context *BpmGenContext) bpmNewFields(bpId int64) (string, error) {

	//entityCode := bpmGetEntityCode(bpId)
	sql := `
select
  GROUP_CONCAT(
      CASE
      WHEN (ea.is_array = 1)
        THEN
          concat(ea.code, ' ', 'longtext', ' ', '')

      WHEN (dt.code = "reference")
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon)
      WHEN (coalesce(ea.len,0) > 0)
        THEN
          CONCAT(ea.code, ' ', dt.db_data_type, '(', coalesce(ea.len,0), ')')
      WHEN (coalesce(ea.len,0) = 0)
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon) END
  )
from bp_process_vars ea,bp_processes e,data_types dt where e.id=ea.process_id
and dt.id=ea.data_type_id
and e.id=?
`

	if GetDbDriverType() == orm.DROracle {
		sql = `
	select
    LISTAGG(
        CASE
        WHEN (ea.is_array = 1)
          THEN
            ea.code || ' CLOB '
  
        WHEN (dt.code = 'reference')
          THEN
            ea.code ||  ' '|| dt.ora_db_data_type || ' ' ||dt.ora_addon
        WHEN (coalesce(ea.len,0) > 0)
          THEN
            ea.code || ' ' || dt.ora_db_data_type || '(' || coalesce(ea.len,0) || ' char)'
        WHEN (coalesce(ea.len,0) = 0)
          THEN
            ea.code || ' ' || dt.ora_db_data_type || ' ' || dt.ora_addon
            
             END
    ,',') WITHIN GROUP (ORDER BY ea.code)
  from bp_process_vars ea,bp_processes e,data_types dt where e.id=ea.process_id
  and dt.id=ea.data_type_id
  and e.id=?
	`

	}

	res := ""
	err := context.O.Raw(DbBindReplace(sql), bpId).QueryRow(&res)
	if err != nil {
		log.Println("Error on bpmNewFields", sql, err)
	}
	return res, err

}

func (context *BpmGenContext) bpmGetTableName(bpId int64) (string, error) {

	result := ""
	err := context.O.Raw(DbBindReplace("select code from bp_processes where id=?"), bpId).QueryRow(&result)
	return "i$" + result, err
}

func (context *BpmGenContext) bpmCreateTable(bpId int64) error {

	newFld, err := context.bpmNewFields(bpId)
	if err != nil {
		log.Println("Error on bpmNewFields", err)
		return err
	}
	//	if newFld == "" {
	//		log.Println("newFld is empty exitting")
	//		return nil
	//	}

	entityCode, err := context.bpmGetTableName(bpId)
	if err != nil {
		log.Println("Error on bpmGetTableName", err)
		return err
	}

	sql := "create table " + entityCode + " (sys$uuid varchar(36),created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  id$ int unsigned, " + newFld + ")"
	if GetDbDriverType() == orm.DROracle {

		sql = "create table " + entityCode + " (sys$uuid varchar2(36),created_at date DEFAULT sysdate, update_at date,  id$ integer, " + newFld + ")"

	}

	if newFld == "" {
		sql = "create table " + entityCode + " (sys$uuid varchar(36),created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  id$ int unsigned)"
		if GetDbDriverType() == orm.DROracle {

			sql = "create table " + entityCode + " (sys$uuid varchar2(36),created_at date DEFAULT sysdate, update_at date,  id$ integer)"

		}
	}
	if os.Getenv("CRM_DEBUG_SQL") == "1" {
		log.Println("==============" + sql)
	}
	_, err = context.O.Raw(DbBindReplace(sql)).Exec()
	if err != nil {
		log.Println("Error on CREATE table ", sql, err)
		return err
	}

	sql = `CREATE TRIGGER before_insert_uuid_` + entityCode + `
	BEFORE INSERT ON ` + entityCode + `
	FOR EACH ROW
		IF new.sys$uuid IS NULL
		THEN
			SET new.sys$uuid = uuid_v4();
		END IF;`

	if GetDbDriverType() == orm.DROracle {

		sql = `create or replace trigger trg_` + entityCode + `_seq
before insert on ` + entityCode + `  
for each row
declare
begin
if :new.sys$uuid is null then
	:new.sys$uuid := uuid_v4();
end if; 

end TRG_` + entityCode + `_SEQ;`
	}
	_, err = context.O.Raw(sql).Exec()

	if err != nil {
		log.Println("error on CREATE trigger ", sql, err)
		return err
	}

	sql = "CREATE UNIQUE INDEX " + entityCode + "_id$_index ON " + entityCode + " (id$)"

	if GetDbDriverType() == orm.DROracle {
		sql = "alter table " + entityCode + " add constraint PK$" + entityCode + " primary key (ID$)"

	}

	_, err = context.O.Raw(DbBindReplace(sql)).Exec()

	if err != nil {
		log.Println("error on CREATE INDEX ", sql, err)
		return err
	}

	/*
	sql = `ALTER TABLE ` + entityCode + `
	ADD CONSTRAINT ` + entityCode + `_bp_instances_id_fk
	FOREIGN KEY (id$) REFERENCES bp_instances (id) ON DELETE CASCADE ON UPDATE CASCADE`

	if GetDbDriverType() == orm.DROracle {

		sql = `alter table ` + entityCode + `
  add constraint FK$` + entityCode + ` foreign key (ID$)
	references bp_instances (ID) on delete cascade`
	}

	_, err = context.O.Raw(DbBindReplace(sql)).Exec()

	if err != nil {
		log.Println("Error on ALTER TABLE ", sql, err)
		return err
	}
	*/

	return nil
}

func (context *BpmGenContext) bpmTableExists(entityCode string) bool {

	i := 0

	sql := "SELECT 1 FROM information_schema.tables	WHERE table_schema = database() AND table_name = ? limit 1"
	if GetDbDriverType() == orm.DROracle {
		sql = "SELECT 1 FROM user_tables WHERE table_name = upper(?) and rownum=1"
	}
	err := context.O.Raw(DbBindReplace(sql), entityCode).QueryRow(&i)
	return err == nil
}

func (context *BpmGenContext) bpmDoCreateUQIndexes(bpId int64) error {

	entityCode, err := context.bpmGetTableName(bpId)
	if err != nil {
		log.Println("Error on bpmDoCreateUQIndexes ", err)
		return err
	}

	sql :=
		`select concat ( 'CREATE UNIQUE INDEX ', concat(e.code,'_',ea.code, '_uindex'),' ON ', e.code , ' (', ea.code,')' ) res
	 	from bp_process_vars ea,bp_processes e where e.id=ea.process_id and ea.uq='1'
	and e.code=?
and not exists(
SELECT *
FROM information_schema.TABLE_CONSTRAINTS
WHERE constraint_type = 'UNIQUE' and table_schema=?
and constraint_name=concat(e.code,'_',ea.code, '_uindex') COLLATE utf8_unicode_ci
)`

	//log.Println(sql)
	type addFieldsRows struct {
		Res string `json:"res"`
	}
	var ws = []addFieldsRows{}
	//log.Println(sql)
	_, err = context.O.Raw(DbBindReplace(sql), entityCode, openshift_db).QueryRows(&ws)

	if err != nil {
		log.Println("Error on ", sql, err)
		return err
	}
	for _, element := range ws {
		if element.Res == "" {
			log.Println("CONTINUE")
			continue
		}
		sql := element.Res
		log.Println("#################" + sql)
		_, err := context.O.Raw(DbBindReplace(sql)).Exec()
		if err != nil {
			return err
		}
	}
	return err

}

func (context *BpmGenContext) bpmDoAlterAddFields(bpId int64) error {

	sql := `
select
      CASE
      WHEN (dt.code = "reference")
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon)
      WHEN (coalesce(ea.len,0) > 0)
        THEN
          CONCAT(ea.code, ' ', dt.db_data_type, '(', coalesce(ea.len,0), ')')
      WHEN (coalesce(ea.len,0) = 0)
        THEN
          concat(ea.code, ' ', dt.db_data_type, ' ', dt.addon) END res
from bp_process_vars ea,bp_processes e,data_types dt where e.id=ea.process_id
                                                    and dt.id=ea.data_type_id
                                                    and e.id=?
and not exists
(select 1 from information_schema.columns i where i.table_schema=database()
  and i.table_name=? and i.column_name=ea.code  COLLATE utf8_unicode_ci
)
	`

	if GetDbDriverType() == orm.DROracle {

		sql = `
		select
		CASE
		WHEN (dt.code = 'reference')
			THEN
				ea.code || ' ' || dt.ora_db_data_type || dt.ora_addon
		WHEN (coalesce(ea.len,0) > 0)
			THEN
				ea.code || ' ' || dt.ora_db_data_type || '(' || coalesce(ea.len,0) ||  ')'
		WHEN (coalesce(ea.len,0) = 0)
			THEN
				ea.code || ' ' || dt.ora_db_data_type || ' ' || dt.ora_addon END as "res"
from bp_process_vars ea,bp_processes e,data_types dt where e.id=ea.process_id
																									and dt.id=ea.data_type_id
																									and e.id=?
and not exists
(select 1 from user_tab_cols i where 
i.table_name=upper(?) and i.column_name=upper(ea.code)
)
	`

	}
	if os.Getenv("CRM_DEBUG_BPMS") == "1" {
		log.Println("altering tables", sql)
	}
	entityCode, err := context.bpmGetTableName(bpId)
	if err != nil {
		log.Println("Error on  bpmDoAlterAddFields 1", err.Error)
		return err
	}

	type addFieldsRows struct {
		Res string `json:"res"`
	}
	var ws = []addFieldsRows{}
	//log.Println(sql)
	_, err = context.O.Raw(DbBindReplace(sql), bpId, entityCode).QueryRows(&ws)

	if err != nil {
		log.Println("Error on bpmDoAlterAddFields 2", err, sql)
		return err
	}
	for _, element := range ws {
		if element.Res == "" {
			log.Println("CONTINUE")
			continue
		}
		sql := "alter table " + entityCode + " add " + element.Res
		log.Println("@@@@@@@@@@" + sql)
		_, err := context.O.Raw(DbBindReplace(sql)).Exec()
		if err != nil {
			log.Println("Error on ", sql, err)
			return err
		}
	}
	return err

}

func (context *BpmGenContext) bpmDoCreateFKIndexes(bpId int64) error {

	entityCode, err := context.bpmGetTableName(bpId)
	if err != nil {
		log.Println("Error on bpmDoCreateFKIndexes ", err.Error())
		return err
	}

	sql :=
		`select concat (
'ALTER TABLE ',
e.code ,
' ADD CONSTRAINT ',
concat(e.code,'_',ea.code, '_', e2.code, '_fk'),' FOREIGN KEY (', ea.code,') REFERENCES ',e2.code,' (id) ',coalesce((select ruletext from entity_attr_update_rules ur where  ur.id=ea.rule_id),'') )
res from bp_process_vars ea,data_types dt, bp_processes e,entities e2
where e.id=ea.process_id
and ea.data_type_id=dt.id and dt.code='Reference'
and ea.entity_link_id = e2.id
and e.code=?
and not exists(
SELECT *
FROM information_schema.TABLE_CONSTRAINTS
WHERE constraint_type = 'FOREIGN KEY' and table_schema=?
and constraint_name=concat(e.code,'_',ea.code,'_', e2.code, '_fk') COLLATE utf8_unicode_ci
)`

	//log.Println(sql)
	type addFieldsRows struct {
		Res string `json:"res"`
	}
	var ws = []addFieldsRows{}
	//log.Println(sql)
	_, err = context.O.Raw(DbBindReplace(sql), entityCode, openshift_db).QueryRows(&ws)

	if err != nil {
		log.Println("Error on bpmDoCreateFKIndexes 1", sql)
		return err
	}
	for _, element := range ws {
		if element.Res == "" {
			log.Println("CONTINUE")
			continue
		}
		sql := element.Res
		log.Println("#################" + sql)
		_, err := context.O.Raw(DbBindReplace(sql)).Exec()
		if err != nil {
			log.Println("Error on bpmDoCreateFKIndexes 2", sql)
			return err
		}
	}
	return err

}

func (context *BpmGenContext) BpmTableGenerate(bpId int64) error {

	entityCode, err := context.bpmGetTableName(bpId)
	if err != nil {
		log.Println("BpmTableGenerate ", err.Error())
		return err
	}

	if !context.bpmTableExists(entityCode) {
		err := context.bpmCreateTable(bpId)
		if err != nil {
			log.Println("Error on bpmCreateTable")
			return err
		}
	} else { //Existing Table
		err := context.bpmDoAlterAddFields(bpId)
		if err != nil {
			log.Println("Error on bpmDoAlterAddFields")
			return err
		}
	}

	//TODO Oracle FK not supported temporaru
	if GetDbDriverType() != orm.DROracle {
		err = context.bpmDoCreateUQIndexes(bpId)
		if err != nil {
			log.Println("error on bpmDoCreateUQIndexes")
			return err
		}

		err = context.bpmDoCreateFKIndexes(bpId)
		if err != nil {
			log.Println("error on bpmDoCreateFKIndexes")
			return err
		}
	}
	return nil
}
