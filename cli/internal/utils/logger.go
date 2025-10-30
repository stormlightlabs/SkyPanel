package utils

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var logger *log.Logger

const (
	// Blue
	ColorPrimary = "#31748f"
	// Yellow/Gold
	ColorAccent = "#f6c177"
	// Red/Pink
	ColorError = "#eb6f92"
	// Light text
	ColorText = "#e0def4"
	// Dark background
	ColorBG = "#191724"
)

// InitLogger initializes the global logger with structured logging
func InitLogger(level log.Level) *log.Logger {
	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
		Level:           level,
	})

	styles := log.DefaultStyles()

	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().SetString("DEBUG").Foreground(lipgloss.Color(ColorText))
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().SetString("INFO").Foreground(lipgloss.Color(ColorPrimary))
	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().SetString("WARN").Foreground(lipgloss.Color(ColorAccent))
	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().SetString("ERROR").Foreground(lipgloss.Color(ColorError))
	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().SetString("FATAL").Foreground(lipgloss.Color(ColorError)).Bold(true)

	logger.SetStyles(styles)

	return logger
}

// GetLogger returns the global logger instance
func GetLogger() *log.Logger {
	if logger == nil {
		return InitLogger(log.InfoLevel)
	}
	return logger
}
