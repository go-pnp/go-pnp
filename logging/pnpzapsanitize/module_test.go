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

func TestSanitizer_DefaultRedactSimpleFields(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.String("password", "secret"), zap.String("user", "ok"), zap.String("Password", "secret2"))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	require.Equal(t, "[REDACTED]", m["password"])
	require.Equal(t, "[REDACTED]", m["Password"])
	require.Equal(t, "ok", m["user"])
}

func TestSanitizer_RedactStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", testStruct{Password: "secret", User: "ok"}))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	data := m["data"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", data["password"])
	require.Equal(t, "ok", data["user"])
}

func TestSanitizer_RedactNestedStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", outerStruct{Nest: nestedStruct{APIKey: "secret", Other: "ok"}}))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	data := m["data"].(map[string]interface{})
	nest := data["nest"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", nest["api_key"])
	require.Equal(t, "ok", nest["other"])
}

func TestSanitizer_RedactMap(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", map[string]interface{}{"token": "secret", "user": "ok", "nested": map[string]string{"client_secret": "secret2"}}))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	data := m["data"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", data["token"])
	require.Equal(t, "ok", data["user"])
	nested := data["nested"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", nested["client_secret"])
}

func TestSanitizer_RedactSlice(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", []interface{}{"ok", map[string]string{"api_key": "secret"}}))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	data := m["data"].([]interface{})
	require.Equal(t, "ok", data[0])
	item := data[1].(map[string]interface{})
	require.Equal(t, "[REDACTED]", item["api_key"])
}

func TestSanitizer_RedactCircular(t *testing.T) {
	c := &circStruct{Name: "a"}
	c.Self = c
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Any("data", c))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	data := m["data"].(map[string]interface{})
	require.Equal(t, "a", data["name"])
	require.Equal(t, "[CIRCULAR_REFERENCE]", data["self"])
}

func TestSanitizer_RedactInlineStruct(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Reflect("data", derivedStruct{baseStruct: baseStruct{Password: "secret"}, Other: "ok"}))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	data := m["data"].(map[string]interface{})
	require.Equal(t, "[REDACTED]", data["password"])
	require.Equal(t, "ok", data["other"])
}

func TestSanitizer_RedactFieldInsideNamespace(t *testing.T) {
	logger, buf := newSanitizedLogger()
	nsLogger := logger.With(zap.Namespace("user_data"))

	nsLogger.Info("test", zap.String("api_key", "should-be-hidden"))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])

	require.Contains(t, m, "user_data")
	ns := m["user_data"].(map[string]interface{})

	require.Equal(t, "[REDACTED]", ns["api_key"])
}

func TestSanitizer_RedactNonStringField(t *testing.T) {
	logger, buf := newSanitizedLogger()
	logger.Info("test", zap.Int("client_id", 123), zap.Int("port", 8080))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	require.Equal(t, "[REDACTED]", m["client_id"])
	require.Equal(t, 8080, int(m["port"].(float64)))
}

func TestSanitizer_CustomRegex(t *testing.T) {
	re := regexp.MustCompile(`(?i)user`)
	logger, buf := newSanitizedLogger(pnpzapsanitize.WithRegex(re))
	logger.Info("test", zap.String("user", "ok"), zap.String("password", "secret"))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	require.Equal(t, "[REDACTED]", m["user"])
	require.Equal(t, "secret", m["password"])
}

func TestSanitizer_CustomRedacted(t *testing.T) {
	logger, buf := newSanitizedLogger(pnpzapsanitize.WithRedacted("***"))
	logger.Info("test", zap.String("password", "secret"))
	require.NoError(t, logger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	require.Equal(t, "***", m["password"])
}

func TestSanitizer_WithContextualFields(t *testing.T) {
	logger, buf := newSanitizedLogger()
	ctxLogger := logger.With(zap.String("token", "secret"), zap.String("info", "ok"))
	ctxLogger.Info("test")
	require.NoError(t, ctxLogger.Sync())

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	require.Equal(t, "test", m["msg"])
	require.Equal(t, "[REDACTED]", m["token"])
	require.Equal(t, "ok", m["info"])
}
