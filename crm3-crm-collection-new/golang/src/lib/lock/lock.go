package lock

import (
	"sync"
	"sync/atomic"
)

type Lock struct {
	m   sync.Map
	seq uint32
}

func (l *Lock) Acquire(key string) (uint32, bool) {
	id := atomic.AddUint32(&l.seq, 1)
	_, loaded := l.m.LoadOrStore(key, id)
	if loaded {
		return 0, false
	}
	return id, true
}

func (l *Lock) Release(key string, id uint32) bool {
	v, ok := l.m.Load(key)
	if !ok {
		return false
	}

	old, ok := v.(uint32)
	if !ok {
		return false
	}
	if old != id {
		return false
	}
	l.m.Delete(key)
	return true
}

var global_lock Lock

func Global() *Lock {
	return &global_lock

}
