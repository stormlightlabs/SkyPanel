package ui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

func TestSuccess(t *testing.T) {
	result := success("test message")
	if !strings.Contains(result, "✓") {
		t.Errorf("success() should contain checkmark, got: %s", result)
	}
	if !strings.Contains(result, "test message") {
		t.Errorf("success() should contain message, got: %s", result)
	}
}

func TestErrorMsg(t *testing.T) {
	result := errorMsg("test error")
	if !strings.Contains(result, "✗") {
		t.Errorf("errorMsg() should contain X mark, got: %s", result)
	}
	if !strings.Contains(result, "test error") {
		t.Errorf("errorMsg() should contain message, got: %s", result)
	}
}

func TestWarning(t *testing.T) {
	result := warning("test warning")
	if !strings.Contains(result, "⚠") {
		t.Errorf("warning() should contain warning symbol, got: %s", result)
	}
	if !strings.Contains(result, "test warning") {
		t.Errorf("warning() should contain message, got: %s", result)
	}
}

func TestInfo(t *testing.T) {
	result := info("test info")
	if !strings.Contains(result, "ℹ") {
		t.Errorf("info() should contain info symbol, got: %s", result)
	}
	if !strings.Contains(result, "test info") {
		t.Errorf("info() should contain message, got: %s", result)
	}
}

func TestTitle(t *testing.T) {
	result := title("test title")
	if !strings.Contains(result, "test title") {
		t.Errorf("title() should contain message, got: %s", result)
	}
}

func TestSubtitle(t *testing.T) {
	result := subtitle("test subtitle")
	if !strings.Contains(result, "test subtitle") {
		t.Errorf("subtitle() should contain message, got: %s", result)
	}
}

func TestBox(t *testing.T) {
	result := box("test content")
	if !strings.Contains(result, "test content") {
		t.Errorf("box() should contain message, got: %s", result)
	}
}

func TestErrorBox(t *testing.T) {
	result := errorBox("test error content")
	if !strings.Contains(result, "test error content") {
		t.Errorf("errorBox() should contain message, got: %s", result)
	}
}

func TestUnexportedFunctionsWithEmptyString(t *testing.T) {
	tests := []struct {
		name           string
		fn             func(string) string
		expectNonEmpty bool
	}{
		{"success", success, true},    // Has icon prefix
		{"errorMsg", errorMsg, true},  // Has icon prefix
		{"warning", warning, true},    // Has icon prefix
		{"info", info, true},          // Has icon prefix
		{"title", title, false},       // No icon, may return empty
		{"subtitle", subtitle, false}, // No icon, may return empty
		{"box", box, true},            // Has border
		{"errorBox", errorBox, true},  // Has border
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("")
			if tt.expectNonEmpty && result == "" {
				t.Errorf("%s(\"\") returned empty string, expected styled content", tt.name)
			}
		})
	}
}

func TestSuccessPrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Success("test %s", "message")
	})
	if !strings.Contains(output, "test message") {
		t.Errorf("Success() should print message, got: %s", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("Success() should print checkmark, got: %s", output)
	}
}

func TestSuccessln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Successln("test %s", "message")
	})
	if !strings.Contains(output, "test message") {
		t.Errorf("Successln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Successln() should end with newline, got: %s", output)
	}
}

func TestErrorPrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Error("test %s", "error")
	})
	if !strings.Contains(output, "test error") {
		t.Errorf("Error() should print message, got: %s", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("Error() should print X mark, got: %s", output)
	}
}

func TestErrorln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Errorln("test %s", "error")
	})
	if !strings.Contains(output, "test error") {
		t.Errorf("Errorln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Errorln() should end with newline, got: %s", output)
	}
}

func TestWarningPrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Warning("test %s", "warning")
	})
	if !strings.Contains(output, "test warning") {
		t.Errorf("Warning() should print message, got: %s", output)
	}
	if !strings.Contains(output, "⚠") {
		t.Errorf("Warning() should print warning symbol, got: %s", output)
	}
}

func TestWarningln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Warningln("test %s", "warning")
	})
	if !strings.Contains(output, "test warning") {
		t.Errorf("Warningln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Warningln() should end with newline, got: %s", output)
	}
}

func TestInfoPrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Info("test %s", "info")
	})
	if !strings.Contains(output, "test info") {
		t.Errorf("Info() should print message, got: %s", output)
	}
	if !strings.Contains(output, "ℹ") {
		t.Errorf("Info() should print info symbol, got: %s", output)
	}
}

func TestInfoln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Infoln("test %s", "info")
	})
	if !strings.Contains(output, "test info") {
		t.Errorf("Infoln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Infoln() should end with newline, got: %s", output)
	}
}

func TestTitlePrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Title("test %s", "title")
	})
	if !strings.Contains(output, "test title") {
		t.Errorf("Title() should print message, got: %s", output)
	}
}

func TestTitleln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Titleln("test %s", "title")
	})
	if !strings.Contains(output, "test title") {
		t.Errorf("Titleln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Titleln() should end with newline, got: %s", output)
	}
}

func TestSubtitlePrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Subtitle("test %s", "subtitle")
	})
	if !strings.Contains(output, "test subtitle") {
		t.Errorf("Subtitle() should print message, got: %s", output)
	}
}

