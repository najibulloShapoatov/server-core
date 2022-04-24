package reflection

import (
	"time"
)

type DiffResult struct {
	Field string      `json:"field"`
	Old   interface{} `json:"old"`
	New   interface{} `json:"new"`
}

func Diff(original, changed interface{}) (res []DiffResult) {
	if original == nil || changed == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	a := New(original)
	b := New(changed)
	for _, name := range a.Names() {
		if !a.Field(name).IsExported() {
			continue
		}
		aval := a.Field(name).Value()
		bval := b.Field(name).Value()
		if aval == nil || bval == nil {
			return
		}
		if IsSimpleType(a.Field(name).Kind()) {
			if aval != bval {
				res = append(res, DiffResult{name, aval, bval})
			}
		} else {
			if a.Field(name).IsTime() && b.Field(name).IsTime() {
				if !aval.(time.Time).Equal(bval.(time.Time)) {
					res = append(res, DiffResult{name, aval, bval})
				}
			} else {
				changes := Diff(aval, bval)
				for _, c := range changes {
					c.Field = name + "." + c.Field
					res = append(res, c)
				}
			}
		}
	}
	return
}
