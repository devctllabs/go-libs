package config

import (
	"encoding"
	"reflect"
	"strings"
)

type fieldNames struct {
	tag         string
	defaultName func(string) string
	fold        bool
}

func decodeStructOverlay(
	target any,
	values map[string]any,
	names fieldNames,
	decode func(target any) error,
) error {
	targetValue := reflect.ValueOf(target)
	decoded := reflect.New(targetValue.Elem().Type())
	if err := decode(decoded.Interface()); err != nil {
		return err
	}

	applyStructOverlay(targetValue.Elem(), decoded.Elem(), values, names)
	return nil
}

func applyStructOverlay(dst, src reflect.Value, values map[string]any, names fieldNames) {
	dstType := dst.Type()
	for i := 0; i < dstType.NumField(); i++ {
		fieldType := dstType.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}

		name, ignored := sourceFieldName(fieldType, names)
		if ignored {
			continue
		}

		raw, ok := sourceValue(values, name, names.fold)
		if !ok {
			continue
		}

		dstField := dst.Field(i)
		srcField := src.Field(i)
		if raw == nil {
			if isNullable(dstField.Kind()) {
				dstField.Set(reflect.Zero(dstField.Type()))
			}
			continue
		}

		nested, isMap := raw.(map[string]any)
		if isMap && overlaysStruct(dstField.Type()) {
			applyNestedStructOverlay(dstField, srcField, nested, names)
			continue
		}

		dstField.Set(srcField)
	}
}

func applyNestedStructOverlay(dst, src reflect.Value, values map[string]any, names fieldNames) {
	if dst.Kind() == reflect.Pointer {
		if src.IsNil() {
			return
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
		src = src.Elem()
	}

	applyStructOverlay(dst, src, values, names)
}

func sourceFieldName(field reflect.StructField, names fieldNames) (string, bool) {
	name := names.defaultName(field.Name)
	if tag, ok := field.Tag.Lookup(names.tag); ok {
		if tag == "-" {
			return "", true
		}
		if taggedName, _, _ := strings.Cut(tag, ","); taggedName != "" {
			name = taggedName
		}
	}

	return name, false
}

func sourceValue(values map[string]any, name string, fold bool) (any, bool) {
	value, ok := values[name]
	if ok || !fold {
		return value, ok
	}

	for key, value := range values {
		if strings.EqualFold(key, name) {
			return value, true
		}
	}

	return nil, false
}

func overlaysStruct(fieldType reflect.Type) bool {
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.Struct {
		return false
	}
	textUnmarshalerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	if reflect.PointerTo(fieldType).Implements(textUnmarshalerType) {
		return false
	}

	for i := 0; i < fieldType.NumField(); i++ {
		if fieldType.Field(i).PkgPath == "" {
			return true
		}
	}

	return false
}

func isStructPointer(target any) bool {
	targetValue := reflect.ValueOf(target)
	return targetValue.IsValid() &&
		targetValue.Kind() == reflect.Pointer &&
		!targetValue.IsNil() &&
		targetValue.Elem().Kind() == reflect.Struct
}

func isNullable(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	default:
		return false
	}
}
