package log

import (
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func LogInterface(logger log.Logger, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		level.Warn(logger).Log("func", "LogInterface", "error", err)
		return
	}
	logger.Log("func", "LogInterface", "data", string(js))
}
