package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/go-restruct/restruct"
)

// InputFunc is the signature a decoder must implement to be registered as valid input decoder
type InputFunc func(ctx *Context, h *handler) ([]interface{}, error)

// OutputFunc is the signature a encoder must implement to be registered as valid output encoder
type OutputFunc func(ctx *Context, params ...interface{}) ([]byte, error)

var (
	// map of registered decoders
	inputDecoders = map[string]InputFunc{}
	// map of registered encoders
	outputEncoder   = map[string]OutputFunc{}
	invalidInputErr = fmt.Errorf("invalid input")
)

// RegisterDecoder registers a InputFunc decoder that can handle the given MIME type
func RegisterDecoder(contentType string, inFunc InputFunc) {
	inputDecoders[contentType] = inFunc
}

// RegisterEncoder registers a OutputFunc decoder that can handle the given MIME type
func RegisterEncoder(contentType string, outFunc OutputFunc) {
	outputEncoder[contentType] = outFunc
}

func xmlInputDecoder(ctx *Context, h *handler) (res []interface{}, err error) {
	defer func() {
		e := recover()
		if e != nil {
			res = nil
			err = invalidInputErr
		}
	}()
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}
	_ = ctx.Request.Body.Close()
	for i := 2; i < h.FuncRef.NumIn(); i++ {
		var typ = h.FuncRef.In(i)
		var x = reflect.New(typ).Interface()

		if typ.Kind() == reflect.Ptr {
			x = reflect.New(typ.Elem()).Interface()
		}
		if bytes.Equal(data, []byte("null")) {
			nilType := h.FuncRef.In(i)
			res = add(res, typ.Kind(), x, reflect.Zero(nilType).Interface())
		} else {
			if err := xml.Unmarshal(data, x); err == nil {
				res = add(res, typ.Kind(), x, x)
			} else {
				if err != nil {
					if err.Error() == "EOF" {
						return nil, invalidInputErr
					}
					switch err.(type) {
					case *xml.SyntaxError, *xml.UnmarshalError, *xml.TagPathError, *xml.UnsupportedTypeError:
						return nil, invalidInputErr
					}
				}
				nilType := h.FuncRef.In(i)
				res = add(res, typ.Kind(), x, reflect.Zero(nilType).Interface())
			}
		}
	}
	return
}

func xmlOutputEncoder(ctx *Context, params ...interface{}) ([]byte, error) {
	if len(params) == 1 {
		return xml.Marshal(params[0])
	}
	return xml.Marshal(params)
}

func jsonInputDecoder(ctx *Context, h *handler) (res []interface{}, err error) {
	defer func() {
		e := recover()
		if e != nil {
			res = nil
			err = invalidInputErr
		}
	}()
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}

	var temp = make([]json.RawMessage, 0)
	if h.FuncRef.NumIn() > 3 && bytes.HasPrefix(data, []byte("[")) {
		_ = json.Unmarshal(data, &temp)
	}

	if h.FuncRef.NumIn() > 3 && len(temp) != h.FuncRef.NumIn()-2 {
		return nil, fmt.Errorf("invalid number of input parameters")
	}

	_ = ctx.Request.Body.Close()
	for i := 2; i < h.FuncRef.NumIn(); i++ {
		var typ = h.FuncRef.In(i)
		var x = reflect.New(typ).Interface()

		if typ.Kind() == reflect.Ptr {
			x = reflect.New(typ.Elem()).Interface()
		}

		src := data
		if len(temp) != 0 {
			src = temp[i-2]
		}
		if bytes.Equal(src, []byte("null")) {
			nilType := h.FuncRef.In(i)
			res = add(res, typ.Kind(), x, reflect.Zero(nilType).Interface())
		} else {
			if err := json.Unmarshal(src, x); err == nil {
				res = add(res, typ.Kind(), x, x)
			} else {
				if err != nil {
					switch err.(type) {
					case *json.UnsupportedTypeError, *json.SyntaxError, *json.UnmarshalTypeError, *json.InvalidUnmarshalError:
						return nil, invalidInputErr
					}
				}
				nilType := h.FuncRef.In(i)
				res = add(res, typ.Kind(), x, reflect.Zero(nilType).Interface())
			}
		}
	}
	return
}

func jsonOutputEncoder(ctx *Context, params ...interface{}) ([]byte, error) {
	if len(params) == 1 {
		return json.Marshal(params[0])
	}
	return json.Marshal(params)
}

