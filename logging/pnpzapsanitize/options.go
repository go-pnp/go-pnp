package pnpzapsanitize

import (
	"regexp"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

var defaultSensitiveKeysRegex = regexp.MustCompile(`(?i)password|api_?key|token|client_(id|secret|key)`)

const defaultRedactedValue = "[REDACTED]"
const circularRefPlaceholder = "[CIRCULAR_REFERENCE]"

type Options struct {
	regex    *regexp.Regexp
	redacted string
}

func newOptions(opts []optionutil.Option[Options]) *Options {
	return optionutil.ApplyOptions(&Options{
		regex:    defaultSensitiveKeysRegex,
		redacted: defaultRedactedValue,
	}, opts...)
}

type Option = optionutil.Option[Options]

func WithRegex(re *regexp.Regexp) Option {
	return func(c *Options) {
		if re != nil {
			c.regex = re
		}
	}
}

func WithRedacted(redacted string) Option {
	return func(c *Options) {
		if redacted != "" {
			c.redacted = redacted
		}
	}
}
