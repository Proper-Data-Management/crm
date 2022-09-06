package restapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/julienschmidt/httprouter"
)

func ImportAdvancedReferenceRestApi(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	type referenceImportEntitiesRequest struct {
		Entities []orm.Params `json:"entities"`
	}
	type referenceImportResponse struct {
		UpdateCount int    `json:"updateCount"`
		InsertCount int    `json:"insertCount"`
		DeleteCount int    `json:"deleteCount"`
		SkipCount   int    `json:"skipCount"`
		ErrorCount  int    `json:"errorCount"`
		ErrorTexts  string `json:"errorTexts"`
	}

	contents, err := ioutil.ReadAll(req.Body)
	log.Println(string(contents))

	defer req.Body.Close()

	var t referenceImportEntitiesRequest
	err = json.Unmarshal(contents, &t)

	if err != nil {
		RestCheckPanic(err, res)
		return
	}

	var entityPriors map[string]int
	entityPriors = make(map[string]int)

	entityPriors["accounts"] = 1
	entityPriors["bi_deals"] = 2
	entityPriors["bi_addresses"] = 1
	entityPriors["bi_nomens"] = 9
	entityPriors["bi_mobilities"] = 9
	entityPriors["bi_constructions"] = 9
	entityPriors["contacts"] = 2
	entityPriors["bi_ind_sites"] = 9
	entityPriors["bi_vehicles"] = 5
	entityPriors["bi_vehicle_vids"] = 4
	entityPriors["bi_drivers"] = 4
	entityPriors["bi_gosnum"] = 6
	entityPriors["bi_individuals"] = 9
	entityPriors["bi_beton_invoices"] = 10

	var resP referenceImportResponse

	o := orm.NewOrm()
	o.Using("default")
	for i := 1; i <= 10; i++ {
		for _, element := range t.Entities {
			entity := element["entity"].(string)
			if entityPriors[entity] == i {
				log.Println("@@@@@@@@@@@@ process " + entity)
				if utils.CheckTableRegexpBool(entity) {
					sqlCnt := "select count(1) cnt from " + entity + " where code=?"
					cnt := 0
					err := o.Raw(utils.DbBindReplace(sqlCnt), element["code"]).QueryRow(&cnt)
					if cnt > 0 {
						_, err = AdvancedImportCaseUpdate(entity, o, element)
						if err == nil {
							resP.UpdateCount++
						}
					} else {
						_, err = AdvancedImportCaseInsert(entity, o, element)
						if err == nil {
							resP.InsertCount++
						}
					}
					if err != nil {
						resP.ErrorCount++
						resP.ErrorTexts += err.Error() + " in " + entity + "\n"
					}
				} else {
					resP.ErrorCount++
					resP.ErrorTexts += "Invalid tablename \"" + entity + "\"\n"
				}
				//break;
			}
		}

	}

	resP.DeleteCount = 0
	resP.SkipCount = 0
	j, _ := json.Marshal(resP)
	fmt.Fprint(res, string(j))
	log.Println(string(j))

}