func grpcInputDecoder(ctx *Context, h *handler) (res []interface{}, err error) {
	defer func() {
		e := recover()
		if e != nil {
			res = nil
			err = invalidInputErr
		}
	}()
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}
	_ = ctx.Request.Body.Close()
	for i := 2; i < h.FuncRef.NumIn(); i++ {
		var typ = h.FuncRef.In(i)
		var x = reflect.New(typ).Interface()

		if typ.Kind() == reflect.Ptr {
			x = reflect.New(typ.Elem()).Interface()
		}
		if bytes.Equal(data, []byte("null")) {
			nilType := h.FuncRef.In(i)
			res = add(res, typ.Kind(), x, reflect.Zero(nilType).Interface())
		} else {
			if err := restruct.Unpack(data, binary.BigEndian, x); err == nil {
				res = add(res, typ.Kind(), x, x)
			} else {
				nilType := h.FuncRef.In(i)
				res = add(res, typ.Kind(), x, reflect.Zero(nilType).Interface())
			}
		}
	}
	return
}

func grpcOutputEncoder(ctx *Context, params ...interface{}) ([]byte, error) {
	if len(params) == 1 {
		return restruct.Pack(binary.BigEndian, params[0])
	}
	return restruct.Pack(binary.BigEndian, params)
}

func binaryInputDecoder(ctx *Context, h *handler) (res []interface{}, err error) {
	defer func() {
		e := recover()
		if e != nil {
			res = nil
			err = invalidInputErr
		}
	}()
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}

	add := func(typ reflect.Kind, def, item interface{}) {
		if typ == reflect.Ptr {
			res = append(res, item)
		} else {
			if item == nil {
				item = def
			}
			if reflect.TypeOf(item).Kind() == reflect.Ptr {
				res = append(res, reflect.ValueOf(item).Elem().Interface())
			} else {
				res = append(res, item)
			}
		}
	}

	_ = ctx.Request.Body.Close()
	for i := 2; i < h.FuncRef.NumIn(); i++ {
		var typ = h.FuncRef.In(i)
		var x = reflect.New(typ).Interface()

		if typ.Kind() == reflect.Ptr {
			x = reflect.New(typ.Elem()).Interface()
		}
		if bytes.Equal(data, []byte("null")) {
			nilType := h.FuncRef.In(i)
			add(typ.Kind(), x, reflect.Zero(nilType).Interface())
		} else {
			if err := restruct.Unpack(data, binary.BigEndian, x); err == nil {
				add(typ.Kind(), x, x)
			} else {
				nilType := h.FuncRef.In(i)
				add(typ.Kind(), x, reflect.Zero(nilType).Interface())
			}
		}
	}
	return
}

func add(res []interface{}, typ reflect.Kind, def, item interface{}) []interface{} {
	if typ == reflect.Ptr {
		res = append(res, item)
	} else {
		if item == nil {
			item = def
		}
		if reflect.TypeOf(item).Kind() == reflect.Ptr {
			res = append(res, reflect.ValueOf(item).Elem().Interface())
		} else {
			res = append(res, item)
		}
	}
	return res
}

func binaryOutputEncoder(ctx *Context, params ...interface{}) ([]byte, error) {
	if len(params) == 1 {
		return restruct.Pack(binary.BigEndian, params[0])
	}
	return restruct.Pack(binary.BigEndian, params)
}

func multipartInputDecoder(ctx *Context, h *handler) ([]interface{}, error) {
	return nil, nil
}

func init() {
	RegisterDecoder("text/xml", xmlInputDecoder)
	RegisterDecoder("application/xml", xmlInputDecoder)
	RegisterDecoder("text/json", jsonInputDecoder)
	RegisterDecoder("application/json", jsonInputDecoder)
	RegisterDecoder("application/grpc+octet-stream", grpcInputDecoder)
	RegisterDecoder("application/octet-stream", binaryInputDecoder)
	RegisterDecoder("multipart/form-data", multipartInputDecoder)

	RegisterEncoder("text/xml", xmlOutputEncoder)
	RegisterEncoder("application/xml", xmlOutputEncoder)
	RegisterEncoder("text/json", jsonOutputEncoder)
	RegisterEncoder("application/json", jsonOutputEncoder)
	RegisterEncoder("application/grpc+octet-stream", grpcOutputEncoder)
	RegisterEncoder("application/octet-stream", binaryOutputEncoder)
}
