package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

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

var (
	PrimaryStyle      = newStyle().Foreground(lipgloss.Color(ColorPrimary))
	AccentStyle       = newStyle().Foreground(lipgloss.Color(ColorAccent))
	ErrorStyle        = newStyle().Foreground(lipgloss.Color(ColorError))
	TextStyle         = newStyle().Foreground(lipgloss.Color(ColorText))
	TitleStyle        = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorAccent))
	SubtitleStyle     = newEmStyle().Foreground(lipgloss.Color(ColorPrimary))
	SuccessStyle      = newBoldStyle().Foreground(lipgloss.Color(ColorPrimary))
	WarningStyle      = newBoldStyle().Foreground(lipgloss.Color(ColorAccent))
	InfoStyle         = newStyle().Foreground(lipgloss.Color(ColorText))
	BoxStyle          = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPrimary))
	ErrorBoxStyle     = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorError))
	ListItemStyle     = newStyle().Foreground(lipgloss.Color(ColorText)).PaddingLeft(2)
	SelectedItemStyle = newBoldStyle().Foreground(lipgloss.Color(ColorAccent)).PaddingLeft(2)
	HeaderStyle       = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorPrimary))
	CellStyle         = newPStyle(0, 1).Foreground(lipgloss.Color(ColorText))
)

func newStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}

func newPStyle(v, h int) lipgloss.Style {
	return lipgloss.NewStyle().Padding(v, h)
}

func newBoldStyle() lipgloss.Style {
	return newStyle().Bold(true)
}

func newPBoldStyle(v, h int) lipgloss.Style {
	return newPStyle(v, h).Bold(true)
}

func newEmStyle() lipgloss.Style {
	return newStyle().Italic(true)
}

var logger *log.Logger

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

// success renders a success message
func success(msg string) string {
	return SuccessStyle.Render("✓ " + msg)
}

// error renders an error message
func errorMsg(msg string) string {
	return ErrorStyle.Render("✗ " + msg)
}

// warning renders a warning message
func warning(msg string) string {
	return WarningStyle.Render("⚠ " + msg)
}

// info renders an info message
func info(msg string) string {
	return InfoStyle.Render("ℹ " + msg)
}

// title renders a title
func title(msg string) string {
	return TitleStyle.Render(msg)
}

// subtitle renders a subtitle
func subtitle(msg string) string {
	return SubtitleStyle.Render(msg)
}

// box wraps content in a styled box
func box(content string) string {
	return BoxStyle.Render(content)
}

// errorBox wraps error content in a styled error box
func errorBox(content string) string {
	return ErrorBoxStyle.Render(content)
}

// Success prints a formatted success message
func Success(format string, a ...any) {
	fmt.Print(success(fmt.Sprintf(format, a...)))
}

// Successln prints a formatted success message with a newline
func Successln(format string, a ...any) {
	fmt.Println(success(fmt.Sprintf(format, a...)))
}

// Error prints a formatted error message
func Error(format string, a ...any) {
	fmt.Print(errorMsg(fmt.Sprintf(format, a...)))
}

// Errorln prints a formatted error message with a newline
func Errorln(format string, a ...any) {
	fmt.Println(errorMsg(fmt.Sprintf(format, a...)))
}

// Warning prints a formatted warning message
func Warning(format string, a ...any) {
	fmt.Print(warning(fmt.Sprintf(format, a...)))
}

// Warningln prints a formatted warning message with a newline
func Warningln(format string, a ...any) {
	fmt.Println(warning(fmt.Sprintf(format, a...)))
}

// Info prints a formatted info message
func Info(format string, a ...any) {
	fmt.Print(info(fmt.Sprintf(format, a...)))
}

// Infoln prints a formatted info message with a newline
func Infoln(format string, a ...any) {
	fmt.Println(info(fmt.Sprintf(format, a...)))
}

// Title prints a formatted title
func Title(format string, a ...any) {
	fmt.Print(title(fmt.Sprintf(format, a...)))
}

// Titleln prints a formatted title with a newline
func Titleln(format string, a ...any) {
	fmt.Println(title(fmt.Sprintf(format, a...)))
}

// Subtitle prints a formatted subtitle
func Subtitle(format string, a ...any) {
	fmt.Print(subtitle(fmt.Sprintf(format, a...)))
}

// Subtitleln prints a formatted subtitle with a newline
func Subtitleln(format string, a ...any) {
	fmt.Println(subtitle(fmt.Sprintf(format, a...)))
}

// Box prints content in a styled box
func Box(format string, a ...any) {
	fmt.Print(box(fmt.Sprintf(format, a...)))
}

// Boxln prints content in a styled box with a newline
func Boxln(format string, a ...any) {
	fmt.Println(box(fmt.Sprintf(format, a...)))
}

// ErrorBox prints error content in a styled error box
func ErrorBox(format string, a ...any) {
	fmt.Print(errorBox(fmt.Sprintf(format, a...)))
}

// ErrorBoxln prints error content in a styled error box with a newline
func ErrorBoxln(format string, a ...any) {
	fmt.Println(errorBox(fmt.Sprintf(format, a...)))
}
