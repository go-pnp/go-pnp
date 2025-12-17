package pnpzapsanitize

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	switch field.Type {
	case zapcore.ReflectType:
		return zap.Reflect(field.Key, sanitizeInterface(field.Interface, c.regex, c.redacted, make(map[uintptr]bool), 0))
	case zapcore.ObjectMarshalerType:
		om, ok := field.Interface.(zapcore.ObjectMarshaler)
		if !ok {
			return field
		}
		return zap.Object(field.Key, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			return om.MarshalLogObject(NewSanitizingObjectEncoder(enc, c.regex, c.redacted, 0))
		}))
	case zapcore.ArrayMarshalerType:
		am, ok := field.Interface.(zapcore.ArrayMarshaler)
		if !ok {
			return field
		}
		return zap.Array(field.Key, zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
			return am.MarshalLogArray(NewSanitizingArrayEncoder(enc, c.regex, c.redacted, 0))
		}))
	default:
		return field
	}
}

const maxDepth = 100
const depthLimitPlaceholder = "[DEPTH_LIMIT_EXCEEDED]"

func sanitizeInterface(data interface{}, re *regexp.Regexp, redacted string, seen map[uintptr]bool, depth int) interface{} {
	if data == nil {
		return nil
	}

	return sanitizeValue(reflect.ValueOf(data), re, redacted, seen, depth)
}

func sanitizeValue(val reflect.Value, re *regexp.Regexp, redacted string, seen map[uintptr]bool, depth int) interface{} {
	if depth > maxDepth {
		return depthLimitPlaceholder
	}

	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}

		val = val.Elem()
	}

	if !val.IsValid() || !val.CanInterface() {
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

	switch val.Kind() {
	case reflect.Struct:
		result = sanitizeStruct(val, re, redacted, seen, depth+1)
	case reflect.Map:
		result = sanitizeMap(val, re, redacted, seen, depth+1)
	case reflect.Slice, reflect.Array:
		result = sanitizeSlice(val, re, redacted, seen, depth+1)
	default:
		result = val.Interface()
	}

	if addrMarked {
		delete(seen, ptr)
	}

	return result
}

func sanitizeMap(val reflect.Value, re *regexp.Regexp, redacted string, seen map[uintptr]bool, depth int) interface{} {
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

		if re.MatchString(keyStr) {
			sanitizedMap[keyStr] = redacted
			continue
		}

		valueI := sanitizeValue(mapRange.Value(), re, redacted, seen, depth+1)
		sanitizedMap[keyStr] = valueI
	}

	return sanitizedMap
}

func sanitizeStruct(val reflect.Value, re *regexp.Regexp, redacted string, seen map[uintptr]bool, depth int) interface{} {
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
			sanitizedVal = sanitizeStruct(fieldVal, re, redacted, seen, depth+1)
		} else {
			if field.PkgPath != "" {
				continue
			}

			if fieldVal.IsValid() {
				sanitizedVal = sanitizeValue(fieldVal, re, redacted, seen, depth+1)
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

		if re.MatchString(key) {
			sanitizedMap[key] = redacted
		} else {
			sanitizedMap[key] = sanitizedVal
		}
	}

	return sanitizedMap
}

func sanitizeSlice(val reflect.Value, re *regexp.Regexp, redacted string, seen map[uintptr]bool, depth int) interface{} {
	sanitizedSlice := make([]interface{}, val.Len())
	for i := range val.Len() {
		sanitizedSlice[i] = sanitizeValue(val.Index(i), re, redacted, seen, depth+1)
	}

	return sanitizedSlice
}

type sanitizingObjectEncoder struct {
	zapcore.ObjectEncoder
	regex    *regexp.Regexp
	redacted string
	depth    int
}

func NewSanitizingObjectEncoder(enc zapcore.ObjectEncoder, re *regexp.Regexp, redacted string, depth int) zapcore.ObjectEncoder {
	return &sanitizingObjectEncoder{
		ObjectEncoder: enc,
		regex:         re,
		redacted:      redacted,
		depth:         depth,
	}
}

