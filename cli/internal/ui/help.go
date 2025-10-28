package ui

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
)

// styleSectionHeader applies accent color to section headers (NAME:, USAGE:, etc.)
func styleSectionHeader(s string) string {
	if strings.HasSuffix(s, ":") {
		return AccentStyle.Bold(true).Render(s)
	}
	return s
}

// styleCommandName applies primary color to command names
func styleCommandName(s string) string {
	return PrimaryStyle.Render(s)
}

// styleDim applies dim styling to descriptions
func styleDim(s string) string {
	return lipgloss.NewStyle().Faint(true).Render(s)
}

// StyledHelpPrinter is a custom help printer that uses our color system
// It extends the default printer with custom styling functions
func StyledHelpPrinter(w io.Writer, templ string, data any) {
	customFuncs := map[string]any{
		"styleHeader":  styleSectionHeader,
		"styleCommand": styleCommandName,
		"styleDim":     styleDim,
		"styleAccent":  func(s string) string { return AccentStyle.Render(s) },
		"stylePrimary": func(s string) string { return PrimaryStyle.Render(s) },
	}

	cli.DefaultPrintHelpCustom(w, templ, data, customFuncs)
}

// RootCommandHelpTemplate uses the default template structure with styling functions
const RootCommandHelpTemplate = `{{styleHeader "NAME:"}}
   {{template "helpNameTemplate" .}}

{{styleHeader "USAGE:"}}
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[global options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

{{styleHeader "VERSION:"}}
   {{stylePrimary .Version}}{{end}}{{end}}{{if .Description}}

{{styleHeader "DESCRIPTION:"}}
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

{{styleHeader "AUTHOR"}}{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

{{styleHeader "COMMANDS:"}}{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

{{styleHeader "GLOBAL OPTIONS:"}}{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

{{styleHeader "GLOBAL OPTIONS:"}}{{template "visibleFlagTemplate" .}}{{end}}{{if .Copyright}}

{{styleHeader "COPYRIGHT:"}}
   {{template "copyrightTemplate" .}}{{end}}
`

// CommandHelpTemplate uses the default template structure with styling
const CommandHelpTemplate = `{{styleHeader "NAME:"}}
   {{template "helpNameTemplate" .}}

{{styleHeader "USAGE:"}}
   {{template "usageTemplate" .}}{{if .Category}}

{{styleHeader "CATEGORY:"}}
   {{.Category}}{{end}}{{if .Description}}

{{styleHeader "DESCRIPTION:"}}
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

{{styleHeader "OPTIONS:"}}{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

{{styleHeader "OPTIONS:"}}{{template "visibleFlagTemplate" .}}{{end}}
`

// SubcommandHelpTemplate uses the default template structure with styling
const SubcommandHelpTemplate = `{{styleHeader "NAME:"}}
   {{template "helpNameTemplate" .}}

{{styleHeader "USAGE:"}}
   {{template "usageTemplate" .}}{{if .Category}}

{{styleHeader "CATEGORY:"}}
   {{.Category}}{{end}}{{if .Description}}

{{styleHeader "DESCRIPTION:"}}
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

{{styleHeader "COMMANDS:"}}{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

{{styleHeader "OPTIONS:"}}{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

{{styleHeader "OPTIONS:"}}{{template "visibleFlagTemplate" .}}{{end}}
`
