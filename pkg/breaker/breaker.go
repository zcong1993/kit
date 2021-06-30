package breaker

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/tal-tech/go-zero/core/breaker"
)

const (
	disableBreaker = "breaker.disable"
	helpText       = "If disable breaker"
)

// OptionFactory is breaker options factory.
type OptionFactory = func() *Option

// Option is breaker option.
type Option struct {
	disable bool
}

// Register register breaker flags to cobra global flags.
func Register(cmd *cobra.Command) OptionFactory {
	var disable bool

	cmd.PersistentFlags().BoolVar(&disable, disableBreaker, false, helpText)

	return func() *Option {
		return &Option{disable: disable}
	}
}

// BrkGetter store breaker by router
// and make it thread safe.
type BrkGetter struct {
	store map[string]breaker.Breaker
	mu    sync.Mutex
}

// NewBrkGetter create a new BrkGetter instance.
func NewBrkGetter() *BrkGetter {
	return &BrkGetter{
		store: make(map[string]breaker.Breaker),
	}
}

// Get get or create a breaker.
func (bg *BrkGetter) Get(key string) breaker.Breaker {
	bg.mu.Lock()
	defer bg.mu.Unlock()
	if brk, ok := bg.store[key]; ok {
		return brk
	}
	bg.store[key] = breaker.NewBreaker(breaker.WithName(key))
	return bg.store[key]
}
