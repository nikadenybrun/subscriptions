package walletlocks

import (
	"sync"

	"github.com/google/uuid"
)

type WalletLocks struct {
	m map[uuid.UUID]*sync.Mutex
	sync.Mutex
}

func NewWalletLocks() *WalletLocks {
	return &WalletLocks{
		m: make(map[uuid.UUID]*sync.Mutex),
	}
}
func (wl *WalletLocks) GetWalletLock(id uuid.UUID) *sync.Mutex {
	wl.Lock()
	defer wl.Unlock()
	if _, exists := wl.m[id]; !exists {
		wl.m[id] = &sync.Mutex{}
	}
	return wl.m[id]
}
