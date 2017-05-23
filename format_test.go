package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/iafan/agenda"
)

func TestFormatNew(t *testing.T) {
	agenda.Run(t, "testdata/format/new", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Msg string `json:"msg"`
			Fmt string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		out := fmt.Sprintf(in.Fmt, New(in.Msg))
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

func TestFormatErrorf(t *testing.T) {
	agenda.Run(t, "testdata/format/errorf", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Msg string `json:"msg"`
			Fmt string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		out := fmt.Sprintf(in.Fmt, Errorf("%s", in.Msg))
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

func TestFormatWrap(t *testing.T) {
	agenda.Run(t, "testdata/format/wrap", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
			Fmt string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		var wrapped error
		for _, e := range in.Errors {
			if wrapped == nil {
				wrapped = io.EOF
				if e.Msg != "" {
					wrapped = New(e.Msg)
				}
			} else {
				wrapped = Wrap(wrapped, e.Msg)
			}
		}

		out := fmt.Sprintf(in.Fmt, wrapped)
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

func TestFormatWrapf(t *testing.T) {
	agenda.Run(t, "testdata/format/wrapf", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Errors []struct {
				Msg string `json:"msg"`
				Val int    `json:"val"`
			} `json:"errors"`
			Fmt string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		var wrapped error
		for _, e := range in.Errors {
			if wrapped == nil {
				wrapped = io.EOF
				if e.Msg != "" {
					wrapped = New(e.Msg)
				}
			} else {
				wrapped = Wrapf(wrapped, e.Msg, e.Val)
			}
		}

		out := fmt.Sprintf(in.Fmt, wrapped)
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

func TestFormatWithStack(t *testing.T) {
	agenda.Run(t, "testdata/format/withstack", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
			ExtraStacks int    `json:"extra_stacks"`
			Fmt         string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		var wrapped error
		for _, e := range in.Errors {
			if wrapped == nil {
				wrapped = io.EOF
				if e.Msg != "" {
					wrapped = New(e.Msg)
				}
			} else {
				wrapped = Wrap(wrapped, e.Msg)
			}
		}
		for i := 0; i < 1+in.ExtraStacks; i++ {
			wrapped = WithStack(wrapped)
		}

		out := fmt.Sprintf(in.Fmt, wrapped)
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

func TestFormatWithMessage(t *testing.T) {
	agenda.Run(t, "testdata/format/withmessage", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
			ExtraMsgs int    `json:"extra_msgs"`
			Fmt       string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		var wrapped error
		for _, e := range in.Errors {
			if wrapped == nil {
				wrapped = io.EOF
				if e.Msg != "" {
					wrapped = New(e.Msg)
				}
			} else {
				wrapped = Wrap(wrapped, e.Msg)
			}
		}
		for i := 0; i < 1+in.ExtraMsgs; i++ {
			wrapped = WithMessage(wrapped, fmt.Sprintf("addition%d", i))
		}

		out := fmt.Sprintf(in.Fmt, wrapped)
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

func TestFormatGeneric(t *testing.T) {
	agenda.Run(t, "testdata/format/generic", func(path string, data []byte) ([]byte, error) {
		in := struct {
			Errors []struct {
				Msg string `json:"msg"`
			} `json:"errors"`
			Fmt string `json:"format"`
		}{}

		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}

		var wrapped error
		for _, e := range in.Errors {
			if wrapped == nil {
				wrapped = io.EOF
				if e.Msg != "" {
					wrapped = New(e.Msg)
				}
			} else {
				wrapped = Wrap(wrapped, e.Msg)
			}
		}

		wrapped = WithMessage(wrapped, "with-message")
		wrapped = WithStack(wrapped)
		wrapped = Wrap(wrapped, "wrap-error")
		wrapped = Wrapf(wrapped, "wrapf-error%d", 1)

		out := fmt.Sprintf(in.Fmt, wrapped)
		return []byte(cleanupErrorStackTraces(out)), nil
	})
}

// Helper function to cleanup stack trace

var simpleMessageR = regexp.MustCompile(`^[\s\w\d\.\,:\-"']+$`)
var packageRelatedR = regexp.MustCompile(`(github\.com\/)\w+(\/errors[\.\/])`)
var stackCleanup1R = regexp.MustCompile(`(github\.com\/)\w+(\/errors\.)`)
var stackCleanup2R = regexp.MustCompile(`(?m)^(\s+).+?(\/github\.com\/)\w+(\/errors\/)`)
var stackCleanup3R = regexp.MustCompile(`(?m):\d+$`)

func cleanupErrorStackTraces(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if simpleMessageR.MatchString(line) {
			out = append(out, line)
			continue
		}

		if packageRelatedR.MatchString(line) {
			line = stackCleanup1R.ReplaceAllString(line, `$1<user>$2`)
			line = stackCleanup2R.ReplaceAllString(line, `$1<path>$2<user>$3`)
			line = stackCleanup3R.ReplaceAllString(line, `:<line>`)
			out = append(out, line)
			continue
		}
	}
	return strings.Join(out, "\n")
}
