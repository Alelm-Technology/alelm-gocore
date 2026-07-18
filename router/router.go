package router

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

func Register[API any](r *gin.RouterGroup, api API) {
	t := reflect.TypeOf(api)
	basePath := "/" + urlize(t.Name())

	for i := range t.NumMethod() {
		method := t.Method(i)
		httpMethod, routePath := parseMethodName(method.Name)
		if httpMethod == "" {
			continue
		}
		if method.Type.NumIn() != 1 {
			continue
		}
		fullPath := basePath
		if routePath != "" {
			fullPath += "/" + routePath
		}
		handlerFn := reflect.ValueOf(api).MethodByName(method.Name).Call(nil)
		if len(handlerFn) != 1 {
			continue
		}
		handler, ok := handlerFn[0].Interface().(gin.HandlerFunc)
		if !ok {
			continue
		}
		r.Handle(httpMethod, fullPath, handler)
	}
}

var httpPrefixes = []struct {
	method string
	prefix string
}{
	{"GET", "Get"}, {"POST", "Post"}, {"PUT", "Put"},
	{"PATCH", "Patch"}, {"DELETE", "Delete"}, {"HEAD", "Head"},
}

func parseMethodName(name string) (httpMethod string, routePath string) {
	for _, mp := range httpPrefixes {
		if strings.HasPrefix(name, mp.prefix) {
			return mp.method, urlize(name[len(mp.prefix):])
		}
	}
	return "", ""
}

func urlize(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('-')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func Validate[API any]() error {
	t := reflect.TypeOf((*API)(nil)).Elem()
	for i := range t.NumMethod() {
		method := t.Method(i)
		httpMethod, _ := parseMethodName(method.Name)
		if httpMethod == "" {
			return fmt.Errorf("method %s does not start with HTTP prefix (Get/Post/Put/Patch/Delete)", method.Name)
		}
		mt := method.Type
		if mt.NumOut() != 1 || mt.Out(0) != reflect.TypeOf((*gin.HandlerFunc)(nil)).Elem() {
			return fmt.Errorf("method %s must return gin.HandlerFunc", method.Name)
		}
	}
	return nil
}
