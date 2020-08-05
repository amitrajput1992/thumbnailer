// +build !go1.13

package thumbnailer

import (
	"strings"
	"unicode/utf8"
)

// Convert string to valid UTF-8.
// Strings passed from C are not guaranteed to be valid.
func sanitize(s *string) {
	if !utf8.ValidString(*s) {
		*s = strings.Map(
			func(r rune) rune {
				if r == utf8.RuneError {
					// Need to replace invalid UTF-8 with a valid UTF-8 marker
					return '?'
				}
				return r
			},
			*s,
		)
	}
}
