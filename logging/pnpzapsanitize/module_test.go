package pnpzapsanitize_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/go-pnp/go-pnp/logging/pnpzapsanitize"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newSanitizedLogger(opts ...optionutil.Option[pnpzapsanitize.Options]) (*zap.Logger, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	var extractedOpts []zap.Option

	app := fx.New(
		fx.NopLogger,
		pnpzapsanitize.Module(opts...),
		fx.Invoke(func(params struct {
			fx.In
			Opts []zap.Option `group:"pnpzap.zap_options"`
		}) {
			extractedOpts = params.Opts
		}),
	)

	if err := app.Err(); err != nil {
		panic(err)
	}

	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.DebugLevel)

	return zap.New(core, extractedOpts...), buf
}

func TestSanitizer_DefaultRedactSimpleFields(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.String("password", "secret"), zap.String("user", "ok"), zap.String("Password", "secret2"))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level":    "info",
		"msg":      "test",
		"password": "[REDACTED]",
		"Password": "[REDACTED]",
		"user":     "ok",
	}
	require.Equal(t, expected, res)
}

type testStruct struct {
	Password string `json:"password"`
	User     string `json:"user"`
}

func TestSanitizer_RedactStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", testStruct{Password: "secret", User: "ok"}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"user":     "ok",
		},
	}
	require.Equal(t, expected, res)
}

type nestedStruct struct {
	APIKey string `json:"api_key"`
	Other  string `json:"other"`
}

type outerStruct struct {
	Nest nestedStruct `json:"nest"`
}

func TestSanitizer_RedactNestedStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", outerStruct{
		Nest: nestedStruct{APIKey: "secret", Other: "ok"},
	}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"nest": map[string]interface{}{
				"api_key": "[REDACTED]",
				"other":   "ok",
			},
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_RedactMap(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", map[string]interface{}{
		"token":  "secret",
		"user":   "ok",
		"nested": map[string]string{"client_secret": "secret2"},
	}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"token":  "[REDACTED]",
			"user":   "ok",
			"nested": map[string]interface{}{"client_secret": "[REDACTED]"},
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_RedactSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", []interface{}{"ok", map[string]string{"api_key": "secret"}}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": []interface{}{
			"ok",
			map[string]interface{}{"api_key": "[REDACTED]"},
		},
	}
	require.Equal(t, expected, res)
}

type circStruct struct {
	Name string      `json:"name"`
	Self *circStruct `json:"self"`
}

func TestSanitizer_RedactCircular(t *testing.T) {
	c := &circStruct{Name: "a"}
	c.Self = c
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"name": "a",
			"self": "[CIRCULAR_REFERENCE]",
		},
	}
	require.Equal(t, expected, res)
}

type baseStruct struct {
	Password string `json:"password"`
}

type derivedStruct struct {
	baseStruct

	Other string `json:"other"`
}

