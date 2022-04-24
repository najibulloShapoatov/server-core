package reflection

import (
	"errors"
	"github.com/najibulloShapoatov/server-core/monitoring/log"
	"net/http"
	"reflect"
)

type Method struct {
	Name      string
	Func      reflect.Value
	typ       reflect.Type
	paramsIn  []reflect.Type
	paramsOut []reflect.Type
}

func (m *Method) NumIn() int {
	return len(m.paramsIn)
}

func (m *Method) NumOut() int {
	return len(m.paramsOut)
}

func (m *Method) In(i int) reflect.Type {
	return m.paramsIn[i]
}

func (m *Method) Out(i int) reflect.Type {
	return m.paramsOut[i]
}

func (m *Method) Call(args ...reflect.Value) (res []interface{}) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("Error calling method %s - %v", m.Name, err)
			res = make([]interface{}, len(m.paramsOut))
			res[len(res)-1] = err
			res[len(res)-2] = http.StatusInternalServerError
		}
	}()
	if len(m.paramsIn) != len(args) {
		res = make([]interface{}, len(m.paramsOut))
		res[len(res)-1] = errors.New("invalid argument")
		res[len(res)-2] = http.StatusBadRequest
		return
	}
	out := m.Func.Call(args)
	for _, x := range out {
		if x.Kind() == reflect.Ptr && x.IsNil() {
			res = append(res, nil)
		} else {
			res = append(res, x.Interface())
		}
	}
	return
}