func (s *sanitizingObjectEncoder) AddBool(key string, value bool) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddBool(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddByteString(key string, value []byte) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddByteString(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddComplex128(key string, value complex128) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddComplex128(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddComplex64(key string, value complex64) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddComplex64(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddDuration(key string, value time.Duration) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddDuration(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddFloat64(key string, value float64) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddFloat64(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddFloat32(key string, value float32) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddFloat32(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddInt(key string, value int) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddInt(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddInt64(key string, value int64) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddInt64(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddInt32(key string, value int32) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddInt32(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddInt16(key string, value int16) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddInt16(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddInt8(key string, value int8) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddInt8(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddString(key, value string) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddString(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddTime(key string, value time.Time) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddTime(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddUint(key string, value uint) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddUint(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddUint64(key string, value uint64) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddUint64(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddUint32(key string, value uint32) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddUint32(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddUint16(key string, value uint16) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddUint16(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddUint8(key string, value uint8) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddUint8(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddUintptr(key string, value uintptr) {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
	} else {
		s.ObjectEncoder.AddUintptr(key, value)
	}
}

func (s *sanitizingObjectEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
		return nil
	}
	if s.depth >= maxDepth {
		s.ObjectEncoder.AddString(key, depthLimitPlaceholder)
		return nil
	}
	return s.ObjectEncoder.AddArray(key, zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		return marshaler.MarshalLogArray(NewSanitizingArrayEncoder(enc, s.regex, s.redacted, s.depth+1))
	}))
}

func (s *sanitizingObjectEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
		return nil
	}
	if s.depth >= maxDepth {
		s.ObjectEncoder.AddString(key, depthLimitPlaceholder)
		return nil
	}
	return s.ObjectEncoder.AddObject(key, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		return marshaler.MarshalLogObject(NewSanitizingObjectEncoder(enc, s.regex, s.redacted, s.depth+1))
	}))
}

func (s *sanitizingObjectEncoder) AddReflected(key string, value interface{}) error {
	if s.regex.MatchString(key) {
		s.ObjectEncoder.AddString(key, s.redacted)
		return nil
	}
	if s.depth >= maxDepth {
		s.ObjectEncoder.AddString(key, depthLimitPlaceholder)
		return nil
	}
	sanitized := sanitizeInterface(value, s.regex, s.redacted, make(map[uintptr]bool), s.depth+1)
	return s.ObjectEncoder.AddReflected(key, sanitized)
}

func (s *sanitizingObjectEncoder) OpenNamespace(key string) {
	s.ObjectEncoder.OpenNamespace(key)
}

type sanitizingArrayEncoder struct {
	zapcore.ArrayEncoder
	regex    *regexp.Regexp
	redacted string
	depth    int
}

func NewSanitizingArrayEncoder(enc zapcore.ArrayEncoder, re *regexp.Regexp, redacted string, depth int) zapcore.ArrayEncoder {
	return &sanitizingArrayEncoder{
		ArrayEncoder: enc,
		regex:        re,
		redacted:     redacted,
		depth:        depth,
	}
}

func (s *sanitizingArrayEncoder) AppendBool(value bool) {
	s.ArrayEncoder.AppendBool(value)
}

func (s *sanitizingArrayEncoder) AppendByteString(value []byte) {
	s.ArrayEncoder.AppendByteString(value)
}

func (s *sanitizingArrayEncoder) AppendComplex128(value complex128) {
	s.ArrayEncoder.AppendComplex128(value)
}

func (s *sanitizingArrayEncoder) AppendComplex64(value complex64) {
	s.ArrayEncoder.AppendComplex64(value)
}

func (s *sanitizingArrayEncoder) AppendDuration(value time.Duration) {
	s.ArrayEncoder.AppendDuration(value)
}

func (s *sanitizingArrayEncoder) AppendFloat64(value float64) {
	s.ArrayEncoder.AppendFloat64(value)
}

func (s *sanitizingArrayEncoder) AppendFloat32(value float32) {
	s.ArrayEncoder.AppendFloat32(value)
}

func (s *sanitizingArrayEncoder) AppendInt(value int) {
	s.ArrayEncoder.AppendInt(value)
}

func (s *sanitizingArrayEncoder) AppendInt64(value int64) {
	s.ArrayEncoder.AppendInt64(value)
}

func (s *sanitizingArrayEncoder) AppendInt32(value int32) {
	s.ArrayEncoder.AppendInt32(value)
}

func (s *sanitizingArrayEncoder) AppendInt16(value int16) {
	s.ArrayEncoder.AppendInt16(value)
}

func (s *sanitizingArrayEncoder) AppendInt8(value int8) {
	s.ArrayEncoder.AppendInt8(value)
}

func (s *sanitizingArrayEncoder) AppendString(value string) {
	s.ArrayEncoder.AppendString(value)
}

func (s *sanitizingArrayEncoder) AppendTime(value time.Time) {
	s.ArrayEncoder.AppendTime(value)
}

func (s *sanitizingArrayEncoder) AppendUint(value uint) {
	s.ArrayEncoder.AppendUint(value)
}

func (s *sanitizingArrayEncoder) AppendUint64(value uint64) {
	s.ArrayEncoder.AppendUint64(value)
}

func (s *sanitizingArrayEncoder) AppendUint32(value uint32) {
	s.ArrayEncoder.AppendUint32(value)
}

func (s *sanitizingArrayEncoder) AppendUint16(value uint16) {
	s.ArrayEncoder.AppendUint16(value)
}

func (s *sanitizingArrayEncoder) AppendUint8(value uint8) {
	s.ArrayEncoder.AppendUint8(value)
}

func (s *sanitizingArrayEncoder) AppendUintptr(value uintptr) {
	s.ArrayEncoder.AppendUintptr(value)
}

func (s *sanitizingArrayEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	if s.depth >= maxDepth {
		s.ArrayEncoder.AppendString(depthLimitPlaceholder)
		return nil
	}
	return s.ArrayEncoder.AppendArray(zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		return marshaler.MarshalLogArray(NewSanitizingArrayEncoder(enc, s.regex, s.redacted, s.depth+1))
	}))
}

func (s *sanitizingArrayEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	if s.depth >= maxDepth {
		s.ArrayEncoder.AppendString(depthLimitPlaceholder)
		return nil
	}
	return s.ArrayEncoder.AppendObject(zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		return marshaler.MarshalLogObject(NewSanitizingObjectEncoder(enc, s.regex, s.redacted, s.depth+1))
	}))
}

func (s *sanitizingArrayEncoder) AppendReflected(value interface{}) error {
	if s.depth >= maxDepth {
		s.ArrayEncoder.AppendString(depthLimitPlaceholder)
		return nil
	}
	sanitized := sanitizeInterface(value, s.regex, s.redacted, make(map[uintptr]bool), s.depth+1)
	return s.ArrayEncoder.AppendReflected(sanitized)
}
