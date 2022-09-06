package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/cached"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

//Проведение по
//op - оперативное проведение
func AccPostByPkOper(o orm.Ormer, pk int64, operCode string, op bool, date string) (int64, error) {

	type tAcc_datas struct {
		IgnoreErrors int64  `json:"ignore_errors"`
		IsMinus      int64  `json:"is_minus"`
		EntityId     int64  `json:"entity_id"`
		Query        string `json:"query"`
		AccCls       string `json:"acc_cls"`
		AccClsId     string `json:"acc_cls_id"`
	}
	var accDatas []tAcc_datas
	var attrs []orm.Params

	entityCode := ""
	err := cached.O().Raw(DbBindReplace(`select e.code 
	from acc_opers ao
	join entities e on e.id=ao.entity_id
	where ao.code=?`), operCode).QueryRow(&entityCode)

	if err != nil {
		log.Println("AccPostByPkOper. error on get entity by operCode", operCode)
		return 0, err
	}
	_, err = cached.O().Raw(DbBindReplace(`select aop.ignore_errors as "ignore_errors",
	 ap.is_minus as "is_minus",
	  ao.entity_id as "entity_id",
	  query as "query",
	ac.code as "acc_cls",
	ac.id as "acc_cls_id"
	from 
	acc_datas ad,acc_posts ap,acc_cls ac,acc_opers ao,acc_oper_posts aop where
  ap.data_id=ad.id and ac.id=ap.cls_id and ao.id=aop.oper_id and aop.post_id=ap.id and ao.code=?
	`), operCode).QueryRows(&accDatas)

	if err != nil {
		return 0, err
	}

	//o.Begin()

	_, err = o.Raw("SAVEPOINT AccPostByPkOper").Exec()
	if err != nil {
		return 0, err
	}

	if len(accDatas) == 0 {
		return 0, nil
	}

	var accMoves TAccMoves

	accMoves.Pk = pk
	accMoves.Entity = accDatas[0].EntityId

	for _, v := range accDatas {

		if os.Getenv("CRM_VERBOSE_ACC") == "1" {
			log.Println("CRM_VERBOSE_ACC", v.Query, accMoves)
		}

		//Altenge в рамках масшабирования
		v.Query = strings.Replace(v.Query, ":acc_cls_id", fmt.Sprintf("%v", v.AccClsId), -1)

		_, err := o.Raw(DbBindReplace(v.Query), pk).Values(&attrs)

		if err != nil {
			log.Println("AccPostByPk error " + err.Error())
			return 0, err
		}
		for _, attr := range attrs {

			a := TAccMove{IsMinus: v.IsMinus == 1, AccCode: v.AccCls, Attrs: attr, IgnoreErrors: v.IgnoreErrors}
			accMoves.Moves = append(accMoves.Moves, a)
		}
		err = AccUndoByEntityIdPk(o, v.EntityId, pk)

		if err != nil {
			log.Println("AccUndoByEntityIdPk error ", err.Error())

			_, errRollBack := o.Raw("ROLLBACK TO AccPostByPkOper").Exec()
			if errRollBack != nil {
				o.Raw("ROLLBACK").Exec()
				return 0, errors.New("Error on AccUndoByEntityIdPk: " + errRollBack.Error() + " and " + err.Error())
			}
			return 0, err
		}
	}

	moveId, err := AccMove(o, accMoves, op, date)
	if os.Getenv("CRM_VERBOSE_ACC") == "1" {
		log.Println("CRM_VERBOSE_ACC", "AccMove1", accMoves, moveId)
	}

	if err != nil {
		log.Println("AccMove Error", err)
		_, errRollBack := o.Raw("ROLLBACK TO AccPostByPkOper").Exec()
		if errRollBack != nil {
			//return 0, errRollBack
			o.Raw("ROLLBACK").Exec()
			return 0, errors.New("Error on AccMove: " + errRollBack.Error() + " and " + err.Error())
		}
		return 0, errors.New("Error on AccMove: " + err.Error())
	}

	_, err = o.Raw(DbBindReplace("update "+entityCode+" set posted = 1,move_id=? where id=?"), moveId, pk).Exec()

	if err != nil {
		_, errRollBack := o.Raw("ROLLBACK TO AccPostByPkOper").Exec()
		if errRollBack != nil {
			//return 0, errRollBack
			//o.Raw("ROLLBACK").Exec()
			return 0, errors.New("Error on AccMove: " + errRollBack.Error() + " and " + err.Error())
		}
	}

	return moveId, err
}
