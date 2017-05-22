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

// Test New()

type TestFmtNew struct {
	input struct {
		Msg string `json:"msg"`
		Fmt string `json:"format"`
	}
	output string
}

func (t *TestFmtNew) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtNew) Run() error {
	t.output = fmt.Sprintf(t.input.Fmt, New(t.input.Msg))
	return nil
}

func (t *TestFmtNew) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatNew(t *testing.T) {
	agenda.Run(t, "testdata/format/new", &TestFmtNew{})
}

// Test Errorf()

type TestFmtErrorf struct {
	input struct {
		Msg string `json:"msg"`
		Fmt string `json:"format"`
	}
	output string
}

func (t *TestFmtErrorf) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtErrorf) Run() error {
	t.output = fmt.Sprintf(t.input.Fmt, Errorf("%s", t.input.Msg))
	return nil
}

func (t *TestFmtErrorf) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatErrorf(t *testing.T) {
	agenda.Run(t, "testdata/format/errorf", &TestFmtErrorf{})
}

// Test Wrap()

type TestFmtWrap struct {
	input struct {
		Errors []struct {
			Msg string `json:"msg"`
		} `json:"errors"`
		Fmt string `json:"format"`
	}
	output string
}

func (t *TestFmtWrap) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtWrap) Run() error {
	var wrapped error
	for _, e := range t.input.Errors {
		if wrapped == nil {
			wrapped = io.EOF
			if e.Msg != "" {
				wrapped = New(e.Msg)
			}
		} else {
			wrapped = Wrap(wrapped, e.Msg)
		}
	}

	t.output = fmt.Sprintf(t.input.Fmt, wrapped)
	return nil
}

func (t *TestFmtWrap) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatWrap(t *testing.T) {
	agenda.Run(t, "testdata/format/wrap", &TestFmtWrap{})
}

// Test Wrapf()

type TestFmtWrapf struct {
	input struct {
		Errors []struct {
			Msg string `json:"msg"`
			Val int    `json:"val"`
		} `json:"errors"`
		Fmt string `json:"format"`
	}
	output string
}

func (t *TestFmtWrapf) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtWrapf) Run() error {
	var wrapped error
	for _, e := range t.input.Errors {
		if wrapped == nil {
			wrapped = io.EOF
			if e.Msg != "" {
				wrapped = New(e.Msg)
			}
		} else {
			wrapped = Wrapf(wrapped, e.Msg, e.Val)
		}
	}

	t.output = fmt.Sprintf(t.input.Fmt, wrapped)
	return nil
}

func (t *TestFmtWrapf) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatWrapf(t *testing.T) {
	agenda.Run(t, "testdata/format/wrapf", &TestFmtWrapf{})
}

// Test WithStack()

type TestFmtWithStack struct {
	input struct {
		Errors []struct {
			Msg string `json:"msg"`
		} `json:"errors"`
		ExtraStacks int    `json:"extra_stacks"`
		Fmt         string `json:"format"`
	}
	output string
}

func (t *TestFmtWithStack) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtWithStack) Run() error {
	var wrapped error
	for _, e := range t.input.Errors {
		if wrapped == nil {
			wrapped = io.EOF
			if e.Msg != "" {
				wrapped = New(e.Msg)
			}
		} else {
			wrapped = Wrap(wrapped, e.Msg)
		}
	}
	for i := 0; i < 1+t.input.ExtraStacks; i++ {
		wrapped = WithStack(wrapped)
	}

	t.output = fmt.Sprintf(t.input.Fmt, wrapped)
	return nil
}

func (t *TestFmtWithStack) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatWithStack(t *testing.T) {
	agenda.Run(t, "testdata/format/withstack", &TestFmtWithStack{})
}

// Test WithMessage()

type TestFmtWithMessage struct {
	input struct {
		Errors []struct {
			Msg string `json:"msg"`
		} `json:"errors"`
		ExtraMsgs int    `json:"extra_msgs"`
		Fmt       string `json:"format"`
	}
	output string
}

func (t *TestFmtWithMessage) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtWithMessage) Run() error {
	var wrapped error
	for _, e := range t.input.Errors {
		if wrapped == nil {
			wrapped = io.EOF
			if e.Msg != "" {
				wrapped = New(e.Msg)
			}
		} else {
			wrapped = Wrap(wrapped, e.Msg)
		}
	}
	for i := 0; i < 1+t.input.ExtraMsgs; i++ {
		wrapped = WithMessage(wrapped, fmt.Sprintf("addition%d", i))
	}

	t.output = fmt.Sprintf(t.input.Fmt, wrapped)
	return nil
}

func (t *TestFmtWithMessage) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatWithMessage(t *testing.T) {
	agenda.Run(t, "testdata/format/withmessage", &TestFmtWithMessage{})
}

// Test multiple wrappers at once

type TestFmtGeneric struct {
	input struct {
		Errors []struct {
			Msg string `json:"msg"`
		} `json:"errors"`
		Fmt string `json:"format"`
	}
	output string
}

func (t *TestFmtGeneric) UnmarshalInput(data []byte) error {
	return json.Unmarshal(data, &t.input)
}

func (t *TestFmtGeneric) Run() error {
	var wrapped error
	for _, e := range t.input.Errors {
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

	t.output = fmt.Sprintf(t.input.Fmt, wrapped)
	return nil
}

func (t *TestFmtGeneric) MarshalOutput() ([]byte, error) {
	return []byte(cleanupErrorStackTraces(t.output)), nil
}

func TestFormatGeneric(t *testing.T) {
	agenda.Run(t, "testdata/format/generic", &TestFmtGeneric{})
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
