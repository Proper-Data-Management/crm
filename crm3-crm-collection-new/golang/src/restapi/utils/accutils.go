package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

type TAccMoves struct {
	Entity int64      `json:"enitity`
	Pk     int64      `json:"pk`
	Moves  []TAccMove `json:"moves`
}
type TAccMove struct {
	IsMinus      bool                   `json:"is_minus"`
	AccCode      string                 `json:"code"`
	Attrs        map[string]interface{} `json:"attrs"`
	IgnoreErrors int64                  `json:"ignore_errors"`
}

//Проведение операции
//Функция запускается из accpostutils,
//Не допустимо использовать transaction. Транзакция покрывается в accpostutils
//op - оперативное проведение
func AccMove(o orm.Ormer, allData TAccMoves, op bool, date string) (int64, error) {

	//log.Println(allData)

	uniQAccCls := make(map[string]string)
	move_id, err := DbInsert(o, DbBindReplace("insert into acc_moves (entity_id,entity_pk) values (?,?)"), allData.Entity, allData.Pk)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	for _, data := range allData.Moves {

		uniQAccCls[data.AccCode] = "1"
		var values []interface{}
		var columns []string
		var columnsSelect []string
		var columnsQ []string

		for k, v := range data.Attrs {

			if os.Getenv("CRM_DEBUG_ACC") == "1" {
				log.Println("CRM_DEBUG_ACC", "AccMove iter", move_id, v)
			}

			if k != "amount$" {
				//log.Println(data.Value)
				columns = append(columns, k)
				columnsSelect = append(columnsSelect, k+"=?")
				columnsQ = append(columnsQ, "?")
				values = append(values, v)
				if v == nil {
					if data.IgnoreErrors == 1 {
						log.Println(fmt.Sprintf("Skipped Error: Value of atribute %v is empty. Entity_id=%v, Entity_pk=%v", k, allData.Entity, allData.Pk))
						//continue
					} else {
						return 0, errors.New(fmt.Sprintf("Error: Value of atribute %v is empty. Entity_id=%v, Entity_pk=%v", k, allData.Entity, allData.Pk))
					}
				}
			}
		}

		sqlId := "select id,value from acc$" + data.AccCode + " where " + strings.Join(columnsSelect, " and ") + " for update"
		if os.Getenv("CRM_DEBUG_ACC") == "1" {
			log.Println("AccMove", move_id, sqlId)
		}

		amount, err := strconv.ParseFloat(fmt.Sprintf("%v", data.Attrs["amount$"]), 64)
		if IsTableNotExists(err) {
			log.Println(err)
			return 0, errors.New("AccCls `" + data.AccCode + "` not found")
		}
		if data.IsMinus {
			amount = amount * -1
		}

		if os.Getenv("CRM_DEBUG_ACC") == "1" {
			log.Println("~~~~~~~~~~~~~~~AccMove amount", move_id, amount)
		}

		id := int64(0)
		lastValue := 0.00
		err = o.Raw(DbBindReplace(sqlId), values).QueryRow(&id, &lastValue)

		if os.Getenv("CRM_DEBUG_ACC") == "1" {
			log.Println("~~~~~~~~~~~~~~~AccMove lastvalue", move_id, lastValue)
		}

		//log.Println("ZZZ", id)

		if IsTableNotExists(err) {
			log.Println(err)
			return 0, errors.New("AccCls `" + data.AccCode + "` not found")
		} else if id == 0 {
			sqlInsert := "insert into acc$" + data.AccCode + " (" + strings.Join(columns, ",") + ",value) values " + "(?," + strings.Join(columnsQ, ",") + ")"
			log.Println("ACcMove sqlInsert", sqlInsert)
			values = append(values, amount)
			id, err = DbInsert(o, DbBindReplace(sqlInsert), values...)
			if err != nil {
				log.Println(err)
				log.Println("values")
				log.Println(values)
				return 0, err
			}

		} else if err != nil {
			if data.IgnoreErrors == 1 {

				log.Println("accmove skip error2", move_id, sqlId, err)
				continue
			} else {
				log.Println("accmove error2", move_id, sqlId, err)
				return 0, err
			}

		} else {
			sqlUpdate := "update acc$" + data.AccCode + " set value=value+? where id=?"

			if os.Getenv("CRM_DEBUG_ACC") == "1" {
				log.Println("~~~~~~~~~~~~~~~AccMove increment", move_id, amount, id)
			}

			//log.Println(sqlUpdate)
			_, err := o.Raw(DbBindReplace(sqlUpdate), amount, id).Exec()
			if err != nil {
				log.Println(err)
				return 0, err
			}
		}

		//if 1 == 0 {
		if op {
			_, err = o.Raw(DbBindReplace("insert into acm$"+data.AccCode+" (move_id,acc_id,value) values (?,?,?)"), move_id, id, amount).Exec()
			if err != nil {
				log.Println(err)
				return 0, err
			}
		} else {

			_, err = o.Raw(DbBindReplace("insert into acm$"+data.AccCode+" (move_id,acc_id,value,move_at) values (?,?,?,?)"), move_id, id, amount, date).Exec()
			if err != nil {
				log.Println(err)
				return 0, err
			}

		}
		//}

	}

	for accKey, _ := range uniQAccCls {

		if 1 == 0 {

			if op {
				_, err = o.Raw(DbBindReplace("insert into ach$"+accKey+" (move_id,acc_id,value) (select ?,id,value from acc$"+accKey+" where id in (select acc_id from acm$"+accKey+" where move_id=?))"), move_id, move_id).Exec()
				if err != nil {
					log.Println(err)
					return 0, err
				}
			} else {

				type Tx struct {
					AccId int
					Value float64
					Bal   float64
				}

				t := []Tx{}
				_, err = o.Raw(DbBindReplace(`select m.acc_id as "acc_id",m.value as "value", coalesce( (select value from  ach$`+accKey+

					` h where acc_id=m.acc_id and bal_at<? order by h.bal_at desc limit 1 ), 0) as "bal" from acm$`+accKey+

					" m left join acc$"+accKey+" a on a.id = m.acc_id where m.move_id=? order by m.move_at "), date,

					move_id).QueryRows(&t)

				//Удалить историю остатков с даты начала транзации
				_, err = o.Raw(DbBindReplace("delete from ach$"+accKey+" where bal_at >= ? and acc_id in (select acc_id from acm$"+accKey+" where move_id=?)"), date, move_id).Exec()
				if err != nil {
					log.Println(err)
					return 0, err
				}

				for _, val := range t {
					log.Println("valEl", val)

					_, err = o.Raw(DbBindReplace("insert into ach$"+accKey+" (move_id,value,acc_id,bal_at)values (?,?,?,?)"), move_id, val.Bal+val.Value, val.AccId, date).Exec()
					if err != nil {
						log.Println(err)
						return 0, err
					}
				}

				//log.Println("ost",)
				/*_, err = o.Raw("insert into ach$" + accKey + " (move_id,acc_id,value,bal_at) (select ?,id,value,? from acc$" + accKey + " where id in (select acc_id from acm$" + accKey + " where move_id=?))", move_id, date, move_id).Exec()
				if err != nil {
					log.Println(err)
					return 0, err
				}*/

				//_, err = o.Raw("update ach$" + accKey + " h join acm$" + accKey + " m on m.acc_id = h.acc_id set h.value=m.value where m.move_id=? and h.bal_at>? and m.move_at>?",  move_id,date,date).Exec()
				if err != nil {
					log.Println(err)
					return 0, err
				}

				//Пересчитать остатки по движениям с даты начала транзации

			}
		}

		_, err := o.Raw(DbBindReplace("insert into acc_move_acc_cls (move_id,acc_cls_id) values (?,(select id from acc_cls where code=?))"), move_id, accKey).Exec()
		if err != nil {
			log.Println(err)
			return 0, err
		}

	}

	//o.Commit()
	return move_id, nil

}

