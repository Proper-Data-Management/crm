package utils

import (
	"encoding/xml"
	"log"
	"os"
	"runtime/debug"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
)

type TypeSequenceFlow struct {
	Id                  string   `xml:"id,attr"`
	Name                string   `xml:"name,attr"`
	SourceRef           string   `xml:"sourceRef,attr"`
	TargetRef           string   `xml:"targetRef,attr"`
	ConditionExpression []string `xml:"conditionExpression"`
	isCondition         int
}
type TypeStartEvent struct {
	Id       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Outgoing []string
}

type TypeStopEvent struct {
	Id       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Incoming []string
}

type TypeIntermediateCatchEvent struct {
	Id                   string `xml:"id,attr"`
	Name                 string `xml:"name,attr"`
	Incoming             []string
	TimerEventDefinition []string `xml:"timerEventDefinition"`
}
type TypeUserTask struct {
	Id       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Outgoing []string `xml:"incoming"`
	Incoming []string `xml:"outgoing"`
}

type TypeServiceTask struct {
	Id                          string   `xml:"id,attr"`
	Name                        string   `xml:"name,attr"`
	Outgoing                    []string `xml:"incoming"`
	Incoming                    []string `xml:"outgoing"`
	StandardLoopCharacteristics []string `xml:"standardLoopCharacteristics"`
}

type TypeSubProcess struct {
	Id                          string   `xml:"id,attr"`
	Name                        string   `xml:"name,attr"`
	Outgoing                    []string `xml:"incoming"`
	Incoming                    []string `xml:"outgoing"`
	StandardLoopCharacteristics []string `xml:"standardLoopCharacteristics"`
}

type TypeScriptTask struct {
	Id                          string   `xml:"id,attr"`
	Name                        string   `xml:"name,attr"`
	Outgoing                    []string `xml:"incoming"`
	Incoming                    []string `xml:"outgoing"`
	StandardLoopCharacteristics []string `xml:"standardLoopCharacteristics"`
}

type TypeManualTask struct {
	Id       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Outgoing []string `xml:"incoming"`
	Incoming []string `xml:"outgoing"`
}

type TypeParallelGateway struct {
	Id       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Outgoing []string `xml:"incoming"`
	Incoming []string `xml:"outgoing"`
}

type TypeInclusiveGateway struct {
	Id       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Outgoing []string `xml:"incoming"`
	Incoming []string `xml:"outgoing"`
}

type TypeExclusiveGateway struct {
	Id       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Outgoing []string `xml:"incoming"`
	Incoming []string `xml:"outgoing"`
}

type TypeProcess struct {
	Id                     string                       `xml:"id,attr"`
	UserTask               []TypeUserTask               `xml:"userTask"`
	ScriptTask             []TypeScriptTask             `xml:"scriptTask"`
	ManualTask             []TypeManualTask             `xml:"manualTask"`
	ServiceTask            []TypeServiceTask            `xml:"serviceTask"`
	SubProcess             []TypeSubProcess             `xml:"subProcess"`
	StartEvent             []TypeStartEvent             `xml:"startEvent"`
	EndEvent               []TypeStopEvent              `xml:"endEvent"`
	IntermediateCatchEvent []TypeIntermediateCatchEvent `xml:"intermediateCatchEvent"`
	SequenceFlow           []TypeSequenceFlow           `xml:"sequenceFlow"`
	ExclusiveGateway       []TypeExclusiveGateway       `xml:"exclusiveGateway"`
	ParallelGateway        []TypeParallelGateway        `xml:"parallelGateway"`
	InclusiveGateway       []TypeInclusiveGateway       `xml:"inclusiveGateway"`
}

type TypeBPMN2 struct {
	XMLName xml.Name    `xml:"definitions"`
	Process TypeProcess `xml:"process"`
}