func TestSanitizer_RedactInlineStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", derivedStruct{baseStruct: baseStruct{Password: "secret"}, Other: "ok"}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"other":    "ok",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_RedactFieldInsideNamespace(t *testing.T) {
	logger, buf := newSanitizedLogger()
	nsLogger := logger.With(zap.Namespace("user_data"))

	nsLogger.Info("test", zap.String("api_key", "should-be-hidden"))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"user_data": map[string]interface{}{
			"api_key": "[REDACTED]",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_RedactNonStringField(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Int("client_id", 123), zap.Int("port", 8080))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level":     "info",
		"msg":       "test",
		"client_id": "[REDACTED]",
		"port":      float64(8080),
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_CustomRegex(t *testing.T) {
	re := regexp.MustCompile(`(?i)user`)
	logger, buf := newSanitizedLogger(pnpzapsanitize.WithRegex(re))
	logger.Info("test", zap.String("user", "ok"), zap.String("password", "secret"))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level":    "info",
		"msg":      "test",
		"user":     "[REDACTED]",
		"password": "secret",
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_CustomRedacted(t *testing.T) {
	logger, buf := newSanitizedLogger(pnpzapsanitize.WithRedacted("***"))
	logger.Info("test", zap.String("password", "secret"))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level":    "info",
		"msg":      "test",
		"password": "***",
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_WithContextualFields(t *testing.T) {
	logger, buf := newSanitizedLogger()
	ctxLogger := logger.With(zap.String("token", "secret"), zap.String("info", "ok"))
	ctxLogger.Info("test")
	require.NoError(t, ctxLogger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"token": "[REDACTED]",
		"info":  "ok",
	}
	require.Equal(t, expected, res)
}

type unexportedNonSensitive struct {
	internal string
	User     string `json:"user"`
}

func TestSanitizer_UnexportedNonSensitiveField(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := unexportedNonSensitive{internal: "unexported-data", User: "ok"}
	logger.Info("test", zap.Reflect("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{ // 'internal' field skipped
			"user": "ok",
		},
	}
	require.Equal(t, expected, res)
}

type structWithUnexportedMap struct {
	internal map[string]string
	Exported string `json:"exported"`
}

func TestSanitizer_UnexportedMap(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := structWithUnexportedMap{
		internal: map[string]string{"password": "secret"},
		Exported: "exported-string",
	}
	logger.Info("test", zap.Reflect("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{ // 'internal' field skipped
			"exported": "exported-string",
		},
	}
	require.Equal(t, expected, res)
}

type circularMap map[string]interface{}

func TestSanitizer_CircularMap(t *testing.T) {
	logger, buf := newSanitizedLogger()
	m := make(circularMap)
	m["self"] = m
	m["other"] = "ok"
	logger.Info("test", zap.Any("data", m))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"self":  "[CIRCULAR_REFERENCE]",
			"other": "ok",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_CircularSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()

	var s []interface{}

	s = append(s, "ok", nil)
	s[1] = s

	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": []interface{}{
			"ok",
			"[CIRCULAR_REFERENCE]",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_NilInterface(t *testing.T) {
	logger, buf := newSanitizedLogger()
	var nilInterface interface{}
	logger.Info("test", zap.Any("data", nilInterface))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  nil,
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_InterfaceHoldingStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	var iface interface{} = testStruct{Password: "secret", User: "alice"}
	logger.Info("test", zap.Any("data", iface))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"user":     "alice",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_InterfaceHoldingPointerToStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	var iface interface{} = &testStruct{Password: "secret", User: "bob"}
	logger.Info("test", zap.Any("data", iface))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"user":     "bob",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_MapWithIntKeys(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", map[int]string{
		1: "one",
		2: "two",
	}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"1": "one",
			"2": "two",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_MapWithBoolKeys(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", map[bool]string{
		true:  "yes",
		false: "no",
	}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"true":  "yes",
			"false": "no",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_NilMap(t *testing.T) {
	logger, buf := newSanitizedLogger()
	var nilMap map[string]string
	logger.Info("test", zap.Any("data", nilMap))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// zap encodes nil map as empty map before it reaches the sanitizer
	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  map[string]interface{}{},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_EmptyMap(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", map[string]string{}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  map[string]interface{}{},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_NilSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()
	var nilSlice []string
	logger.Info("test", zap.Any("data", nilSlice))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// zap encodes nil slice as empty array before it reaches the sanitizer
	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  []interface{}{},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_EmptySlice(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", []string{}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  []interface{}{},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_SliceOfPointers(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s1 := &testStruct{Password: "secret1", User: "user1"}
	s2 := &testStruct{Password: "secret2", User: "user2"}
	var nilPtr *testStruct
	logger.Info("test", zap.Any("data", []*testStruct{s1, nilPtr, s2}))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": []interface{}{
			map[string]interface{}{"password": "[REDACTED]", "user": "user1"},
			nil,
			map[string]interface{}{"password": "[REDACTED]", "user": "user2"},
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_NestedSlices(t *testing.T) {
	logger, buf := newSanitizedLogger()
	nested := [][]map[string]string{
		{{"api_key": "secret1", "name": "a"}},
		{{"password": "secret2", "value": "b"}, {"token": "secret3"}},
	}
	logger.Info("test", zap.Any("data", nested))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": []interface{}{
			[]interface{}{map[string]interface{}{"api_key": "[REDACTED]", "name": "a"}},
			[]interface{}{
				map[string]interface{}{"password": "[REDACTED]", "value": "b"},
				map[string]interface{}{"token": "[REDACTED]"},
			},
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_FunctionType(t *testing.T) {
	logger, buf := newSanitizedLogger()
	fn := func(x int) int { return x * 2 }
	logger.Info("test", zap.Any("data", fn))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	require.Equal(t, "info", res["level"])
	require.Equal(t, "test", res["msg"])
	require.Contains(t, res["data"], "<func(int) int>")
}

type structWithFunc struct {
	Handler func() `json:"handler"`
	Name    string `json:"name"`
}

func TestSanitizer_StructWithFunction(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := structWithFunc{
		Handler: func() {},
		Name:    "test-func",
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	require.Equal(t, "info", res["level"])
	data := res["data"].(map[string]interface{})
	require.Contains(t, data["handler"], "<func()>")
	require.Equal(t, "test-func", data["name"])
}

func TestSanitizer_ChannelType(t *testing.T) {
	logger, buf := newSanitizedLogger()
	ch := make(chan int, 10)
	logger.Info("test", zap.Any("data", ch))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	require.Equal(t, "info", res["level"])
	require.Contains(t, res["data"], "<chan int>")
}

type structWithChan struct {
	Events chan string `json:"events"`
	ID     int         `json:"id"`
}

func TestSanitizer_StructWithChannel(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := structWithChan{
		Events: make(chan string),
		ID:     42,
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Contains(t, data["events"], "<chan string>")
	require.Equal(t, float64(42), data["id"])
}

// --- Tests for pointers ---

func TestSanitizer_NilPointer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	var nilPtr *testStruct
	logger.Info("test", zap.Any("data", nilPtr))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  nil,
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_DoublePointer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := &testStruct{Password: "secret", User: "user1"}
	pp := &s
	logger.Info("test", zap.Any("data", pp))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"user":     "user1",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_TriplePointer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := &testStruct{Password: "secret", User: "user1"}
	pp := &s
	ppp := &pp
	logger.Info("test", zap.Any("data", ppp))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"user":     "user1",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_Complex64(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := complex64(3 + 4i)
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	require.Equal(t, "info", res["level"])
	require.Equal(t, "3+4i", res["data"])
}

func TestSanitizer_Complex128(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := complex128(1.5 + 2.5i)
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	require.Equal(t, "info", res["level"])
	require.Equal(t, "1.5+2.5i", res["data"])
}

type structWithComplex struct {
	Value complex128 `json:"value"`
	Name  string     `json:"name"`
}

func TestSanitizer_StructWithComplex(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := structWithComplex{Value: 1 + 2i, Name: "complex-test"}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "(1+2i)", data["value"])
	require.Equal(t, "complex-test", data["name"])
}

func TestSanitizer_Array(t *testing.T) {
	logger, buf := newSanitizedLogger()
	arr := [3]string{"a", "b", "c"}
	logger.Info("test", zap.Any("data", arr))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data":  []interface{}{"a", "b", "c"},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_ArrayOfMaps(t *testing.T) {
	logger, buf := newSanitizedLogger()
	arr := [2]map[string]string{
		{"password": "secret1", "name": "a"},
		{"api_key": "secret2", "value": "b"},
	}
	logger.Info("test", zap.Any("data", arr))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": []interface{}{
			map[string]interface{}{"password": "[REDACTED]", "name": "a"},
			map[string]interface{}{"api_key": "[REDACTED]", "value": "b"},
		},
	}
	require.Equal(t, expected, res)
}

type customMarshaler struct {
	Password string
	Value    string
}

func (c customMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"custom": c.Value})
}

func TestSanitizer_JSONMarshaler(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := customMarshaler{Password: "should-not-appear", Value: "custom-value"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// json.Marshaler is passed through without sanitization
	data := res["data"].(map[string]interface{})
	require.Equal(t, "custom-value", data["custom"])
	require.Nil(t, data["Password"])
}

type valueReceiverStringer struct {
	Password string
	Value    string
}

func (c valueReceiverStringer) String() string {
	return "stringer:" + c.Value
}

func TestSanitizer_FmtStringer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := valueReceiverStringer{Password: "should-not-appear", Value: "stringer-value"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// fmt.Stringer is passed through - zap will use String() method
	require.Equal(t, "stringer:stringer-value", res["data"])
}

type level3 struct {
	Token string `json:"token"`
	Value string `json:"value"`
}

type level2 struct {
	Level3   level3 `json:"level3"`
	Password string `json:"password"`
}

type level1 struct {
	Level2 level2 `json:"level2"`
	APIKey string `json:"api_key"`
}

func TestSanitizer_DeeplyNestedStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := level1{
		APIKey: "secret1",
		Level2: level2{
			Password: "secret2",
			Level3: level3{
				Token: "secret3",
				Value: "ok",
			},
		},
	}
	logger.Info("test", zap.Any("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"api_key": "[REDACTED]",
			"level2": map[string]interface{}{
				"password": "[REDACTED]",
				"level3": map[string]interface{}{
					"token": "[REDACTED]",
					"value": "ok",
				},
			},
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_DeeplyNestedMaps(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"password": "secret",
					"value":    "ok",
				},
			},
		},
	}
	logger.Info("test", zap.Any("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"password": "[REDACTED]",
						"value":    "ok",
					},
				},
			},
		},
	}
	require.Equal(t, expected, res)
}

type mixedStruct struct {
	Func    func()              `json:"func"`
	Chan    chan int            `json:"chan"`
	Map     map[string]string   `json:"map"`
	Slice   []string            `json:"slice"`
	Ptr     *testStruct         `json:"ptr"`
	Iface   interface{}         `json:"iface"`
	Complex complex128          `json:"complex"`
	Nested  map[string][]string `json:"nested"`
}

func TestSanitizer_MixedTypes(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := mixedStruct{
		Func:    func() {},
		Chan:    make(chan int),
		Map:     map[string]string{"password": "secret", "name": "ok"},
		Slice:   []string{"a", "b"},
		Ptr:     &testStruct{Password: "secret", User: "alice"},
		Iface:   map[string]string{"token": "secret"},
		Complex: 1 + 2i,
		Nested:  map[string][]string{"api_key": {"secret1", "secret2"}, "values": {"v1"}},
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Contains(t, data["func"], "<func()>")
	require.Contains(t, data["chan"], "<chan int>")
	require.Equal(t, map[string]interface{}{"password": "[REDACTED]", "name": "ok"}, data["map"])
	require.Equal(t, []interface{}{"a", "b"}, data["slice"])
	require.Equal(t, map[string]interface{}{"password": "[REDACTED]", "user": "alice"}, data["ptr"])
	require.Equal(t, map[string]interface{}{"token": "[REDACTED]"}, data["iface"])
	require.Equal(t, "(1+2i)", data["complex"])
	require.Equal(t, map[string]interface{}{
		"api_key": "[REDACTED]",
		"values":  []interface{}{"v1"},
	}, data["nested"])
}

type structWithIgnoredField struct {
	Password string `json:"-"`
	User     string `json:"user"`
}

func TestSanitizer_StructWithJsonIgnore(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := structWithIgnoredField{Password: "secret", User: "alice"}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"user": "alice",
		},
	}
	require.Equal(t, expected, res)
}

type structWithOmitempty struct {
	Password string `json:"password,omitempty"`
	User     string `json:"user,omitempty"`
	Empty    string `json:"empty,omitempty"`
}

func TestSanitizer_StructWithOmitempty(t *testing.T) {
	logger, buf := newSanitizedLogger()
	s := structWithOmitempty{Password: "secret", User: "alice", Empty: ""}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"password": "[REDACTED]",
			"user":     "alice",
			"empty":    "",
		},
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_PrimitiveTypes(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test",
		zap.Any("int", 42),
		zap.Any("float", 3.14),
		zap.Any("bool", true),
		zap.Any("string", "hello"),
	)
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level":  "info",
		"msg":    "test",
		"int":    float64(42),
		"float":  3.14,
		"bool":   true,
		"string": "hello",
	}
	require.Equal(t, expected, res)
}

