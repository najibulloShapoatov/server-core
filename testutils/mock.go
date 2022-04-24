package testutils

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

type anything = int
type ofType = int

type fnCall struct {
	fn   interface{}
	args []interface{}
}

type staged struct {
	fn     interface{}
	result []interface{}
}

type mockTime struct {
	d time.Duration
}

type Mock struct {
	calls []*fnCall
	stage []*staged
}

func NewMock() *Mock {
	return &Mock{
		calls: nil,
	}
}

func (m *Mock) ResetCalls() {
	m.calls = nil
}

func (m *Mock) Reset() {
	m.calls = nil
	m.stage = nil
}

func (m *Mock) Anything() anything {
	return anything(1)
}

func (m *Mock) OfType() ofType {
	return ofType(1)
}

func (m *Mock) Trace(fn interface{}, args ...interface{}) {
	m.calls = append(m.calls, &fnCall{fn, args})
}

func (m *Mock) Stage(fn interface{}, result ...interface{}) {
	m.stage = append(m.stage, &staged{fn, result})
}

func (m *Mock) TimeRange(duration time.Duration) mockTime {
	return mockTime{duration}
}

func (m *Mock) FromStage(fn interface{}, resultPtr ...interface{}) {
	for i, staged := range m.stage {
		if runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name() ==
			runtime.FuncForPC(reflect.ValueOf(staged.fn).Pointer()).Name() {
			if len(resultPtr) != len(staged.result) {
				continue
			}
			for idx, val := range staged.result {
				targetVal := reflect.ValueOf(val)
				destVal := indirect(reflect.ValueOf(resultPtr[idx]), true)

				if val != nil {

					if _, ok := val.(ofType); ok {
						x := indirect(destVal, false)
						targetVal = x
						if destVal.Kind() == reflect.Ptr {
							targetVal = targetVal.Addr()
						}
					}

					v := indirect(destVal, true)
					if v.CanSet() {
						v.Set(targetVal)
					}
				}
			}

			if len(m.stage) == 1 {
				m.stage = m.stage[:0]
			} else {
				m.stage = append(m.stage[:i], m.stage[i+1:]...)
			}
			return
		}
	}
}

func (m *Mock) WasCalledNTimes(fn interface{}, n int) error {
	var count int
	for _, call := range m.calls {
		if runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name() ==
			runtime.FuncForPC(reflect.ValueOf(call.fn).Pointer()).Name() {
			count++
		}
	}
	if n != count {
		return fmt.Errorf("expected function to be called %d times but was called %d times", n, count)
	}
	return nil
}

func (m *Mock) WasCalledWith(fn interface{}, args ...interface{}) error {
	var calledWithOtherArgs error
	for _, call := range m.calls {
		if runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name() !=
			runtime.FuncForPC(reflect.ValueOf(call.fn).Pointer()).Name() {
			continue
		}
		if len(args) != len(call.args) {
			continue
		}
		var allMatched = true
		for idx, a := range args {
			expectedArg := a
			realArg := call.args[idx]

			if isNil(expectedArg) && isNil(realArg) {
				continue
			}
			if mTime, ok := expectedArg.(mockTime); ok {
				var t time.Time
				if t1, ok := realArg.(time.Time); ok {
					t = t1
				} else if t1, ok := realArg.(*time.Time); ok {
					if t1 != nil {
						t = *t1
					}
				}
				now := time.Now()
				if !now.Add(mTime.d).After(t) && !now.Add(-mTime.d).Before(t) {
					calledWithOtherArgs = fmt.Errorf("expected argument %d to be %v but was %v", idx, expectedArg, realArg)
					allMatched = false
					break
				}
			} else if _, ok := expectedArg.(anything); ok {
				//
			} else if !reflect.DeepEqual(expectedArg, realArg) {
				calledWithOtherArgs = fmt.Errorf("expected argument %d to be %v but was %v", idx, expectedArg, realArg)
				allMatched = false
				break
			}
		}
		if allMatched {
			return nil
		}
	}

	if calledWithOtherArgs != nil {
		return calledWithOtherArgs
	}
	return errors.New("function was never called")
}

func isNil(v interface{}) bool {
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr || reflect.ValueOf(v).Kind() == reflect.Slice || reflect.ValueOf(v).Kind() == reflect.Map) && reflect.ValueOf(v).IsNil() {
		return true
	}
	return false
}

func (m *Mock) LastCalledWith(fn interface{}, args ...interface{}) error {
	for idx := len(m.calls); idx > 0; idx-- {
		call := m.calls[idx]

		if runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name() !=
			runtime.FuncForPC(reflect.ValueOf(call.fn).Pointer()).Name() {
			continue
		}

		if len(args) != len(call.args) {
			return fmt.Errorf("exepected function to have been called with %d argumets but was called with %d arguments", len(args), len(call.args))
		}
		for idx, a := range args {
			if !reflect.DeepEqual(a, call.args[idx]) {
				return fmt.Errorf("expected argument %d to be %v but was %v", idx, a, call.args[idx])
			}
		}
		return nil
	}
	return errors.New("function was never called")
}

func indirect(v reflect.Value, decodingNull bool) reflect.Value {
	v0 := v
	haveAddr := false

	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				haveAddr = false
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if decodingNull && v.CanSet() {
			break
		}

		// Prevent infinite loop if v is an interface pointing to its own address:
		//     var v interface{}
		//     v = &v
		if v.Elem().Kind() == reflect.Interface && v.Elem().Elem() == v {
			v = v.Elem()
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		if haveAddr {
			v = v0 // restore original value after round-trip Value.Addr().Elem()
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return v
}