//Функция запускается из accpostutils,
//Не допустимо использовать transaction. Транзакция покрывается в accpostutils
func AccUndoByEntityIdPk(o orm.Ormer, entity_id, pk int64) error {

	entityCode := ""
	err := cached.O().Raw(DbBindReplace("select code from entities where id=?"), entity_id).QueryRow(&entityCode)
	if err != nil {
		return nil
	}
	posted := 0
	err = o.Raw(DbBindReplace("select  posted from "+entityCode+" where id=?"), pk).QueryRow(&posted)
	if err != nil {
		return nil
	}
	if posted == 0 {
		return nil
	}

	_, err = o.Raw(DbBindReplace("update "+entityCode+" set posted=0,move_id=null where id=?"), pk).Exec()
	if err != nil {
		return nil
	}

	moveId := int64(0)
	err = o.Raw(DbBindReplace("select id from acc_moves where entity_id=? and entity_pk=? limit 1"), entity_id, pk).QueryRow(&moveId)
	if err != nil && IsNoRowFound(err) {
		return nil
	} else if err != nil {
		return nil
	} else {
		AccUndo(o, moveId)
	}
	return nil
}

//Функция запускается из accpostutils,
//Не допустимо использовать transaction. Транзакция покрывается в accpostutils
func AccUndo(o orm.Ormer, moveId int64) error {

	err := o.Raw(DbBindReplace("select id from acc_moves where id=? for update"), moveId).QueryRow(&moveId)
	if err != nil {
		log.Println("AccUndo " + err.Error())

		return err
	}
	var accClsArr []string
	_, err = o.Raw(DbBindReplace("select ac.code from acc_move_acc_cls ma,acc_cls ac where ma.acc_cls_id=ac.id and ma.move_id=?"), moveId).QueryRows(&accClsArr)
	if err != nil {
		log.Println("AccUndo " + err.Error())

		return err
	}
	type tUndoAccMove struct {
		AccId int64   `json:"acc_id"`
		Value float64 `json:"value"`
	}
	var undoAccMove []tUndoAccMove

	for _, clsItem := range accClsArr {
		_, err = o.Raw(DbBindReplace(`select acc_id as "acc_id",value as "value" from acm$`+clsItem+" where move_id=?"), moveId).QueryRows(&undoAccMove)
		if err != nil {
			log.Println("AccUndo " + err.Error())

			return err
		}
		for _, v := range undoAccMove {

			_, err = o.Raw(DbBindReplace("delete from ach$"+clsItem+" where move_id=?"), moveId).Exec()

			if err != nil {
				log.Println("AccUndo " + err.Error())

				return err
			}

			_, err = o.Raw(DbBindReplace("delete from acm$"+clsItem+" where move_id=?"), moveId).Exec()

			if err != nil {
				log.Println("AccUndo " + err.Error())

				return err
			}

			_, err = o.Raw(DbBindReplace("update acc$"+clsItem+" set value=value-? where id=?"), v.Value, v.AccId).Exec()
			if err != nil {
				log.Println("AccUndo " + err.Error())

				return err
			}
		}

	}

	_, err = o.Raw(DbBindReplace("delete from acc_move_acc_cls where move_id=?"), moveId).Exec()

	if err != nil {
		log.Println("AccUndo " + err.Error())
		return err
	}

	_, err = o.Raw(DbBindReplace("delete from acc_moves where id=?"), moveId).Exec()

	if err != nil {
		log.Println("AccUndo " + err.Error())
		return err
	}

	//o.Commit()
	return nil

}