func TestSanitizer_MapWithStructValues(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := map[string]testStruct{
		"user1": {Password: "secret1", User: "alice"},
		"user2": {Password: "secret2", User: "bob"},
	}
	logger.Info("test", zap.Any("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	dataMap := res["data"].(map[string]interface{})
	user1 := dataMap["user1"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", user1["password"])
	require.Equal(t, "alice", user1["user"])
	user2 := dataMap["user2"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", user2["password"])
	require.Equal(t, "bob", user2["user"])
}

func TestSanitizer_SliceOfMixedInterfaces(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := []interface{}{
		42,
		"string",
		true,
		map[string]string{"password": "secret"},
		testStruct{Password: "secret", User: "alice"},
		nil,
	}
	logger.Info("test", zap.Any("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": []interface{}{
			float64(42),
			"string",
			true,
			map[string]interface{}{"password": "[REDACTED]"},
			map[string]interface{}{"password": "[REDACTED]", "user": "alice"},
			nil,
		},
	}
	require.Equal(t, expected, res)
}

// Test []byte is preserved as base64, not converted to int array
func TestSanitizer_ByteSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := []byte(`{"user_id":18,"value":"test"}`)
	logger.Info("test", zap.Any("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// []byte should be base64 encoded, not an array of integers
	dataStr, ok := res["data"].(string)
	require.True(t, ok, "data should be a string (base64), got %T", res["data"])
	require.Equal(t, "eyJ1c2VyX2lkIjoxOCwidmFsdWUiOiJ0ZXN0In0=", dataStr)
}

func TestSanitizer_ByteSliceInStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type messageStruct struct {
		ID   int    `json:"id"`
		Data []byte `json:"data"`
	}

	msg := messageStruct{ID: 1, Data: []byte("hello")}
	logger.Info("test", zap.Any("message", msg))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	message := res["message"].(map[string]interface{})
	// Data should be base64 encoded string
	dataStr, ok := message["data"].(string)
	require.True(t, ok, "data should be a string (base64), got %T", message["data"])
	require.Equal(t, "aGVsbG8=", dataStr) // base64 of "hello"
}

// Test encoding.TextMarshaler is preserved
type customTextMarshaler struct {
	Password string
	Value    string
}

func (c customTextMarshaler) MarshalText() ([]byte, error) {
	return []byte("text:" + c.Value), nil
}

func TestSanitizer_TextMarshaler(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := customTextMarshaler{Password: "should-not-appear", Value: "marshaled-value"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// encoding.TextMarshaler is passed through - zap will use MarshalText() method
	require.Equal(t, "text:marshaled-value", res["data"])
}

// Test error interface is preserved
type customError struct {
	Code    int
	Message string
}

func (e customError) Error() string {
	return e.Message
}

func TestSanitizer_ErrorInterface(t *testing.T) {
	logger, buf := newSanitizedLogger()
	err := customError{Code: 500, Message: "something went wrong"}
	logger.Info("test", zap.Any("error", err))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// error interface is passed through - zap will use Error() method
	require.Equal(t, "something went wrong", res["error"])
}

func TestSanitizer_ErrorInStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type response struct {
		Success bool  `json:"success"`
		Error   error `json:"error"`
	}

	resp := response{Success: false, Error: customError{Code: 404, Message: "not found"}}
	logger.Info("test", zap.Any("response", resp))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	respData := res["response"].(map[string]interface{})
	require.Equal(t, "not found", respData["error"])
}

func TestSanitizer_StringerInStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type wrapper struct {
		Value fmt.Stringer `json:"value"`
		Name  string       `json:"name"`
	}

	w := wrapper{Value: valueReceiverStringer{Password: "secret", Value: "stringer-val"}, Name: "test"}
	logger.Info("test", zap.Any("data", w))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "stringer:stringer-val", data["value"])
	require.Equal(t, "test", data["name"])
}

func TestSanitizer_ErrorInMap(t *testing.T) {
	logger, buf := newSanitizedLogger()

	m := map[string]interface{}{
		"error": customError{Code: 500, Message: "internal error"},
		"other": "ok",
	}
	logger.Info("test", zap.Any("data", m))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "internal error", data["error"])
	require.Equal(t, "ok", data["other"])
}

func TestSanitizer_StringerInMap(t *testing.T) {
	logger, buf := newSanitizedLogger()

	m := map[string]interface{}{
		"stringer": valueReceiverStringer{Password: "secret", Value: "map-val"},
		"other":    "ok",
	}
	logger.Info("test", zap.Any("data", m))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "stringer:map-val", data["stringer"])
	require.Equal(t, "ok", data["other"])
}

func TestSanitizer_StringerInMapPointer(t *testing.T) {
	logger, buf := newSanitizedLogger()

	m := map[string]interface{}{
		"stringer": pointerReceiverStringer{Password: "secret", Value: "map-val"},
		"other":    "ok",
	}
	logger.Info("test", zap.Any("data", m))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "ptr_stringer:map-val", data["stringer"])
	require.Equal(t, "ok", data["other"])
}

