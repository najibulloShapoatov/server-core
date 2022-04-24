package server

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/najibulloShapoatov/server-core/platform"
	"github.com/najibulloShapoatov/server-core/utils"
	"github.com/najibulloShapoatov/server-core/utils/reflection"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var routes = map[string]map[string]handler{}

// Register all services handlers
func RegisterRoute(module platform.Module) error {
	if _, ok := module.(platform.Service); ok {
		// Register the service name
		serviceName := strings.ToLower(module.ID()) + "-" + strings.ToLower(module.Version())
		handlers, err := analyze(module)
		if err != nil {
			return err
		}
		if len(handlers) != 0 {
			routes[serviceName] = handlers
		}
	}
	return nil
}

type handler struct {
	// trimmed method name with lowercase
	Name string
	// reference to module
	Module platform.Module
	// GET, POST, PUT, PATCH, DELETE
	HTTPMethod string
	// How is the endpoint called
	RestEndpoint string
	// reference to function and reflection
	FuncRef *reflection.Method
}

func analyze(module platform.Module) (map[string]handler, error) {

	res := map[string]handler{}
	m := reflection.New(module)
	errInterf := reflect.TypeOf((*error)(nil)).Elem()

	var ctx *Context
	// Iterate all methods and look for the ones of the form xxx(*Context.....
	for _, method := range m.Methods() {
		numOut := method.NumOut()

		// Check input params to be at least 1 and the first to be of type *server.Context
		if method.NumIn() < 2 || method.In(0).AssignableTo(reflect.TypeOf(ctx)) {
			continue
		}

		// Check output params to be at least 2 and of type (int, error)
		if numOut < 2 {
			continue
		}
		// Check return types. Last return type to be of type error or the penultimate to be of type int
		if !method.Out(numOut-1).Implements(errInterf) || method.Out(numOut-2).Kind() != reflect.Int {
			continue
		}

		h := handler{
			Module:  module,
			FuncRef: method,
		}

		switch {
		case strings.HasPrefix(method.Name, "Get"):
			h.do(http.MethodGet, []string{"Get"})

		case strings.HasPrefix(method.Name, "Create"):
			fallthrough
		case strings.HasPrefix(method.Name, "Add"):
			h.do(http.MethodPost, []string{"Add", "Create"})

		case strings.HasPrefix(method.Name, "Update"):
			fallthrough
		case strings.HasPrefix(method.Name, "Edit"):
			h.do(http.MethodPut, []string{"Edit", "Update"})

		case strings.HasPrefix(method.Name, "Delete"):
			fallthrough
		case strings.HasPrefix(method.Name, "Remove"):
			h.do(http.MethodDelete, []string{"Delete", "Remove"})

		case strings.HasPrefix(method.Name, "Do"):
			fallthrough
		default:
			h.do(http.MethodGet, []string{"Do"})
		}

		key := h.HTTPMethod + h.Name

		if previous, exists := res[key]; exists {
			return nil, fmt.Errorf("duplicate method handlers for '%s' with methods %s and %s",
				h.Name, h.FuncRef.Name, previous.FuncRef.Name)
		}

		// Register as method name and lowercase
		res[key] = h
	}

	return res, nil
}

func (h *handler) do(httpMethod string, prefixes []string) {
	// If method name Prefix is `Do` and total inbound parameters > 1 it's a POST request
	h.HTTPMethod = httpMethod
	args := make([]string, 0)
	if h.FuncRef.NumIn() > 2 && (h.HTTPMethod == http.MethodGet || h.HTTPMethod == http.MethodDelete) {
		for i := 2; i <= h.FuncRef.NumIn()-1; i++ {
			if !reflection.IsSimpleType(h.FuncRef.In(i).Kind()) {
				h.HTTPMethod = http.MethodPost
				break
			} else {
				args = append(args, ":"+h.FuncRef.In(i).Name())
			}
		}
	}

	h.Name = strings.ToLower(h.FuncRef.Name)
	for _, prefix := range prefixes {
		h.Name = strings.TrimPrefix(h.Name, strings.ToLower(prefix))
	}
	h.RestEndpoint = "/" + h.Module.ID() + "/" + h.Module.Version() + "/" + h.Name
	if len(args) != 0 && (h.HTTPMethod == http.MethodGet || h.HTTPMethod == http.MethodDelete) {
		h.RestEndpoint += "/" + strings.Join(args, "/")
	}
}

