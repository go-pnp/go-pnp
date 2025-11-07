package pnpzapsanitize_test

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/go-pnp/go-pnp/logging/pnpzapsanitize"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newSanitizedLogger(opts ...pnpzapsanitize.Option) (*zap.Logger, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.DebugLevel)

	return zap.New(core, pnpzapsanitize.Module(opts...)), buf
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

func TestSanitizer_RedactNestedStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", outerStruct{Nest: nestedStruct{APIKey: "secret", Other: "ok"}}))
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
	logger.Info("test", zap.Any("data", map[string]interface{}{"token": "secret", "user": "ok", "nested": map[string]string{"client_secret": "secret2"}}))
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

func TestSanitizer_UnexportedNonSensitiveField(t *testing.T) {
	logger, buf := newSanitizedLogger()
	data := unexportedNonSensitive{internal: "secret-data", User: "ok"}
	logger.Info("test", zap.Reflect("data", data))
	require.NoError(t, logger.Sync())

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &res))

	expected := map[string]interface{}{
		"level": "info",
		"msg":   "test",
		"data": map[string]interface{}{
			"internal": nil, // nil since unexported
			"user":     "ok",
		},
	}
	require.Equal(t, expected, res)
}
