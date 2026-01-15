package pnpzapsanitize

import (
	"regexp"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

var defaultSensitiveKeysRegex = regexp.MustCompile(`(?i)password|api_?key|token|client_(id|secret|key)`)

const defaultRedactedValue = "[REDACTED]"
const circularRefPlaceholder = "[CIRCULAR_REFERENCE]"

type options struct {
	regex    *regexp.Regexp
	redacted string
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		regex:    defaultSensitiveKeysRegex,
		redacted: defaultRedactedValue,
	}, opts...)
}

type Option = optionutil.Option[options]

func WithRegex(re *regexp.Regexp) Option {
	return func(c *options) {
		if re != nil {
			c.regex = re
		}
	}
}

func WithRedacted(redacted string) Option {
	return func(c *options) {
		if redacted != "" {
			c.redacted = redacted
		}
	}
}
