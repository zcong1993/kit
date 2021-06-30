package log

import "go.uber.org/zap"

const (
	// ServiceKey is zap field key for service.
	ServiceKey = "service"
	// ComponentKey is zap field key for component.
	ComponentKey = "component"
	// ErrorKey is zap field key for error.
	ErrorKey = "error"
)

// Component add zap field with component key.
func Component(name string) zap.Field {
	return zap.String(ComponentKey, name)
}

// Service add zap field with service key.
func Service(name string) zap.Field {
	return zap.String(ServiceKey, name)
}

// ErrorMsg add zap field with error key.
func ErrorMsg(err error) zap.Field {
	return zap.String(ErrorKey, err.Error())
}