func (h *handler) Handler(ctx *Context) (err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = errors.New("bad request")
			return
		}
	}()
	var inParams = make([]reflect.Value, 0)

	inParams = append(inParams, reflect.ValueOf(h.Module))
	inParams = append(inParams, reflect.ValueOf(ctx))

	if strings.Contains(h.RestEndpoint, ":") {
		urlParts := strings.Split(ctx.Request.URL.Path, "/")
		parts := strings.Split(h.RestEndpoint, "/")
		if len(urlParts) == len(parts) {
			for idx, part := range parts {
				if !strings.HasPrefix(part, ":") {
					continue
				}
				part = strings.TrimPrefix(part, ":")
				switch part {
				case "string":
					inParams = append(inParams, reflect.ValueOf(urlParts[idx]))
				case "int":
					if utils.IsInt(urlParts[idx]) {
						intVal, err := strconv.ParseInt(urlParts[idx], 10, 64)
						if err != nil {
							ctx.BadRequest(fmt.Errorf("failed to parse argument: %s", err))
							return nil
						}
						inParams = append(inParams, reflect.ValueOf(intVal))
					} else {
						ctx.BadRequest(fmt.Errorf("failed to parse argument: %s", err))
						return nil
					}
				case "bool":
					if utils.IsTruthy(urlParts[idx]) {
						inParams = append(inParams, reflect.ValueOf(utils.Truthy(urlParts[idx])))
					} else {
						ctx.BadRequest(fmt.Errorf("failed to parse argument: %s", err))
						return nil
					}
				}
			}
		}
	}

	// determine whatever in params we can
	// and call IN decoders
	if ctx.Request.ContentLength != 0 {
		contentType := ctx.Request.Header.Get("Content-Type")
		if strings.Contains(contentType, ";") {
			contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
		}
		parser, ok := inputDecoders[contentType]
		if !ok {
			ctx.BadRequest(fmt.Errorf("invalid input format"))
			return nil
		}
		args, err := parser(ctx, h)
		if err != nil {
			ctx.BadRequest(fmt.Errorf("failed to parse input: %s", err))
			return nil
		}

		for _, x := range args {
			inParams = append(inParams, reflect.ValueOf(x))
		}
	}

	outParams := h.FuncRef.Call(inParams...)

	var outEncoder OutputFunc
	acceptEncoding := ctx.Request.Header.Get("Accept")
	acceptedEncodings := make([]string, 0)
	outContentType := ctx.Response.Header().Get("Content-Type")
	contentTypeSent := outContentType != ""

	if strings.Contains(acceptEncoding, ";") {
		acceptedEncodings = strings.Split(acceptEncoding, ";")
	} else {
		acceptedEncodings = append(acceptedEncodings, acceptEncoding)
	}

	for _, encoding := range acceptedEncodings {
		encoding = strings.TrimSpace(encoding)
		if strings.Contains(encoding, ";") {
			encoding = strings.TrimSpace(strings.Split(encoding, ";")[0])
		}
		if encoding == "*/*" {
			outEncoder = outputEncoder["application/json"]
			if outContentType == "" {
				outContentType = "application/json"
			}
			break
		} else {
			var ok bool
			if outEncoder, ok = outputEncoder[encoding]; ok {
				if outContentType == "" {
					outContentType = encoding
				}
				break
			}
		}
	}
	if outEncoder == nil {
		outEncoder = outputEncoder["application/json"]
		if outContentType == "" {
			outContentType = "application/json"
		}
	}

	if !contentTypeSent {
		ctx.Response.Header().Set("Content-Type", outContentType)
	}

	if !ctx.Response.Committed {
		ctx.Response.WriteHeader(outParams[len(outParams)-2].(int))
	}

	// Handler returned an error
	if err, ok := outParams[len(outParams)-1].(error); ok && err != nil {
		data, _ := outEncoder(ctx, struct {
			XMLName xml.Name `xml:"error" json:"-" struct:"-"`
			Error   string   `json:"error" xml:"message,attr" struct:"[64]byte"`
		}{Error: err.Error()},
		)
		_, err = ctx.Response.Write(data)
		return err
	}

	// call OUT encoder
	if len(outParams) > 2 {
		data, err := outEncoder(ctx, outParams[:len(outParams)-2]...)
		if err != nil {
			ctx.BadRequest(fmt.Errorf("invalid output format"))
			return nil
		}
		_, err = ctx.Response.Write(data)
	}

	return err
}

// Remove service handler
func UnregisterRoute(name string) {
	delete(routes, strings.ToLower(name))
}

// Remove all service handlers
func UnregisterRoutes() {
	for route := range routes {
		UnregisterRoute(route)
	}
}
