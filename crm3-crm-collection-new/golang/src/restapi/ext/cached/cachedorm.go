package cached

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"sync"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

type dbresult struct {
	cnt int64
	val []interface{}
}

type valsMap map[string]*dbresult
type queryMap map[string]valsMap
type CachedOrm struct {
	sync.Mutex
	o     orm.Ormer
	cache queryMap
}

// raw query seter
type rawSet struct {
	query string
	args  []interface{}
	orm   *CachedOrm
}

var _cachedOrm CachedOrm
var _onceInit sync.Once

func cachedOrmCreate() {
	o := orm.NewOrm()
	o.Using("default")
	_cachedOrm.o = o
	_cachedOrm.cache = make(queryMap)
}
func cachedOrmSingletone() *CachedOrm {
	_onceInit.Do(func() {
		cachedOrmCreate()
	})
	return &_cachedOrm
}

func O() orm.Ormer {
	return cachedOrmSingletone()
}
func ClearCache() {
	cachedOrmSingletone().ClearCache()
}

//----------------------------
func (o *CachedOrm) ClearCache() {
	o.Lock()
	defer o.Unlock()
	o.cache = make(queryMap)
}
func (o *CachedOrm) BeginTx(ctx context.Context, opts *sql.TxOptions) error {
	panic("not implemented")
}
func (o *CachedOrm) Read(md interface{}, cols ...string) error {
	panic("not implemented")
}
func (o *CachedOrm) ReadForUpdate(md interface{}, cols ...string) error {
	panic("not implemented")
}
func (o *CachedOrm) ReadOrCreate(md interface{}, col1 string, cols ...string) (bool, int64, error) {
	panic("not implemented")
}
func (o *CachedOrm) Insert(interface{}) (int64, error) {
	panic("not implemented")
}

func (o *CachedOrm) InsertOrUpdate(md interface{}, colConflitAndArgs ...string) (int64, error) {
	panic("not implemented")
}
func (o *CachedOrm) InsertMulti(bulk int, mds interface{}) (int64, error) {
	panic("not implemented")
}

func (o *CachedOrm) Update(md interface{}, cols ...string) (int64, error) {
	panic("not implemented")
}
func (o *CachedOrm) Delete(md interface{}, cols ...string) (int64, error) {
	panic("not implemented")
}
func (o *CachedOrm) LoadRelated(md interface{}, name string, args ...interface{}) (int64, error) {
	panic("not implemented")
}
func (o *CachedOrm) QueryM2M(md interface{}, name string) orm.QueryM2Mer {
	panic("not implemented")
}
func (o *CachedOrm) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	panic("not implemented")
}
func (o *CachedOrm) Using(name string) error {
	panic("not implemented")
}
func (o *CachedOrm) Begin() error {
	panic("not implemented")
}
func (o *CachedOrm) Commit() error {
	panic("not implemented")
}
func (o *CachedOrm) Rollback() error {
	panic("not implemented")
}
func (o *CachedOrm) Raw(query string, args ...interface{}) orm.RawSeter {
	r := new(rawSet)
	r.query = query
	r.args = args
	r.orm = o
	return r
}
func (o *CachedOrm) Driver() orm.Driver {
	panic("not implemented")
}

//----------------------------

func (r *rawSet) Exec() (sql.Result, error) {
	panic("not implemented")
}
func (r *rawSet) QueryRow(containers ...interface{}) error {

	b, err := json.Marshal(r.args)
	if err != nil {
		panic(err)
	}
	key := "QueryRow:" + string(b)

	r.orm.Lock()
	vals, queryExists := r.orm.cache[r.query]
	if !queryExists {
		vals = make(valsMap)
		r.orm.cache[r.query] = vals
	}

	res, valueExists := vals[key]
	if !valueExists {
		err := r.orm.o.Raw(r.query, r.args...).QueryRow(containers...)
		if err == nil {
			res = new(dbresult)
			vals[key] = res
			res.val = containers
		}
		r.orm.Unlock()
		return err
	} else {
		r.orm.Unlock()
		for i, container := range containers {
			p := reflect.ValueOf(container)
			v := reflect.ValueOf(res.val[i])
			p.Elem().Set(v.Elem())
		}
		return nil
	}
}
func (r *rawSet) QueryRows(containers ...interface{}) (int64, error) {
	b, err := json.Marshal(r.args)
	if err != nil {
		panic(err)
	}
	key := "QueryRows:" + string(b)
	r.orm.Lock()
	vals, queryExists := r.orm.cache[r.query]
	if !queryExists {
		vals = make(valsMap)
		r.orm.cache[r.query] = vals
	}

	res, valueExists := vals[key]
	if !valueExists {
		cnt, err := r.orm.o.Raw(r.query, r.args...).QueryRows(containers...)
		if err == nil {
			res = new(dbresult)
			vals[key] = res
			res.cnt = cnt
			res.val = containers
		}
		r.orm.Unlock()
		return cnt, err
	} else {
		r.orm.Unlock()
		for i, container := range containers {
			p := reflect.ValueOf(container)
			v := reflect.ValueOf(res.val[i])
			p.Elem().Set(v.Elem())
		}
		return res.cnt, nil
	}

}
func (r *rawSet) SetArgs(...interface{}) orm.RawSeter {
	panic("not implemented")
}
func (r *rawSet) Values(container *[]orm.Params, cols ...string) (int64, error) {
	panic("not implemented")
}
func (r *rawSet) ValuesList(container *[]orm.ParamsList, cols ...string) (int64, error) {
	panic("not implemented")
}
func (r *rawSet) ValuesFlat(container *orm.ParamsList, cols ...string) (int64, error) {
	panic("not implemented")
}
func (r *rawSet) RowsToMap(result *orm.Params, keyCol, valueCol string) (int64, error) {
	panic("not implemented")
}
func (r *rawSet) RowsToStruct(ptrStruct interface{}, keyCol, valueCol string) (int64, error) {
	panic("not implemented")
}
func (r *rawSet) Prepare() (orm.RawPreparer, error) {
	panic("not implemented")
}
func (r *rawSet) Columns() ([]string, error) {
	panic("not implemented")
}
