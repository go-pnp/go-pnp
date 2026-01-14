package pnpzapsanitize_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/go-pnp/go-pnp/logging/pnpzapsanitize"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newSanitizedLogger(opts ...pnpzapsanitize.Option) (*zap.Logger, *bytes.Buffer) {
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

type testStruct struct {
	Password string `json:"password"`
	User     string `json:"user"`
}

type nestedStruct struct {
	APIKey string `json:"api_key"`
	Other  string `json:"other"`
}

type outerStruct struct {
	Nest nestedStruct `json:"nest"`
}

type circStruct struct {
	Name string      `json:"name"`
	Self *circStruct `json:"self"`
}

type baseStruct struct {
	Password string `json:"password"`
}

type derivedStruct struct {
	baseStruct

	Other string `json:"other"`
}

type unexportedNonSensitive struct {
	internal string
	User     string `json:"user"`
}

type structWithUnexportedMap struct {
	internal map[string]string
	Exported string `json:"exported"`
}

type circularMap map[string]interface{}

type structWithFunc struct {
	Handler func() `json:"handler"`
	Name    string `json:"name"`
}

type structWithChan struct {
	Events chan string `json:"events"`
	ID     int         `json:"id"`
}

type structWithComplex struct {
	Value complex128 `json:"value"`
	Name  string     `json:"name"`
}

type customMarshaler struct {
	Password string
	Value    string
}

func (c customMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"custom": c.Value})
}

type valueReceiverStringer struct {
	Password string
	Value    string
}

func (c valueReceiverStringer) String() string {
	return "stringer:" + c.Value
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

type structWithIgnoredField struct {
	Password string `json:"-"`
	User     string `json:"user"`
}

type structWithOmitempty struct {
	Password string `json:"password,omitempty"`
	User     string `json:"user,omitempty"`
	Empty    string `json:"empty,omitempty"`
}

type customTextMarshaler struct {
	Password string
	Value    string
}

func (c customTextMarshaler) MarshalText() ([]byte, error) {
	return []byte("text:" + c.Value), nil
}

type customError struct {
	Code    int
	Message string
}

func (e customError) Error() string {
	return e.Message
}

type pointerReceiverMarshaler struct {
	Password string
	Value    string
}

func (c *pointerReceiverMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"ptr_value": c.Value})
}

type pointerReceiverStringer struct {
	Password string
	Value    string
}

func (c *pointerReceiverStringer) String() string {
	return "ptr_stringer:" + c.Value
}

