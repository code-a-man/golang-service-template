package logfx

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
)

type mockErr struct {
	msg   string
	stack []uintptr
}

func (m *mockErr) StackTrace() []uintptr {
	return m.stack
}

func (m *mockErr) Error() string {
	return m.msg
}

func (m *mockErr) Add(ptr uintptr) *mockErr {
	m.stack = append(m.stack, ptr)
	return m
}

func TestReplacerGenerator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		prettyMode bool
		groups     []string
		attr       slog.Attr
		expected   slog.Attr
	}{
		{
			name:       "PrettyMode=true, Key=slog.TimeKey",
			prettyMode: true,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.TimeKey},
			expected:   slog.Attr{},
		},
		{
			name:       "PrettyMode=true, Key=slog.LevelKey",
			prettyMode: true,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.LevelKey},
			expected:   slog.Attr{},
		},
		{
			name:       "PrettyMode=true, Key=slog.MessageKey",
			prettyMode: true,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.MessageKey},
			expected:   slog.Attr{},
		},
		{
			name:       "PrettyMode=false, Key=slog.TimeKey",
			prettyMode: false,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.TimeKey},
			expected:   slog.Attr{Key: slog.TimeKey},
		},
		{
			name:       "PrettyMode=false, Key=slog.LevelKey",
			prettyMode: false,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.LevelKey},
			expected:   slog.Attr{Key: slog.LevelKey},
		},
		{
			name:       "PrettyMode=false, Key=slog.MessageKey",
			prettyMode: false,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.MessageKey},
			expected:   slog.Attr{Key: slog.MessageKey},
		},
		{
			name:       "PrettyMode=false, Key=slog.TimeKey, Value=error",
			prettyMode: false,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.TimeKey, Value: slog.AnyValue(errors.New("test error"))},
			expected:   slog.Attr{Key: slog.TimeKey, Value: slog.GroupValue(slog.String("msg", "test error"))},
		},
		{
			name:       "PrettyMode=false, Key=slog.TimeKey, Value=error with StackTracer",
			prettyMode: false,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.TimeKey, Value: slog.AnyValue(errors.New("test error"))},
			expected:   slog.Attr{Key: slog.TimeKey, Value: slog.GroupValue(slog.String("msg", "test error"))},
		},
		{
			name:       "PrettyMode=false, Key=slog.TimeKey, Value=error with mockErr",
			prettyMode: false,
			groups:     []string{},
			attr:       slog.Attr{Key: slog.TimeKey, Value: slog.AnyValue(&mockErr{msg: "test error"})},
			expected:   slog.Attr{Key: slog.TimeKey, Value: slog.GroupValue(slog.String("msg", "test error"), slog.Any("trace", []string{}))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacerGenerator(tt.prettyMode)(tt.groups, tt.attr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTraceLines(t *testing.T) {
	t.Parallel()
	stackGenerator := func() []uintptr {
		var pc [32]uintptr
		n := runtime.Callers(0, pc[:])
		return pc[:n]
	}

	stackGeneratorWithUnknownFunctionAddr := func() []uintptr {
		var pc [32]uintptr
		n := runtime.Callers(0, pc[:])
		pc[0] = 0
		return pc[:n]
	}

	pwd := func() string {
		_, file, _, _ := runtime.Caller(0)
		return file
	}

	tests := []struct {
		name     string
		stack    []uintptr
		expected []string
	}{
		{
			name:     "Empty Stack",
			stack:    []uintptr{},
			expected: []string{},
		},
		{
			name:  "Non-Empty Stack",
			stack: stackGenerator(),
			expected: []string{
				"runtime.Callers /usr/local/go/src/runtime/extern.go:331",
				fmt.Sprint("github.com/eser/go-service/pkg/bliss/logfx.TestTraceLines.func1 ", pwd(), ":117"),
				fmt.Sprint("github.com/eser/go-service/pkg/bliss/logfx.TestTraceLines ", pwd(), ":145"),
				"testing.tRunner /usr/local/go/src/testing/testing.go:1690",
			},
		},
		{
			name:  "Non-Empty Stack with Unknown Function Address",
			stack: stackGeneratorWithUnknownFunctionAddr(),
			expected: []string{
				"unknown",
				fmt.Sprint("github.com/eser/go-service/pkg/bliss/logfx.TestTraceLines ", pwd(), ":147"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := traceLines(tt.stack)
			if len(tt.expected) == 0 {
				assert.Equal(t, tt.expected, result)
			} else {
				for i := range tt.expected {
					ext := strings.Split(tt.expected[i], " ")
					assert.Contains(t, result[i], ext[0])
				}
			}
		})
	}
}