func (context *BpmGenContext) genUserTaskFormPointsByProcessId(process_id int64, typeText string) error {

	if typeText != "usertask" && typeText != "endevent" {
		return nil
	}
	type tPointStruct struct {
		Id    int64
		Title string
	}
	var arr []tPointStruct
	context.O.Raw(DbBindReplace("select p.id,p.title from bp_points p,bp_point_types pt where pt.id=p.type_id and pt.code=? and p.process_id=? and p.is_auto=1"), typeText, process_id).QueryRows(&arr)
	for _, v := range arr {
		err := context.genUserTaskFormPoint(v.Id, v.Title, typeText)
		if err != nil {
			return err
		}
	}

	return nil
}
func (context *BpmGenContext) genUserTaskFormPoint(point_id int64, titleText, typeText string) error {

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("genUserTaskFormPoint...")
	}
	condButtons := ""

	sql := `select
  replace(
  group_concat(
	  concat('<button class="btn btn-primary" translate ng-click="runUserTaskWithGetParams({\'condition_',sf.id,'\':\'1\'})">',coalesce(nullif(sf.title, ''), 
	  
	  (select nullif(p2.title,'') from 
		bp_point_sfs ps2 
		join bp_points p2 on p2.id=ps2.point_id
		where ps2.is_incoming=1
		and ps2.sf_id=ps.sf_id
		and ps2.is_active=1
		and p2.is_active=1
		limit 1
		),
	 
	 
	  'NoTitle'),'</button>')
  ),",","\r\n")
from bp_sequence_flows sf,bp_point_sfs ps WHERE
  sf.id=ps.sf_id and ps.point_id=? and ps.is_incoming=0
  and sf.is_condition=1 and sf.is_active=1 and ps.is_active=1`

	//log.Println(sql)
	err := context.O.Raw(DbBindReplace(sql), point_id).QueryRow(&condButtons)

	//log.Println("condButtons = ")
	//log.Println(condButtons)
	//log.Println(err)
	//log.Println(point_id)

	if err != nil {
		log.Println("error1 on genUserTaskFormPoint " + err.Error())
		return nil
	}
	template := ""
	err = context.O.Raw(DbBindReplace(`select
  replace(
      group_concat(

      	concat('<div ng-if="errorVars.',pv.code,'" class="form-group"><label class="text-danger">', pv.title, ' {{errorVars.', pv.code, ' | translate}}</label></div>\r\n',
          CASE
          WHEN v.is_output = '1'
            THEN
              CASE WHEN dt.code = 'reference'
                THEN
                  concat('<bu-select query-code="', (SELECT q.code
                                                     FROM entities e, queries q
                                                     WHERE q.id = e.def_sel_query_id AND e.id = pv.entity_link_id),
                         '" label="', pv.title, '" id-model="var.', pv.code, '" ></bu-select>\r\n')
              WHEN dt.code = 'double'
                THEN
                  concat('<bu-input-number-label label="', pv.title, '" id-model="var.', pv.code,
                         '" ></bu-input-number-label>\r\n')
              WHEN dt.code = 'integer'
                THEN
                  concat('<bu-input-number-label label="', pv.title, '" id-model="var.', pv.code,
'" ></bu-input-number-label>\r\n')
              WHEN dt.code = 'text'
                THEN
                  concat('<div> 
				  <div class="form-group" >
					  <label>', pv.title, '</label>
					  <div style=" height: 400px; "
						   ng-model="var.', pv.code,'"
						   ng-change="edit()"
						   ui-ace="{
					  useWrapMode : false,
					  showGutter: true,
					  theme:''twilight'',
					  mode: ''html'',
					  firstLineNumber: 1,
					  onLoad: aceLoaded,
					  onChange: aceChanged
					  }">test</div>
				  </div>
			  </div>\r\n')						 
              WHEN dt.code = 'varchar'
                THEN
                  concat('<bu-input-label label="', pv.title, '" id-model="var.', pv.code, '" ></bu-input-label>\r\n')
              WHEN dt.code = 'date'
                THEN
                  concat('<bu-date label="', pv.title, '" id-model="var.', pv.code, '" ></bu-date>\r\n')
              WHEN dt.code = 'timestamp'
                THEN
                  concat('<bu-date-time label="', pv.title, '" id-model="var.', pv.code, '" ></bu-date-time>\r\n')
              WHEN dt.code = 'boolean'
                THEN
                  concat('<bu-checkbox label="', pv.title, '" id-model="var.', pv.code, '" ></bu-checkbox>\r\n')
              END

          WHEN v.is_input = '1'
            THEN
              CASE WHEN dt.code = 'reference'
                THEN
                  concat('<bu-select readonly=true query-code="', (SELECT q.code
                                                     FROM entities e, queries q
                                                     WHERE q.id = e.def_sel_query_id AND e.id = pv.entity_link_id),
                         '" label="', pv.title, '" id-model="var.', pv.code, '" ></bu-select>\r\n')
              WHEN dt.code = 'double'
                THEN
                  concat('<bu-input-number-label readonly=true label="', pv.title, '" id-model="var.', pv.code,
                         '" ></bu-input-number-label>\r\n')
              WHEN dt.code = 'integer'
                THEN
                  concat('<bu-input-number-label readonly=true label="', pv.title, '" id-model="var.', pv.code,
                         '" ></bu-input-number-label>\r\n')
              WHEN dt.code = 'varchar'
                THEN
                  concat('<bu-input-label readonly=true  label="', pv.title, '" id-model="var.', pv.code, '" ></bu-input-label>\r\n')
              WHEN dt.code = 'date'
                THEN
                  concat('<bu-date readonly=true label="', pv.title, '" id-model="var.', pv.code, '" ></bu-date>\r\n')
              WHEN dt.code = 'timestamp'
                THEN
                  concat('<bu-date-time readonly=true label="', pv.title, '" id-model="var.', pv.code, '" ></bu-date-time>\r\n')
              WHEN dt.code = 'boolean'
                THEN
                  concat('<bu-checkbox readonly=true label="', pv.title, '" id-model="var.', pv.code, '" ></bu-checkbox>\r\n')
              END
          END)
      ),
      '\r\n,','\r\n')
from bp_point_vars v,bp_process_vars pv,data_types dt where v.point_id=?
                                                            and dt.id=pv.data_type_id
                                                            and pv.id=v.var_id
	`), point_id).QueryRow(&template)

	log.Println(template)

	cntCond := 0
	err = context.O.Raw(DbBindReplace(`select count(1) from bp_point_sfs ps 
	join bp_sequence_flows sf on sf.id=ps.sf_id
	where ps.point_id=? and ps.is_incoming=0
	and sf.is_active=1 and ps.is_active=1
	and sf.is_condition=1`), point_id).QueryRow(&cntCond)

	if err != nil {
		log.Println("error225 on genUserTaskFormPoint " + err.Error())
		return err
	}

	if cntCond == 0 {
		condButtons = "<button class=\"btn btn-primary\" translate ng-click=\"runUserTask()\">Next</button>\r\n" + condButtons
	}
	if typeText == "usertask" {
		template = "<h3>" + titleText + "</h3>\r\n" + template +
			condButtons + "\r\n"
	} else {
		template = "<h3>" + titleText + "</h3>\r\n" + template + "<button class=\"btn btn-primary\" translate ng-click=\"finish()\">Finish</button>"
	}
	if err != nil {
		log.Println("error2 on genUserTaskFormPoint " + err.Error())
		return err
	}

	_, err = context.O.Raw(DbBindReplace("update bp_points set form=? where id=? and is_auto=1"), template, point_id).Exec()

	if err != nil {
		log.Println("error3 on genUserTaskFormPoint " + err.Error())
		return err
	}

	return nil
}

