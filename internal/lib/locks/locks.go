package locks

import (
	"sync"
)

type Locks struct {
	m map[string]*sync.Mutex
	sync.Mutex
}

func NewLocks() *Locks {
	return &Locks{
		m: make(map[string]*sync.Mutex),
	}
}
func (wl *Locks) GetLock(id string) *sync.Mutex {
	wl.Lock()
	defer wl.Unlock()
	if _, exists := wl.m[id]; !exists {
		wl.m[id] = &sync.Mutex{}
	}
	return wl.m[id]
}
