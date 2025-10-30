package utils

import (
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestColorConstants(t *testing.T) {
	colors := map[string]string{
		"ColorPrimary": ColorPrimary,
		"ColorAccent":  ColorAccent,
		"ColorError":   ColorError,
		"ColorText":    ColorText,
		"ColorBG":      ColorBG,
	}

	for name, color := range colors {
		t.Run(name, func(t *testing.T) {
			if color == "" {
				t.Errorf("%s should not be empty", name)
			}
			if !strings.HasPrefix(color, "#") {
				t.Errorf("%s should be a hex color, got: %s", name, color)
			}
			if len(color) != 7 {
				t.Errorf("%s should be 7 characters (#RRGGBB), got: %s", name, color)
			}
		})
	}
}

func TestLogger(t *testing.T) {
	t.Run("InitLogger", func(t *testing.T) {
		originalLogger := logger
		defer func() {
			logger = originalLogger
		}()

		logger = nil

		l := InitLogger(log.DebugLevel)
		if l == nil {
			t.Fatal("InitLogger() should return a logger")
		}
		if logger == nil {
			t.Error("InitLogger() should set global logger")
		}
	})

	t.Run("GetLogger", func(t *testing.T) {
		originalLogger := logger
		defer func() {
			logger = originalLogger
		}()

		logger = nil

		l1 := GetLogger()
		if l1 == nil {
			t.Fatal("GetLogger() should return a logger")
		}

		l2 := GetLogger()
		if l2 == nil {
			t.Fatal("GetLogger() should return a logger on second call")
		}
		if l1 != l2 {
			t.Error("GetLogger() should return the same logger instance")
		}
	})

	t.Run("GetLoggerWithExistingLogger", func(t *testing.T) {
		originalLogger := logger
		defer func() {
			logger = originalLogger
		}()

		customLogger := InitLogger(log.WarnLevel)

		retrieved := GetLogger()
		if retrieved != customLogger {
			t.Error("GetLogger() should return existing logger")
		}
	})
}