func (context *BpmGenContext) importPoint(titleText string, typeText string, elementId string, processId int64, loop bool, timerEvent bool) error {

	cnt := 0

	iLoop := 0
	if loop {
		iLoop = 1
	}
	iTimerEvent := 0
	if timerEvent {
		iTimerEvent = 1
	}
	err := context.O.Raw(DbBindReplace("select count(1) cnt from bp_points where code=?"), elementId).QueryRow(&cnt)
	if err != nil {
		panic(err)
	}
	if cnt == 0 {

		if typeText == "usertask" {
			_, err := context.O.Raw(
				DbBindReplace(`insert into bp_points (is_active,is_loop,is_timerevent,code,title,type_id,process_id,actor_id,is_auto,actor_type_id)
				values (1,?,?,?,?,(select id from bp_point_types where code=?),?,(select id from bp_actors where code='initiator'),1, (select id from bp_actor_types where code='actor') )`),
				iLoop, iTimerEvent, elementId, titleText, typeText, processId).Exec()
			if err != nil {
				return err
			}
		} else {

			_, err := context.O.Raw(
				DbBindReplace(`insert into bp_points (is_active,is_loop,is_timerevent,code,title,type_id,process_id,is_auto)
				values (1,?,?,?,?,(select id from bp_point_types where code=?),?,1)`),
				iLoop, iTimerEvent, elementId, titleText, typeText, processId).Exec()
			if err != nil {
				//log.Println("test66666",err);
				return err
			}

		}

	} else {
		_, err = context.O.Raw(
			DbBindReplace(`update bp_points  set is_active=1, is_loop=?,is_timerevent=?, title=?, type_id=(select id from bp_point_types where code=?),process_id=?
				where code=?`),
			iLoop, iTimerEvent, titleText, typeText, processId, elementId).Exec()
	}

	return err
}

