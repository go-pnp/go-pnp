package pnpzapsanitize

import "regexp"

type options struct {
	regex    *regexp.Regexp
	redacted string
}

type Option func(*options)

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
