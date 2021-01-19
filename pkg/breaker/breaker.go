package breaker

import (
	"sync"

	"github.com/tal-tech/go-zero/core/breaker"
)

type BrkGetter struct {
	store map[string]breaker.Breaker
	mu    sync.Mutex
}

func NewBrkGetter() *BrkGetter {
	return &BrkGetter{
		store: make(map[string]breaker.Breaker),
	}
}

func (bg *BrkGetter) Get(key string) breaker.Breaker {
	bg.mu.Lock()
	defer bg.mu.Unlock()
	if brk, ok := bg.store[key]; ok {
		return brk
	}
	bg.store[key] = breaker.NewBreaker(breaker.WithName(key))
	return bg.store[key]
}
