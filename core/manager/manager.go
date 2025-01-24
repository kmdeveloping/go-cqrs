package manager

import (
	"reflect"
)

type ICqrsManager interface {
	ExecuteTask(any, any) error
	ExecuteTaskWithResult(any) (any, error)
}

type CqrsManager struct {
}

var _ ICqrsManager = (*CqrsManager)(nil)

func NewCqrsManager() ICqrsManager {
	return &CqrsManager{}
}

func (m *CqrsManager) ExecuteTask(THandler any, TCommand any) error {
	task := []reflect.Value{}
	t := append(task, reflect.ValueOf(TCommand))

	reflect.ValueOf(THandler).MethodByName("execute").Call(t)
	return nil
}

func (m *CqrsManager) ExecuteTaskWithResult(T any) (any, error) {
	return nil, nil
}
