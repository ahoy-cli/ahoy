package main

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestSanitiseLogValue(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"leaves ordinary text alone", "/path/to/.ahoy.yml", "/path/to/.ahoy.yml"},
		{"escapes a line feed", "before\nafter", `before\nafter`},
		{"escapes a carriage return", "before\rafter", `before\rafter`},
		{"collapses CRLF to one escape", "before\r\nafter", `before\nafter`},
		{"escapes repeated breaks", "a\n\nb", `a\n\nb`},
		{"handles an empty string", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitiseLogValue(tc.in); got != tc.want {
				t.Errorf("sanitiseLogValue(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

// captureLog runs fn with the standard logger redirected, returning its output.
func captureLog(fn func()) string {
	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(original)
		log.SetFlags(0)
	}()
	fn()
	return buf.String()
}

// A config-derived value must not be able to forge an extra log line carrying a
// different level prefix.
func TestLoggerValuesCannotForgeLogLines(t *testing.T) {
	s := newAppState()
	hostile := "missing.yml\n[error] forged failure"

	out := captureLog(func() {
		s.logger(logLevelWarn, "Circular import detected for '"+sanitiseLogValue(hostile)+"', skipping.")
	})

	if got := strings.Count(strings.TrimRight(out, "\n"), "\n"); got != 0 {
		t.Errorf("expected a single log line, got %d extra line(s): %q", got, out)
	}
	if strings.Contains(out, "\n[error]") {
		t.Errorf("forged log line was not neutralised: %q", out)
	}
	if !strings.HasPrefix(out, "[warn] ") {
		t.Errorf("expected the warn prefix to be preserved, got %q", out)
	}
}

// Sanitisation is applied to interpolated values, not to the assembled message,
// so messages that deliberately span several lines keep their formatting.
func TestLoggerPreservesIntentionalMultilineMessages(t *testing.T) {
	s := newAppState()
	msg := "Command [foo] has 'imports' set, but no commands were found." +
		"\n\nSolutions:" +
		"\n1. Create the missing files"

	out := captureLog(func() {
		s.logger(logLevelError, msg)
	})

	if strings.Contains(out, `\n`) {
		t.Errorf("intentional line breaks were escaped: %q", out)
	}
	if !strings.Contains(out, "\n\nSolutions:\n1. Create the missing files") {
		t.Errorf("multi-line formatting was not preserved: %q", out)
	}
}
