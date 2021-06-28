package log

import "go.uber.org/zap"

const (
	ServiceKey   = "service"
	ComponentKey = "component"
	ErrorKey     = "error"
)

func Component(name string) zap.Field {
	return zap.String(ComponentKey, name)
}

func Service(name string) zap.Field {
	return zap.String(ServiceKey, name)
}

func ErrorMsg(err error) zap.Field {
	return zap.String(ErrorKey, err.Error())
}