func TestSanitizer_ErrorInSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()

	s := []interface{}{
		"ok",
		customError{Code: 400, Message: "bad request"},
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].([]interface{})
	require.Equal(t, "ok", data[0])
	require.Equal(t, "bad request", data[1])
}

func TestSanitizer_StringerInSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()

	s := []interface{}{
		"ok",
		valueReceiverStringer{Password: "secret", Value: "slice-val"},
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].([]interface{})
	require.Equal(t, "ok", data[0])
	require.Equal(t, "stringer:slice-val", data[1])
}

func TestSanitizer_PointerReceiverStringerInStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type wrapper struct {
		Value fmt.Stringer `json:"value"`
	}

	w := wrapper{Value: &pointerReceiverStringer{Password: "secret", Value: "ptr-stringer-val"}}
	logger.Info("test", zap.Any("data", w))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "ptr_stringer:ptr-stringer-val", data["value"])
}

// Test pointer receiver implementations
type pointerReceiverMarshaler struct {
	Password string
	Value    string
}

func (c *pointerReceiverMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"ptr_value": c.Value})
}

func TestSanitizer_PointerReceiverMarshaler(t *testing.T) {
	logger, buf := newSanitizedLogger()
	// Pass by value, but the type implements json.Marshaler with pointer receiver
	c := pointerReceiverMarshaler{Password: "should-not-appear", Value: "pointer-value"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// Should use the pointer receiver MarshalJSON method
	data := res["data"].(map[string]interface{})
	require.Equal(t, "pointer-value", data["ptr_value"])
	require.Nil(t, data["Password"])
}

type pointerReceiverStringer struct {
	Password string
	Value    string
}

func (c *pointerReceiverStringer) String() string {
	return "ptr_stringer:" + c.Value
}

func TestSanitizer_PointerReceiverStringer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := pointerReceiverStringer{Password: "should-not-appear", Value: "stringer-value"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// Should use the pointer receiver String() method
	require.Equal(t, "ptr_stringer:stringer-value", res["data"])
}

func TestSanitizer_ValueReceiverMarshalerAsPointer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := &customMarshaler{Password: "should-not-appear", Value: "value-as-ptr"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// Should use the value receiver MarshalJSON method
	data := res["data"].(map[string]interface{})
	require.Equal(t, "value-as-ptr", data["custom"])
	require.Nil(t, data["Password"])
}

func TestSanitizer_ValueReceiverMarshaler(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := customMarshaler{Password: "should-not-appear", Value: "value-as-ptr"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// Should use the value receiver MarshalJSON method
	data := res["data"].(map[string]interface{})
	require.Equal(t, "value-as-ptr", data["custom"])
	require.Nil(t, data["Password"])
}

func TestSanitizer_ValueReceiverStringerAsPointer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	// Pass by pointer, but the type implements fmt.Stringer with value receiver
	c := &valueReceiverStringer{Password: "should-not-appear", Value: "stringer-as-ptr"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// Should use the value receiver String() method
	require.Equal(t, "stringer:stringer-as-ptr", res["data"])
}

func TestSanitizer_ValueReceiverStringer(t *testing.T) {
	logger, buf := newSanitizedLogger()
	c := valueReceiverStringer{Password: "should-not-appear", Value: "stringer-as-ptr"}
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	// Should use the value receiver String() method
	require.Equal(t, "stringer:stringer-as-ptr", res["data"])
}

func TestSanitizer_ValueReceiverMarshalerInStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type wrapper struct {
		Value customMarshaler `json:"value"`
		Name  string          `json:"name"`
	}

	w := wrapper{Value: customMarshaler{Password: "secret", Value: "struct-val"}, Name: "test"}
	logger.Info("test", zap.Any("data", w))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	value := data["value"].(map[string]interface{})
	require.Equal(t, "struct-val", value["custom"])
	require.Equal(t, "test", data["name"])
}

func TestSanitizer_ValueReceiverMarshalerInStructAsPointer(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type wrapper struct {
		Value *customMarshaler `json:"value"`
		Name  string           `json:"name"`
	}

	w := wrapper{Value: &customMarshaler{Password: "secret", Value: "struct-val"}, Name: "test"}
	logger.Info("test", zap.Any("data", w))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	value := data["value"].(map[string]interface{})
	require.Equal(t, "struct-val", value["custom"])
	require.Equal(t, "test", data["name"])
}

func TestSanitizer_ValueReceiverStringerInStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()

	type wrapper struct {
		Value *valueReceiverStringer `json:"value"`
		Name  string                 `json:"name"`
	}

	w := wrapper{Value: &valueReceiverStringer{Password: "secret", Value: "stringer-struct-val"}, Name: "test"}
	logger.Info("test", zap.Any("data", w))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "stringer:stringer-struct-val", data["value"])
	require.Equal(t, "test", data["name"])
}

