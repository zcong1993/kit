package breaker

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/tal-tech/go-zero/core/breaker"
)

const (
	disableBreaker = "breaker.disable"
	helpText       = "If disable breaker."
)

type OptionFactory = func() *Option

type Option struct {
	disable bool
}

func Register(cmd *cobra.Command) OptionFactory {
	var disable bool

	cmd.PersistentFlags().BoolVar(&disable, disableBreaker, false, helpText)

	return func() *Option {
		return &Option{disable: disable}
	}
}

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