func TestSanitizer(t *testing.T) {
	tests := []struct {
		name string
		opts []pnpzapsanitize.Option
		log  func(*zap.Logger)
		want map[string]interface{}
	}{
		{
			name: "DefaultRedactSimpleFields",
			log: func(l *zap.Logger) {
				l.Info("test", zap.String("password", "secret"), zap.String("user", "ok"), zap.String("Password", "secret2"))
			},
			want: map[string]interface{}{
				"level":    "info",
				"msg":      "test",
				"password": "[REDACTED]",
				"Password": "[REDACTED]",
				"user":     "ok",
			},
		},
		{
			name: "RedactStruct",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Reflect("data", testStruct{Password: "secret", User: "ok"}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"user":     "ok",
				},
			},
		},
		{
			name: "RedactNestedStruct",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Reflect("data", outerStruct{
					Nest: nestedStruct{APIKey: "secret", Other: "ok"},
				}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"nest": map[string]interface{}{
						"api_key": "[REDACTED]",
						"other":   "ok",
					},
				},
			},
		},
		{
			name: "RedactMap",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Any("data", map[string]interface{}{
					"token":  "secret",
					"user":   "ok",
					"nested": map[string]string{"client_secret": "secret2"},
				}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"token":  "[REDACTED]",
					"user":   "ok",
					"nested": map[string]interface{}{"client_secret": "[REDACTED]"},
				},
			},
		},
		{
			name: "RedactSlice",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Any("data", []interface{}{"ok", map[string]string{"api_key": "secret"}}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					"ok",
					map[string]interface{}{"api_key": "[REDACTED]"},
				},
			},
		},
		{
			name: "RedactCircular",
			log: func(l *zap.Logger) {
				c := &circStruct{Name: "a"}
				c.Self = c
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"name": "a",
					"self": "[CIRCULAR_REFERENCE]",
				},
			},
		},
		{
			name: "RedactInlineStruct",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Reflect("data", derivedStruct{baseStruct: baseStruct{Password: "secret"}, Other: "ok"}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"other":    "ok",
				},
			},
		},
		{
			name: "RedactFieldInsideNamespace",
			log: func(l *zap.Logger) {
				nsLogger := l.With(zap.Namespace("user_data"))
				nsLogger.Info("test", zap.String("api_key", "should-be-hidden"))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"user_data": map[string]interface{}{
					"api_key": "[REDACTED]",
				},
			},
		},
		{
			name: "RedactNonStringField",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Int("client_id", 123), zap.Int("port", 8080))
			},
			want: map[string]interface{}{
				"level":     "info",
				"msg":       "test",
				"client_id": "[REDACTED]",
				"port":      float64(8080),
			},
		},
		{
			name: "CustomRegex",
			opts: []pnpzapsanitize.Option{pnpzapsanitize.WithRegex(regexp.MustCompile(`(?i)user`))},
			log: func(l *zap.Logger) {
				l.Info("test", zap.String("user", "ok"), zap.String("password", "secret"))
			},
			want: map[string]interface{}{
				"level":    "info",
				"msg":      "test",
				"user":     "[REDACTED]",
				"password": "secret",
			},
		},
		{
			name: "CustomRedacted",
			opts: []pnpzapsanitize.Option{pnpzapsanitize.WithRedacted("***")},
			log: func(l *zap.Logger) {
				l.Info("test", zap.String("password", "secret"))
			},
			want: map[string]interface{}{
				"level":    "info",
				"msg":      "test",
				"password": "***",
			},
		},
		{
			name: "WithContextualFields",
			log: func(l *zap.Logger) {
				ctxLogger := l.With(zap.String("token", "secret"), zap.String("info", "ok"))
				ctxLogger.Info("test")
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"token": "[REDACTED]",
				"info":  "ok",
			},
		},
		{
			name: "UnexportedNonSensitiveField",
			log: func(l *zap.Logger) {
				data := unexportedNonSensitive{internal: "unexported-data", User: "ok"}
				l.Info("test", zap.Reflect("data", data))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"user": "ok",
				},
			},
		},
		{
			name: "UnexportedMap",
			log: func(l *zap.Logger) {
				data := structWithUnexportedMap{
					internal: map[string]string{"password": "secret"},
					Exported: "exported-string",
				}
				l.Info("test", zap.Reflect("data", data))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"exported": "exported-string",
				},
			},
		},
		{
			name: "CircularMap",
			log: func(l *zap.Logger) {
				m := make(circularMap)
				m["self"] = m
				m["other"] = "ok"
				l.Info("test", zap.Any("data", m))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"self":  "[CIRCULAR_REFERENCE]",
					"other": "ok",
				},
			},
		},
		{
			name: "CircularSlice",
			log: func(l *zap.Logger) {
				var s []interface{}
				s = append(s, "ok", nil)
				s[1] = s
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					"ok",
					"[CIRCULAR_REFERENCE]",
				},
			},
		},
		{
			name: "NilInterface",
			log: func(l *zap.Logger) {
				var nilInterface interface{}
				l.Info("test", zap.Any("data", nilInterface))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  nil,
			},
		},
		{
			name: "InterfaceHoldingStruct",
			log: func(l *zap.Logger) {
				var iface interface{} = testStruct{Password: "secret", User: "alice"}
				l.Info("test", zap.Any("data", iface))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"user":     "alice",
				},
			},
		},
		{
			name: "InterfaceHoldingPointerToStruct",
			log: func(l *zap.Logger) {
				var iface interface{} = &testStruct{Password: "secret", User: "bob"}
				l.Info("test", zap.Any("data", iface))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"user":     "bob",
				},
			},
		},
		{
			name: "MapWithIntKeys",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Any("data", map[int]string{
					1: "one",
					2: "two",
				}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"1": "one",
					"2": "two",
				},
			},
		},
		{
			name: "MapWithBoolKeys",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Any("data", map[bool]string{
					true:  "yes",
					false: "no",
				}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"true":  "yes",
					"false": "no",
				},
			},
		},
		{
			name: "NilMap",
			log: func(l *zap.Logger) {
				var nilMap map[string]string
				l.Info("test", zap.Any("data", nilMap))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  map[string]interface{}{},
			},
		},
		{
			name: "EmptyMap",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Any("data", map[string]string{}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  map[string]interface{}{},
			},
		},
		{
			name: "NilSlice",
			log: func(l *zap.Logger) {
				var nilSlice []string
				l.Info("test", zap.Any("data", nilSlice))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  []interface{}{},
			},
		},
		{
			name: "EmptySlice",
			log: func(l *zap.Logger) {
				l.Info("test", zap.Any("data", []string{}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  []interface{}{},
			},
		},
		{
			name: "SliceOfPointers",
			log: func(l *zap.Logger) {
				s1 := &testStruct{Password: "secret1", User: "user1"}
				s2 := &testStruct{Password: "secret2", User: "user2"}
				var nilPtr *testStruct
				l.Info("test", zap.Any("data", []*testStruct{s1, nilPtr, s2}))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					map[string]interface{}{"password": "[REDACTED]", "user": "user1"},
					nil,
					map[string]interface{}{"password": "[REDACTED]", "user": "user2"},
				},
			},
		},
		{
			name: "NestedSlices",
			log: func(l *zap.Logger) {
				nested := [][]map[string]string{
					{{"api_key": "secret1", "name": "a"}},
					{{"password": "secret2", "value": "b"}, {"token": "secret3"}},
				}
				l.Info("test", zap.Any("data", nested))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					[]interface{}{map[string]interface{}{"api_key": "[REDACTED]", "name": "a"}},
					[]interface{}{
						map[string]interface{}{"password": "[REDACTED]", "value": "b"},
						map[string]interface{}{"token": "[REDACTED]"},
					},
				},
			},
		},
		{
			name: "FunctionType",
			log: func(l *zap.Logger) {
				fn := func(x int) int { return x * 2 }
				l.Info("test", zap.Any("data", fn))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "<func(int) int>",
			},
		},
		{
			name: "StructWithFunction",
			log: func(l *zap.Logger) {
				s := structWithFunc{
					Handler: func() {},
					Name:    "test-func",
				}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"handler": "<func()>",
					"name":    "test-func",
				},
			},
		},
		{
			name: "ChannelType",
			log: func(l *zap.Logger) {
				ch := make(chan int, 10)
				l.Info("test", zap.Any("data", ch))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "<chan int>",
			},
		},
		{
			name: "StructWithChannel",
			log: func(l *zap.Logger) {
				s := structWithChan{
					Events: make(chan string),
					ID:     42,
				}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"events": "<chan string>",
					"id":     float64(42),
				},
			},
		},
		{
			name: "NilPointer",
			log: func(l *zap.Logger) {
				var nilPtr *testStruct
				l.Info("test", zap.Any("data", nilPtr))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  nil,
			},
		},
		{
			name: "DoublePointer",
			log: func(l *zap.Logger) {
				s := &testStruct{Password: "secret", User: "user1"}
				pp := &s
				l.Info("test", zap.Any("data", pp))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"user":     "user1",
				},
			},
		},
		{
			name: "TriplePointer",
			log: func(l *zap.Logger) {
				s := &testStruct{Password: "secret", User: "user1"}
				pp := &s
				ppp := &pp
				l.Info("test", zap.Any("data", ppp))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"user":     "user1",
				},
			},
		},
		{
			name: "Complex64",
			log: func(l *zap.Logger) {
				c := complex64(3 + 4i)
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "3+4i",
			},
		},
		{
			name: "Complex128",
			log: func(l *zap.Logger) {
				c := complex128(1.5 + 2.5i)
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "1.5+2.5i",
			},
		},
		{
			name: "StructWithComplex",
			log: func(l *zap.Logger) {
				s := structWithComplex{Value: 1 + 2i, Name: "complex-test"}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"value": "(1+2i)",
					"name":  "complex-test",
				},
			},
		},
		{
			name: "Array",
			log: func(l *zap.Logger) {
				arr := [3]string{"a", "b", "c"}
				l.Info("test", zap.Any("data", arr))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  []interface{}{"a", "b", "c"},
			},
		},
		{
			name: "ArrayOfMaps",
			log: func(l *zap.Logger) {
				arr := [2]map[string]string{
					{"password": "secret1", "name": "a"},
					{"api_key": "secret2", "value": "b"},
				}
				l.Info("test", zap.Any("data", arr))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					map[string]interface{}{"password": "[REDACTED]", "name": "a"},
					map[string]interface{}{"api_key": "[REDACTED]", "value": "b"},
				},
			},
		},
		{
			name: "JSONMarshaler",
			log: func(l *zap.Logger) {
				c := customMarshaler{Password: "should-not-appear", Value: "custom-value"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"custom": "custom-value",
				},
			},
		},
		{
			name: "FmtStringer",
			log: func(l *zap.Logger) {
				c := valueReceiverStringer{Password: "should-not-appear", Value: "stringer-value"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "stringer:stringer-value",
			},
		},
		{
			name: "DeeplyNestedStruct",
			log: func(l *zap.Logger) {
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
				l.Info("test", zap.Any("data", data))
			},
			want: map[string]interface{}{
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
			},
		},
		{
			name: "DeeplyNestedMaps",
			log: func(l *zap.Logger) {
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
				l.Info("test", zap.Any("data", data))
			},
			want: map[string]interface{}{
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
			},
		},
		{
			name: "MixedTypes",
			log: func(l *zap.Logger) {
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
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"func": "<func()>",
					"chan": "<chan int>",
					"map":  map[string]interface{}{"password": "[REDACTED]", "name": "ok"},
					"slice": []interface{}{
						"a",
						"b",
					},
					"ptr": map[string]interface{}{"password": "[REDACTED]", "user": "alice"},
					"iface": map[string]interface{}{
						"token": "[REDACTED]",
					},
					"complex": "(1+2i)",
					"nested": map[string]interface{}{
						"api_key": "[REDACTED]",
						"values":  []interface{}{"v1"},
					},
				},
			},
		},
		{
			name: "StructWithJsonIgnore",
			log: func(l *zap.Logger) {
				s := structWithIgnoredField{Password: "secret", User: "alice"}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"user": "alice",
				},
			},
		},
		{
			name: "StructWithOmitempty",
			log: func(l *zap.Logger) {
				s := structWithOmitempty{Password: "secret", User: "alice", Empty: ""}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"password": "[REDACTED]",
					"user":     "alice",
					"empty":    "",
				},
			},
		},
		{
			name: "PrimitiveTypes",
			log: func(l *zap.Logger) {
				l.Info("test",
					zap.Any("int", 42),
					zap.Any("float", 3.14),
					zap.Any("bool", true),
					zap.Any("string", "hello"),
				)
			},
			want: map[string]interface{}{
				"level":  "info",
				"msg":    "test",
				"int":    float64(42),
				"float":  3.14,
				"bool":   true,
				"string": "hello",
			},
		},
		{
			name: "MapWithStructValues",
			log: func(l *zap.Logger) {
				data := map[string]testStruct{
					"user1": {Password: "secret1", User: "alice"},
					"user2": {Password: "secret2", User: "bob"},
				}
				l.Info("test", zap.Any("data", data))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"user1": map[string]interface{}{
						"password": "[REDACTED]",
						"user":     "alice",
					},
					"user2": map[string]interface{}{
						"password": "[REDACTED]",
						"user":     "bob",
					},
				},
			},
		},
		{
			name: "SliceOfMixedInterfaces",
			log: func(l *zap.Logger) {
				data := []interface{}{
					42,
					"string",
					true,
					map[string]string{"password": "secret"},
					testStruct{Password: "secret", User: "alice"},
					nil,
				}
				l.Info("test", zap.Any("data", data))
			},
			want: map[string]interface{}{
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
			},
		},
		{
			name: "ByteSlice",
			log: func(l *zap.Logger) {
				data := []byte(`{"user_id":18,"value":"test"}`)
				l.Info("test", zap.Any("data", data))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "eyJ1c2VyX2lkIjoxOCwidmFsdWUiOiJ0ZXN0In0=",
			},
		},
		{
			name: "ByteSliceInStruct",
			log: func(l *zap.Logger) {
				type messageStruct struct {
					ID   int    `json:"id"`
					Data []byte `json:"data"`
				}
				msg := messageStruct{ID: 1, Data: []byte("hello")}
				l.Info("test", zap.Any("message", msg))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"message": map[string]interface{}{
					"id":   float64(1),
					"data": "aGVsbG8=",
				},
			},
		},
		{
			name: "TextMarshaler",
			log: func(l *zap.Logger) {
				c := customTextMarshaler{Password: "should-not-appear", Value: "marshaled-value"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "text:marshaled-value",
			},
		},
		{
			name: "ErrorInterface",
			log: func(l *zap.Logger) {
				err := customError{Code: 500, Message: "something went wrong"}
				l.Info("test", zap.Any("error", err))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"error": "something went wrong",
			},
		},
		{
			name: "ErrorInStruct",
			log: func(l *zap.Logger) {
				type response struct {
					Success bool  `json:"success"`
					Error   error `json:"error"`
				}
				resp := response{Success: false, Error: customError{Code: 404, Message: "not found"}}
				l.Info("test", zap.Any("response", resp))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"response": map[string]interface{}{
					"success": false,
					"error":   "not found",
				},
			},
		},
		{
			name: "StringerInStruct",
			log: func(l *zap.Logger) {
				type wrapper struct {
					Value fmt.Stringer `json:"value"`
					Name  string       `json:"name"`
				}
				w := wrapper{Value: valueReceiverStringer{Password: "secret", Value: "stringer-val"}, Name: "test"}
				l.Info("test", zap.Any("data", w))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"value": "stringer:stringer-val",
					"name":  "test",
				},
			},
		},
		{
			name: "ErrorInMap",
			log: func(l *zap.Logger) {
				m := map[string]interface{}{
					"error": customError{Code: 500, Message: "internal error"},
					"other": "ok",
				}
				l.Info("test", zap.Any("data", m))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"error": "internal error",
					"other": "ok",
				},
			},
		},
		{
			name: "StringerInMap",
			log: func(l *zap.Logger) {
				m := map[string]interface{}{
					"stringer": valueReceiverStringer{Password: "secret", Value: "map-val"},
					"other":    "ok",
				}
				l.Info("test", zap.Any("data", m))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"stringer": "stringer:map-val",
					"other":    "ok",
				},
			},
		},
		{
			name: "StringerInMapPointer",
			log: func(l *zap.Logger) {
				m := map[string]interface{}{
					"stringer": pointerReceiverStringer{Password: "secret", Value: "map-val"},
					"other":    "ok",
				}
				l.Info("test", zap.Any("data", m))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"stringer": "ptr_stringer:map-val",
					"other":    "ok",
				},
			},
		},
		{
			name: "ErrorInSlice",
			log: func(l *zap.Logger) {
				s := []interface{}{
					"ok",
					customError{Code: 400, Message: "bad request"},
				}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					"ok",
					"bad request",
				},
			},
		},
		{
			name: "StringerInSlice",
			log: func(l *zap.Logger) {
				s := []interface{}{
					"ok",
					valueReceiverStringer{Password: "secret", Value: "slice-val"},
				}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					"ok",
					"stringer:slice-val",
				},
			},
		},
		{
			name: "PointerReceiverStringerInStruct",
			log: func(l *zap.Logger) {
				type wrapper struct {
					Value fmt.Stringer `json:"value"`
				}
				w := wrapper{Value: &pointerReceiverStringer{Password: "secret", Value: "ptr-stringer-val"}}
				l.Info("test", zap.Any("data", w))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"value": "ptr_stringer:ptr-stringer-val",
				},
			},
		},
		{
			name: "PointerReceiverMarshaler",
			log: func(l *zap.Logger) {
				c := pointerReceiverMarshaler{Password: "should-not-appear", Value: "pointer-value"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"ptr_value": "pointer-value",
				},
			},
		},
		{
			name: "PointerReceiverStringer",
			log: func(l *zap.Logger) {
				c := pointerReceiverStringer{Password: "should-not-appear", Value: "stringer-value"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "ptr_stringer:stringer-value",
			},
		},
		{
			name: "ValueReceiverMarshalerAsPointer",
			log: func(l *zap.Logger) {
				c := &customMarshaler{Password: "should-not-appear", Value: "value-as-ptr"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"custom": "value-as-ptr",
				},
			},
		},
		{
			name: "ValueReceiverMarshaler",
			log: func(l *zap.Logger) {
				c := customMarshaler{Password: "should-not-appear", Value: "value-as-ptr"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"custom": "value-as-ptr",
				},
			},
		},
		{
			name: "ValueReceiverStringerAsPointer",
			log: func(l *zap.Logger) {
				c := &valueReceiverStringer{Password: "should-not-appear", Value: "stringer-as-ptr"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "stringer:stringer-as-ptr",
			},
		},
		{
			name: "ValueReceiverStringer",
			log: func(l *zap.Logger) {
				c := valueReceiverStringer{Password: "should-not-appear", Value: "stringer-as-ptr"}
				l.Info("test", zap.Any("data", c))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data":  "stringer:stringer-as-ptr",
			},
		},
		{
			name: "ValueReceiverMarshalerInStruct",
			log: func(l *zap.Logger) {
				type wrapper struct {
					Value customMarshaler `json:"value"`
					Name  string          `json:"name"`
				}
				w := wrapper{Value: customMarshaler{Password: "secret", Value: "struct-val"}, Name: "test"}
				l.Info("test", zap.Any("data", w))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"value": map[string]interface{}{
						"custom": "struct-val",
					},
					"name": "test",
				},
			},
		},
		{
			name: "ValueReceiverMarshalerInStructAsPointer",
			log: func(l *zap.Logger) {
				type wrapper struct {
					Value *customMarshaler `json:"value"`
					Name  string           `json:"name"`
				}
				w := wrapper{Value: &customMarshaler{Password: "secret", Value: "struct-val"}, Name: "test"}
				l.Info("test", zap.Any("data", w))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"value": map[string]interface{}{
						"custom": "struct-val",
					},
					"name": "test",
				},
			},
		},
		{
			name: "ValueReceiverStringerInStruct",
			log: func(l *zap.Logger) {
				type wrapper struct {
					Value *valueReceiverStringer `json:"value"`
					Name  string                 `json:"name"`
				}
				w := wrapper{Value: &valueReceiverStringer{Password: "secret", Value: "stringer-struct-val"}, Name: "test"}
				l.Info("test", zap.Any("data", w))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"value": "stringer:stringer-struct-val",
					"name":  "test",
				},
			},
		},
		{
			name: "ValueReceiverMarshalerInMap",
			log: func(l *zap.Logger) {
				m := map[string]interface{}{
					"marshaler": &customMarshaler{Password: "secret", Value: "map-val"},
					"other":     "ok",
				}
				l.Info("test", zap.Any("data", m))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"marshaler": map[string]interface{}{
						"custom": "map-val",
					},
					"other": "ok",
				},
			},
		},
		{
			name: "ValueReceiverStringerInMap",
			log: func(l *zap.Logger) {
				m := map[string]interface{}{
					"stringer": &valueReceiverStringer{Password: "secret", Value: "map-stringer-val"},
					"other":    "ok",
				}
				l.Info("test", zap.Any("data", m))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": map[string]interface{}{
					"stringer": "stringer:map-stringer-val",
					"other":    "ok",
				},
			},
		},
		{
			name: "ValueReceiverMarshalerInSlice",
			log: func(l *zap.Logger) {
				s := []interface{}{
					"ok",
					&customMarshaler{Password: "secret", Value: "slice-val"},
				}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					"ok",
					map[string]interface{}{
						"custom": "slice-val",
					},
				},
			},
		},
		{
			name: "ValueReceiverStringerInSlice",
			log: func(l *zap.Logger) {
				s := []interface{}{
					"ok",
					&valueReceiverStringer{Password: "secret", Value: "slice-stringer-val"},
				}
				l.Info("test", zap.Any("data", s))
			},
			want: map[string]interface{}{
				"level": "info",
				"msg":   "test",
				"data": []interface{}{
					"ok",
					"stringer:slice-stringer-val",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := newSanitizedLogger(tt.opts...)
			tt.log(logger)
			require.NoError(t, logger.Sync())

			var res map[string]interface{}
			require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

			if tt.want != nil {
				require.Equal(t, tt.want, res)
			}
		})
	}
}