func TestSanitizer_ValueReceiverMarshalerInMap(t *testing.T) {
	logger, buf := newSanitizedLogger()

	m := map[string]interface{}{
		"marshaler": &customMarshaler{Password: "secret", Value: "map-val"},
		"other":     "ok",
	}
	logger.Info("test", zap.Any("data", m))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	marshaler := data["marshaler"].(map[string]interface{})
	require.Equal(t, "map-val", marshaler["custom"])
	require.Equal(t, "ok", data["other"])
}

func TestSanitizer_ValueReceiverStringerInMap(t *testing.T) {
	logger, buf := newSanitizedLogger()

	m := map[string]interface{}{
		"stringer": &valueReceiverStringer{Password: "secret", Value: "map-stringer-val"},
		"other":    "ok",
	}
	logger.Info("test", zap.Any("data", m))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].(map[string]interface{})
	require.Equal(t, "stringer:map-stringer-val", data["stringer"])
	require.Equal(t, "ok", data["other"])
}

func TestSanitizer_ValueReceiverMarshalerInSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()

	s := []interface{}{
		"ok",
		&customMarshaler{Password: "secret", Value: "slice-val"},
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].([]interface{})
	require.Equal(t, "ok", data[0])
	marshaler := data[1].(map[string]interface{})
	require.Equal(t, "slice-val", marshaler["custom"])
}

func TestSanitizer_ValueReceiverStringerInSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()

	s := []interface{}{
		"ok",
		&valueReceiverStringer{Password: "secret", Value: "slice-stringer-val"},
	}
	logger.Info("test", zap.Any("data", s))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	data := res["data"].([]interface{})
	require.Equal(t, "ok", data[0])
	require.Equal(t, "stringer:slice-stringer-val", data[1])
}
