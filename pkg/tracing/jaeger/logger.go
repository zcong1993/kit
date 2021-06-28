// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package jaeger

import (
	"fmt"

	"github.com/zcong1993/x/pkg/log"
)

type jaegerLogger struct {
	logger *log.Logger
}

func (l *jaegerLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *jaegerLogger) Error(msg string) {
	l.logger.Error(msg)
}
