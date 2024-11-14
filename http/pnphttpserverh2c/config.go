package pnphttpserverh2c

import "time"

// Config is a copy of the http2.Server fields
type Config struct {
	MaxHandlers int `env:"MAX_HANDLERS"`

	// MaxConcurrentStreams optionally specifies the number of
	// concurrent streams that each client may have open at a
	// time. This is unrelated to the number of http.Handler goroutines
	// which may be active globally, which is MaxHandlers.
	// If zero, MaxConcurrentStreams defaults to at least 100, per
	// the HTTP/2 spec's recommendations.
	MaxConcurrentStreams uint32 `env:"MAX_CONCURRENT_STREAMS"`

	// MaxDecoderHeaderTableSize optionally specifies the http2
	// SETTINGS_HEADER_TABLE_SIZE to send in the initial settings frame. It
	// informs the remote endpoint of the maximum size of the header compression
	// table used to decode header blocks, in octets. If zero, the default value
	// of 4096 is used.
	MaxDecoderHeaderTableSize uint32 `env:"MAX_DECODER_HEADER_TABLE_SIZE"`

	// MaxEncoderHeaderTableSize optionally specifies an upper limit for the
	// header compression table used for encoding request headers. Received
	// SETTINGS_HEADER_TABLE_SIZE settings are capped at this limit. If zero,
	// the default value of 4096 is used.
	MaxEncoderHeaderTableSize uint32 `env:"MAX_ENCODER_HEADER_TABLE_SIZE"`

	// MaxReadFrameSize optionally specifies the largest frame
	// this server is willing to read. A valid value is between
	// 16k and 16M, inclusive. If zero or otherwise invalid, a
	// default value is used.
	MaxReadFrameSize uint32 `env:"MAX_READ_FRAME_SIZE"`

	// PermitProhibitedCipherSuites, if true, permits the use of
	// cipher suites prohibited by the HTTP/2 spec.
	PermitProhibitedCipherSuites bool `env:"PERMIT_PROHIBITED_CIPHER_SUITES"`

	// IdleTimeout specifies how long until idle clients should be
	// closed with a GOAWAY frame. PING frames are not considered
	// activity for the purposes of IdleTimeout.
	// If zero or negative, there is no timeout.
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT"`

	// ReadIdleTimeout is the timeout after which a health check using a ping
	// frame will be carried out if no frame is received on the connection.
	// If zero, no health check is performed.
	ReadIdleTimeout time.Duration `env:"READ_IDLE_TIMEOUT"`

	// PingTimeout is the timeout after which the connection will be closed
	// if a response to a ping is not received.
	// If zero, a default of 15 seconds is used.
	PingTimeout time.Duration `env:"PING_TIMEOUT"`

	// WriteByteTimeout is the timeout after which a connection will be
	// closed if no data can be written to it. The timeout begins when data is
	// available to write, and is extended whenever any bytes are written.
	// If zero or negative, there is no timeout.
	WriteByteTimeout time.Duration `env:"WRITE_BYTE_TIMEOUT"`

	// MaxUploadBufferPerConnection is the size of the initial flow
	// control window for each connections. The HTTP/2 spec does not
	// allow this to be smaller than 65535 or larger than 2^32-1.
	// If the value is outside this range, a default value will be
	// used instead.
	MaxUploadBufferPerConnection int32 `env:"MAX_UPLOAD_BUFFER_PER_CONNECTION"`

	// MaxUploadBufferPerStream is the size of the initial flow control
	// window for each stream. The HTTP/2 spec does not allow this to
	// be larger than 2^32-1. If the value is zero or larger than the
	// maximum, a default value will be used instead.
	MaxUploadBufferPerStream int32 `env:"MAX_UPLOAD_BUFFER_PER_STREAM"`
}
