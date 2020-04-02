package gateway

import (
	"fmt"
	"reflect"
	"time"
)

var (
	apis []*API
)

func RegisterApiWithExppire(group string, key, name string, handler Handler, expire time.Duration) {

	req, resp, reqType := getHandlerInOutParamInfo(handler)

	// 构建接口文档
	apis = append(apis, &API{
		LineNum:  0,
		Key:      key,
		Name:     name,
		Group:    group,
		ReqType:  reqType,
		Request:  req,
		Response: resp,
	})

	apiKey := fmt.Sprintf("%s.%s", group, key)
	//  注册到中间件中
	if _, ok := apiHandlerFuncMap[apiKey]; ok {
		panic(fmt.Errorf("%s already registered", key))
	}
	apiHandlerFuncMap[apiKey] = &HandlerInfo{
		reqType: reqType,
		handler: reflect.ValueOf(handler).Type(),
		expire:  expire,
	}
}

// RegisterAPI 格式化的返回
func RegisterAPI(group string, key, name string, handler Handler) {
	RegisterApiWithExppire(group, key, name, handler, time.Minute*10)
}

func getHandlerInOutParamInfo(handler Handler) (in, out *DTOInfo, reqType reflect.Type) {
	req, ok := reflect.ValueOf(handler).Type().FieldByName("Request")
	if !ok {
		panic("not contains Request field")
	}

	return getDTOFieldInfo(reflect.ValueOf(handler).FieldByName("Request").Type()), getDTOFieldInfo(reflect.ValueOf(handler).FieldByName("Response").Type()), req.Type
}

func getDTOFieldInfo(dto reflect.Type) *DTOInfo {

	fields := make([]*FieldInfo, 0, 10)
	types := make([]*TypeInfo, 0)
	for i := 0; i < dto.NumField(); i++ {
		field := dto.Field(i)
		tag := field.Tag

		filedInfo := FieldInfo{
			name:     tag.Get("json"),
			desc:     tag.Get("desc"),
			typ:      field.Type.String(),
			required: true,
			note:     "todo for binding",
		}

		// fmt.Printf("%s\t%d\n", field.Type.String(), field.Type.Kind())

		if field.Type.Kind() == reflect.Slice {
			types = append(types, getTypeInfo(field.Type.Elem())...)

			// field.Type.Elem().
		}
		if field.Type.Kind() == reflect.Struct {
			types = append(types, getTypeInfo(field.Type)...)
		}

		fields = append(fields, &filedInfo)
	}

	return &DTOInfo{
		fields: fields,
		types:  types,
	}
}

func getTypeInfo(fieldType reflect.Type) []*TypeInfo {

	name := fieldType.String()
	types := make([]*TypeInfo, 0)

	if name == "time.Time" {
		return types
	}

	fields := make([]*FieldInfo, 0, 10)

	for i := 0; i < fieldType.NumField(); i++ {
		field := fieldType.Field(i)
		tag := field.Tag

		filedInfo := FieldInfo{
			name:     tag.Get("json"),
			desc:     tag.Get("desc"),
			typ:      field.Type.String(),
			required: true,
			note:     "todo for binding",
		}

		fields = append(fields, &filedInfo)
		if field.Type.Kind() == reflect.Struct {
			info := getTypeInfo(field.Type)
			types = append(types, info...)
		}
	}

	typeInfo := &TypeInfo{
		name:   name,
		fields: fields,
	}
	types = append([]*TypeInfo{typeInfo}, types...)

	return types
}