func TestSubtitleln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Subtitleln("test %s", "subtitle")
	})
	if !strings.Contains(output, "test subtitle") {
		t.Errorf("Subtitleln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Subtitleln() should end with newline, got: %s", output)
	}
}

func TestBoxPrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Box("test %s", "content")
	})
	if !strings.Contains(output, "test content") {
		t.Errorf("Box() should print message, got: %s", output)
	}
}

func TestBoxln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		Boxln("test %s", "content")
	})
	if !strings.Contains(output, "test content") {
		t.Errorf("Boxln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Boxln() should end with newline, got: %s", output)
	}
}

func TestErrorBoxPrint(t *testing.T) {
	output := utils.CaptureOutput(func() {
		ErrorBox("test %s", "error")
	})
	if !strings.Contains(output, "test error") {
		t.Errorf("ErrorBox() should print message, got: %s", output)
	}
}

func TestErrorBoxln(t *testing.T) {
	output := utils.CaptureOutput(func() {
		ErrorBoxln("test %s", "error")
	})
	if !strings.Contains(output, "test error") {
		t.Errorf("ErrorBoxln() should print message, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("ErrorBoxln() should end with newline, got: %s", output)
	}
}

func TestPrintingFunctions(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		tests := []struct {
			name string
			fn   func(string, ...any)
		}{
			{"Success", Success},
			{"Error", Error},
			{"Warning", Warning},
			{"Info", Info},
			{"Title", Title},
			{"Subtitle", Subtitle},
			{"Box", Box},
			{"ErrorBox", ErrorBox},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				output := utils.CaptureOutput(func() {
					tt.fn("plain message")
				})
				if !strings.Contains(output, "plain message") {
					t.Errorf("%s() should print plain message, got: %s", tt.name, output)
				}
			})
		}
	})

	t.Run("with multiple args", func(t *testing.T) {
		output := utils.CaptureOutput(func() {
			Success("test %s %d %v", "message", 42, true)
		})
		if !strings.Contains(output, "test message 42 true") {
			t.Errorf("Success() should handle multiple format args, got: %s", output)
		}
	})

	t.Run("with empty string", func(t *testing.T) {
		output := utils.CaptureOutput(func() {
			Success("")
		})

		if len(output) == 0 {
			t.Errorf("Success(\"\") should print something, got empty output")
		}
	})
}

func TestStyleHelperFunctions(t *testing.T) {
	t.Run("NewStyle", func(t *testing.T) {
		style := newStyle()
		rendered := style.Render("test")
		if rendered != "test" {
			t.Errorf("newStyle() should render plain text, got: %s", rendered)
		}
	})

	t.Run("NewPStyle", func(t *testing.T) {
		style := newPStyle(1, 2)
		rendered := style.Render("test")
		if !strings.Contains(rendered, "test") {
			t.Errorf("newPStyle() should render text, got: %s", rendered)
		}
		if len(rendered) <= len("test") {
			t.Errorf("newPStyle() should add padding, got: %s", rendered)
		}
	})

	t.Run("NewBoldStyle", func(t *testing.T) {
		style := newBoldStyle()
		rendered := style.Render("test")
		if !strings.Contains(rendered, "test") {
			t.Errorf("newBoldStyle() should render text, got: %s", rendered)
		}
	})

	t.Run("NewPBoldStyle", func(t *testing.T) {
		style := newPBoldStyle(1, 2)
		rendered := style.Render("test")
		if !strings.Contains(rendered, "test") {
			t.Errorf("newPBoldStyle() should render text, got: %s", rendered)
		}
		if len(rendered) <= len("test") {
			t.Errorf("newPBoldStyle() should add padding, got: %s", rendered)
		}
	})

	t.Run("NewEmStyle", func(t *testing.T) {
		style := newEmStyle()
		rendered := style.Render("test")
		if !strings.Contains(rendered, "test") {
			t.Errorf("newEmStyle() should render text, got: %s", rendered)
		}
	})
}

func TestStyleVariables(t *testing.T) {
	styles := []struct {
		name  string
		style any
	}{
		{"PrimaryStyle", PrimaryStyle},
		{"AccentStyle", AccentStyle},
		{"ErrorStyle", ErrorStyle},
		{"TextStyle", TextStyle},
		{"TitleStyle", TitleStyle},
		{"SubtitleStyle", SubtitleStyle},
		{"SuccessStyle", SuccessStyle},
		{"WarningStyle", WarningStyle},
		{"InfoStyle", InfoStyle},
		{"BoxStyle", BoxStyle},
		{"ErrorBoxStyle", ErrorBoxStyle},
		{"ListItemStyle", ListItemStyle},
		{"SelectedItemStyle", SelectedItemStyle},
		{"HeaderStyle", HeaderStyle},
		{"CellStyle", CellStyle},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			if s.style == nil {
				t.Errorf("%s should be initialized", s.name)
			}
		})
	}
}

func BenchmarkSuccess(b *testing.B) {
	for b.Loop() {
		success("test message")
	}
}

func BenchmarkSuccessPrint(b *testing.B) {
	old := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = old }()

	for b.Loop() {
		fmt.Print(success(fmt.Sprintf("test %s", "message")))
	}
}

func BenchmarkNewStyle(b *testing.B) {
	for b.Loop() {
		newStyle()
	}
}
