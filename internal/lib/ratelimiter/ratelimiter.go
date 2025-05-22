package ratelimiter

import (
	"sync"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type WalletRateLimiter struct {
	mu       sync.Mutex
	limiters map[uuid.UUID]*rate.Limiter
	r        rate.Limit
	b        int
}

func NewWalletRateLimiter(r rate.Limit, b int) *WalletRateLimiter {
	return &WalletRateLimiter{
		limiters: make(map[uuid.UUID]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (w *WalletRateLimiter) getLimiter(walletID uuid.UUID) *rate.Limiter {
	w.mu.Lock()
	defer w.mu.Unlock()

	limiter, exists := w.limiters[walletID]
	if !exists {
		limiter = rate.NewLimiter(w.r, w.b)
		w.limiters[walletID] = limiter
	}
	return limiter
}
func (w *WalletRateLimiter) Allow(walletID uuid.UUID) bool {
	limiter := w.getLimiter(walletID)
	return limiter.Allow()
}
