package pnpzapsanitize

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	stringerType      = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	errorType         = reflect.TypeOf((*error)(nil)).Elem()
)

type fieldHidingCore struct {
	zapcore.Core

	regex    *regexp.Regexp
	redacted string
}

func NewFieldHidingCore(core zapcore.Core, regex *regexp.Regexp, redacted string) zapcore.Core {
	return &fieldHidingCore{
		Core:     core,
		regex:    regex,
		redacted: redacted,
	}
}

func (c *fieldHidingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	sanitizedFields := c.sanitizeFields(fields)

	return c.Core.Write(entry, sanitizedFields)
}

func (c *fieldHidingCore) With(fields []zapcore.Field) zapcore.Core {
	sanitizedFields := c.sanitizeFields(fields)

	return &fieldHidingCore{
		Core:     c.Core.With(sanitizedFields),
		regex:    c.regex,
		redacted: c.redacted,
	}
}

func (c *fieldHidingCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *fieldHidingCore) sanitizeFields(fields []zapcore.Field) []zapcore.Field {
	sanitizedFields := make([]zapcore.Field, 0, len(fields))
	for i := range fields {
		sanitizedFields = append(sanitizedFields, c.sanitizeField(fields[i]))
	}

	return sanitizedFields
}

func (c *fieldHidingCore) sanitizeField(field zapcore.Field) zapcore.Field {
	if c.regex.MatchString(field.Key) {
		return zap.String(field.Key, c.redacted)
	}

	switch field.Type { //nolint:exhaustive
	case zapcore.ReflectType, zapcore.ObjectMarshalerType, zapcore.ArrayMarshalerType:
		return zap.Any(field.Key, c.sanitizeInterface(field.Interface))
	default:
		return field
	}
}

func (c *fieldHidingCore) sanitizeValue(val reflect.Value, seen map[uintptr]bool) interface{} {
	if !val.IsValid() {
		return nil
	}

	if preserved, ok := c.getPreservedValue(val); ok {
		return preserved
	}

	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}

		val = val.Elem()
	}

	if !val.IsValid() {
		return nil
	}

	// Check again after dereferencing
	if preserved, ok := c.getPreservedValue(val); ok {
		return preserved
	}

	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Sprintf("<%s>", val.Type().String())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", val.Complex())
	}

	if !val.CanInterface() {
		return nil
	}

	var ptr uintptr

	addrMarked := false

	if val.Kind() == reflect.Map || val.Kind() == reflect.Slice {
		ptr = val.Pointer()
	} else if val.Kind() == reflect.Struct && val.CanAddr() {
		ptr = val.Addr().Pointer()
	}

	if ptr != 0 {
		if seen[ptr] {
			return circularRefPlaceholder
		}

		seen[ptr] = true
		addrMarked = true
	}

	var result interface{}

	switch val.Kind() { //nolint:exhaustive
	case reflect.Struct:
		result = c.sanitizeStruct(val, seen)
	case reflect.Map:
		result = c.sanitizeMap(val, seen)
	case reflect.Slice, reflect.Array:
		result = c.sanitizeSlice(val, seen)
	default:
		result = val.Interface()
	}

	if addrMarked {
		delete(seen, ptr)
	}

	return result
}

func (c *fieldHidingCore) sanitizeMap(val reflect.Value, seen map[uintptr]bool) interface{} {
	sanitizedMap := make(map[string]interface{}, val.Len())
	mapRange := val.MapRange()

	for mapRange.Next() {
		keyVal := mapRange.Key()
		if !keyVal.IsValid() || !keyVal.CanInterface() {
			continue
		}

		keyI := keyVal.Interface()

		keyStr, ok := keyI.(string)
		if !ok {
			keyStr = fmt.Sprint(keyI)
		}

		if c.regex.MatchString(keyStr) {
			sanitizedMap[keyStr] = c.redacted

			continue
		}

		valueI := c.sanitizeValue(mapRange.Value(), seen)
		sanitizedMap[keyStr] = valueI
	}

	return sanitizedMap
}

func (c *fieldHidingCore) sanitizeInterface(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	return c.sanitizeValue(reflect.ValueOf(data), make(map[uintptr]bool))
}

func (c *fieldHidingCore) sanitizeStruct(val reflect.Value, seen map[uintptr]bool) interface{} {
	sanitizedMap := make(map[string]interface{}, val.NumField())

	valType := val.Type()
	for i := range val.NumField() {
		field := valType.Field(i)
		fieldVal := val.Field(i)

		key := field.Tag.Get("json")
		if keyTagParts := strings.Split(key, ","); len(keyTagParts) > 0 {
			key = keyTagParts[0]
		}

		if key == "-" {
			continue
		}

		isInline := (key == "" && field.Anonymous && fieldVal.Kind() == reflect.Struct)

		var sanitizedVal interface{}
		if isInline {
			sanitizedVal = c.sanitizeStruct(fieldVal, seen)
		} else {
			if field.PkgPath != "" {
				continue
			}

			if fieldVal.IsValid() {
				sanitizedVal = c.sanitizeValue(fieldVal, seen)
			}
		}

		if isInline {
			if subMap, ok := sanitizedVal.(map[string]interface{}); ok {
				for k, v := range subMap {
					sanitizedMap[k] = v
				}

				continue
			}
		}

		if key == "" {
			key = field.Name
		}

		if c.regex.MatchString(key) {
			sanitizedMap[key] = c.redacted
		} else {
			sanitizedMap[key] = sanitizedVal
		}
	}

	return sanitizedMap
}

func (c *fieldHidingCore) sanitizeSlice(val reflect.Value, seen map[uintptr]bool) interface{} {
	sanitizedSlice := make([]interface{}, val.Len())
	for i := range val.Len() {
		sanitizedSlice[i] = c.sanitizeValue(val.Index(i), seen)
	}

	return sanitizedSlice
}

// getPreservedValue checks if a value should be preserved (not sanitized) because it
// implements a special interface like json.Marshaler, encoding.TextMarshaler, error, or fmt.Stringer.
func (c *fieldHidingCore) getPreservedValue(val reflect.Value) (interface{}, bool) {
	if !val.CanInterface() {
		return nil, false
	}

	t := val.Type()

	// Preserve byte slices for base64 encoding by zap
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return val.Interface(), true
	}

	if result, ok := c.getPreservedIfaceValue(val, t); ok {
		return result, true
	}

	// Check pointer type for pointer receiver implementations
	// This handles cases where the method is defined with a pointer receiver
	// but we have a value (e.g., passed by value to zap.Any)
	if t.Kind() != reflect.Ptr && c.implementsPreservedInterface(reflect.PointerTo(t)) {
		// Create an addressable copy and return a pointer to it
		// so that the pointer receiver methods can be called
		ptr := reflect.New(t)
		ptr.Elem().Set(val)

		return c.getPreservedIfaceValue(ptr, ptr.Type())
	}

	return nil, false
}

func (c *fieldHidingCore) getPreservedIfaceValue(val reflect.Value, t reflect.Type) (interface{}, bool) {
	if t.Implements(jsonMarshalerType) || t.Implements(textMarshalerType) {
		return val.Interface(), true
	}

	if t.Implements(errorType) {
		return val.Interface().(error).Error(), true
	}

	if t.Implements(stringerType) {
		return val.Interface().(fmt.Stringer).String(), true
	}

	return nil, false
}

func (c *fieldHidingCore) implementsPreservedInterface(t reflect.Type) bool {
	return t.Implements(jsonMarshalerType) ||
		t.Implements(textMarshalerType) ||
		t.Implements(errorType) ||
		t.Implements(stringerType)
}
