package pnpzapsanitize

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultSensitiveKeysRegex = regexp.MustCompile(`(?i)password|api_?key|token|client_(id|secret|key)`)

const defaultRedactedValue = "[REDACTED]"
const circularRefPlaceholder = "[CIRCULAR_REFERENCE]"

type fieldHidingCore struct {
	zapcore.Core

	regex    *regexp.Regexp
	redacted string
}

func Module(opts ...Option) zap.Option {
	cfg := &options{
		regex:    defaultSensitiveKeysRegex,
		redacted: defaultRedactedValue,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &fieldHidingCore{
			Core:     core,
			regex:    cfg.regex,
			redacted: cfg.redacted,
		}
	})
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
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}

		val = val.Elem()
	}

	switch val.Kind() { //nolint:exhaustive
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
	default:
		return val.Interface()
	}

	var ptr uintptr

	addrMarked := false

	if val.CanAddr() {
		ptr = val.Addr().Pointer()
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
		keyI := mapRange.Key().Interface()

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
		if fieldVal.IsValid() {
			sanitizedVal = c.sanitizeValue(fieldVal, seen)
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
