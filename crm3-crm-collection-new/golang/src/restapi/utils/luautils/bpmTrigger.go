package luautils

import (
	"log"
	"strconv"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
)

func DoTriggerOnUpdateOne(o orm.Ormer, entityId int64, user_id int64, id int) {

	//func QueryByUrl(urlStr string,host string, user_id int64)([] orm.Params,error){

	var triggers []orm.Params
	var vars []orm.Params
	//o := orm.NewOrm()
	//o.Using("default")
	o.Raw(utils.DbBindReplace(`select su.url,sut.process_id,sut.id trigger_id from filter_save_users su, filter_save_user_triggers sut,bp_processes p
		where  sut.process_id=p.id and su.id=sut.parent_id and p.action_entity_id=?`), entityId).Values(&triggers)

	for _, v := range triggers {
		log.Println("DoTriggerOnUpdateOne...")
		log.Println(v)
		log.Println("DoTriggerOnUpdateOne id")
		log.Println(id)
		//url := "?code=accounts&flt$896$func1$=850210301899&ids=" + strconv.Itoa(id);
		url := v["url"].(string) + "&ids=" + strconv.Itoa(id)
		log.Println("DoTriggerOnUpdateOne url=" + url)
		_, p, err := utils.QueryByUrl(o, url, "", user_id, false, "ru")
		if err != nil {
			log.Println("error on DoTriggerOnUpdateOne " + err.Error())
			return
		}

		log.Println("DoTriggerOnUpdateOne len")
		log.Println(len(p))
		if len(p) > 0 {

			o.Raw(utils.DbBindReplace(`select v.code,trv.by_value_id,trv.value_id
			 from filter_su_tr_vars trv,bp_process_vars v where trv.parent_id=?
			and v.id=trv.process_var_id`), v["trigger_id"]).Values(&vars)
			var input = []NameValue{}
			input = append(input, NameValue{Name: "url", Value: url})
			for _, v1 := range vars {
				if v1["by_value_id"].(string) == "1" {
					input = append(input, NameValue{Name: v1["code"].(string), Value: v1["value_id"].(string)})
				}
			}

			log.Println("DoTriggerOnUpdateOne input =")
			log.Println(input)
			processId, err := strconv.Atoi(v["process_id"].(string))

			context := InstanceContext{}
			context.O = o
			context.InstanceVars = make(map[int64][]NameValue)
			context.InstanceTables = make(map[int64]string)

			if err != nil {
				log.Println("error strconv.Atoi " + err.Error())
			} else {
				context.CreateInstance(nil, int64(processId), 1, input, 0)
			}
		}
	}

}
