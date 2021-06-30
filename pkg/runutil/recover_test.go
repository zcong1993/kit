package runutil_test

import (
	"testing"

	"github.com/zcong1993/kit/pkg/log"
	"github.com/zcong1993/kit/pkg/runutil"
)

func TestWithRecover(t *testing.T) {
	logger, _ := log.DefaultLogger()
	runutil.WithRecover(func() {
		panic(map[string]string{"a": "x"})
	}, logger)
}