func (context *BpmGenContext) Publish(processId int64) error {

	cached.ClearCache()

	err := context.BpmTableGenerate(processId)
	if err != nil {
		log.Println("Error on BpmTableGenerate ", err)
		debug.PrintStack()
		return err
	}
	diagram := ""
	err = context.O.Raw(DbBindReplace("select diagram from bp_processes where id=?"), processId).QueryRow(&diagram)
	if err != nil {
		log.Println("Error on select diagram ", err)
		debug.PrintStack()
		return err
	}
	err = context.ImportBPMN2(diagram, processId)
	if err != nil {
		log.Println("Error on ImportBPMN2 ", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func (context *BpmGenContext) ImportBPMN2(xmlStr string, processId int64) error {

	v := &TypeBPMN2{}
	err := xml.Unmarshal([]byte(xmlStr), &v)
	if err != nil {
		return err
	}

	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println(xmlStr)
	}

	_, err = context.O.Raw(DbBindReplace("update bp_points set is_active=0 where process_id=?"), processId).Exec()
	if err != nil {
		log.Println("Error on ImportBPMN2 1")
		return err
	}

	_, err = context.O.Raw(DbBindReplace("update bp_sequence_flows set is_active=0 where process_id=?"), processId).Exec()
	if err != nil {
		log.Println("Error on ImportBPMN2 2")
		return err
	}

	_, err = context.O.Raw(DbBindReplace("update bp_point_sfs set is_active=0 where point_id in (select id from bp_points where process_id=?)"), processId).Exec()
	if err != nil {
		log.Println("Error on ImportBPMN2 3")
		return err
	}
	cnt := 0
	for _, element := range v.Process.StartEvent {
		err = context.importPoint(element.Name, "startevent", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 4")
			return err
		}
	}
	for _, element := range v.Process.EndEvent {
		err = context.importPoint(element.Name, "endevent", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 5")
			return err
		}
	}
	for _, element := range v.Process.IntermediateCatchEvent {

		err = context.importPoint(element.Name, "intermediatecatchevent", element.Id, processId, false, len(element.TimerEventDefinition) > 0)
		if err != nil {
			log.Println("Error on ImportBPMN2 6")
			return err
		}
	}
	for _, element := range v.Process.UserTask {
		err = context.importPoint(element.Name, "usertask", element.Id, processId, false, false)

	}
	for _, element := range v.Process.ServiceTask {
		err = context.importPoint(element.Name, "servicetask", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 7")
			return err
		}
		if len(element.StandardLoopCharacteristics) > 0 {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("LOOP" + element.Name)
			}
			err = context.importPoint(element.Name, "servicetask", element.Id, processId, true, false)
			if err != nil {
				log.Println("Error on ImportBPMN2 8")
				return err
			}
		} else {
			err = context.importPoint(element.Name, "servicetask", element.Id, processId, false, false)
			if err != nil {
				log.Println("Error on ImportBPMN2 9")
				return err
			}
		}
	}
	for _, element := range v.Process.SubProcess {
		err = context.importPoint(element.Name, "subprocess", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 10")
			return err
		}
		if len(element.StandardLoopCharacteristics) > 0 {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("LOOP" + element.Name)
			}
			err = context.importPoint(element.Name, "subprocess", element.Id, processId, true, false)
			if err != nil {
				log.Println("Error on ImportBPMN2 11")
				return err
			}
		} else {
			err = context.importPoint(element.Name, "subprocess", element.Id, processId, false, false)
			if err != nil {
				log.Println("Error on ImportBPMN2 12")
				return err
			}
		}
	}
	for _, element := range v.Process.ScriptTask {
		err = context.importPoint(element.Name, "scripttask", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 13")
			return err
		}
		if len(element.StandardLoopCharacteristics) > 0 {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("LOOP" + element.Name)
			}
			err = context.importPoint(element.Name, "scripttask", element.Id, processId, true, false)
			if err != nil {
				log.Println("Error on ImportBPMN2 14")
				return err
			}
		} else {
			err = context.importPoint(element.Name, "scripttask", element.Id, processId, false, false)
			if err != nil {
				log.Println("Error on ImportBPMN2 15")
				return err
			}
		}
	}
	for _, element := range v.Process.ManualTask {
		err = context.importPoint(element.Name, "manualtask", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 16")
			return err
		}
	}
	for _, element := range v.Process.ExclusiveGateway {
		err = context.importPoint(element.Name, "exclusivegateway", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 17")
			return err
		}
	}

	for _, element := range v.Process.InclusiveGateway {
		err = context.importPoint(element.Name, "inclusivegateway", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 18")
			return err
		}
	}

	for _, element := range v.Process.ParallelGateway {
		err = context.importPoint(element.Name, "parallelgateway", element.Id, processId, false, false)
		if err != nil {
			log.Println("Error on ImportBPMN2 19")
			return err
		}
	}

	for _, element := range v.Process.SequenceFlow {
		if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
			log.Println("SourceRef=" + element.SourceRef)
		}

		context.O.Raw(DbBindReplace("select count(1) cnt from bp_sequence_flows where code=?"), element.Id).QueryRow(&cnt)

		if len(element.ConditionExpression) > 0 {
			element.isCondition = 1
		} else {
			element.isCondition = 0
		}

		if cnt == 0 {
			lid, err := DbInsert(context.O, DbBindReplace("insert into bp_sequence_flows (is_active,code,title,process_id,is_condition,is_auto) values (1,?,?,?,?,1)"), element.Id, element.Name, processId, element.isCondition)
			if err != nil {
				log.Println("Error on ImportBPMN2 20")
				return err

			}

			if element.isCondition == 1 {
				if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
					log.Println("creating ...")
				}

				sql := "update bp_sequence_flows set cond = concat('request.get.condition_',?,' == \"1\"') where id=? "

				if GetDbDriverType() == orm.DROracle {
					sql = "update bp_sequence_flows set cond = 'request.get.condition_'||?||' == \"1\"' where id=? "
				}
				_, err = context.O.Raw(DbBindReplace(sql), lid, lid).Exec()
				if err != nil {
					log.Println("err " + err.Error())
					return err
				}
			}

		} else {
			_, err = context.O.Raw(DbBindReplace("update bp_sequence_flows set is_active=1, title=?,process_id=?,is_condition=? where code=?"), element.Name, processId, element.isCondition, element.Id).Exec()
			if err != nil {
				return err
			}

			if element.isCondition == 1 {

				sql := "update bp_sequence_flows set cond = concat('request.get.condition_',id,' == \"1\"') ,is_active=1, title=?,process_id=?,is_condition=? where code=? and is_auto=1"

				if GetDbDriverType() == orm.DROracle {
					sql = "update bp_sequence_flows set cond = 'request.get.condition_'||id||' == \"1\"' ,is_active=1, title=?,process_id=?,is_condition=? where code=? and is_auto=1"
				}

				_, err = context.O.Raw(DbBindReplace(sql), element.Name, processId, element.isCondition, element.Id).Exec()
				if err != nil {
					return err
				}
			}
		}
		cnt = 0
		if element.SourceRef != "" {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("element.SourceRef=" + element.SourceRef)
			}
			context.O.Raw(DbBindReplace("select count(1) cnt from bp_point_sfs where is_incoming=0 and sf_id=(select id from bp_sequence_flows where code=?) "), element.Id).QueryRow(&cnt)
			if cnt == 0 {
				context.O.Raw(DbBindReplace(`insert into bp_point_sfs
			(is_incoming,sf_id,point_id,is_active)
			 values
			 (0,(select id from bp_sequence_flows where code=?),(select id from bp_points where code=?) ,1 ) `),
					element.Id, element.SourceRef).Exec()
			} else {
				context.O.Raw(DbBindReplace(`update bp_point_sfs
			set is_active=1,point_id=(select id from bp_points where code=?)
			 where sf_id=(select id from bp_sequence_flows where code=?) and is_incoming=0`),
					element.SourceRef, element.Id).Exec()
			}
		}
		if element.TargetRef != "" {
			if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
				log.Println("element.TargetRef=" + element.TargetRef)
			}
			context.O.Raw(DbBindReplace("select count(1) cnt from bp_point_sfs where is_incoming=1 and sf_id=(select id from bp_sequence_flows where code=?) "), element.Id).QueryRow(&cnt)
			if cnt == 0 {
				context.O.Raw(DbBindReplace(`insert into bp_point_sfs
			(is_incoming,sf_id,point_id,is_active)
			 values
			 (1,(select id from bp_sequence_flows where code=?),(select id from bp_points where code=?),1 ) `),
					element.Id, element.TargetRef).Exec()
			} else {
				context.O.Raw(DbBindReplace(`update bp_point_sfs
			set is_active=1,point_id=(select id from bp_points where code=?)
			 where sf_id=(select id from bp_sequence_flows where code=?) and is_incoming=1`),
					element.TargetRef, element.Id).Exec()
			}
		}
	}

	if GetDbDriverType() != orm.DROracle {

		err = context.genUserTaskFormPointsByProcessId(processId, "usertask")
		if err != nil {
			return err
		}
		err = context.genUserTaskFormPointsByProcessId(processId, "endevent")
		return err
	}
	return nil

}
