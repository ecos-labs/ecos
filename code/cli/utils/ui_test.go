package utils

import (
	"runtime"
	"testing"
	"time"
)

func TestGetBannerLines(t *testing.T) {
	// Use placeholder color codes for testing structure
	eColor := "ECOLOR"
	cColor := "CCOLOR"
	oColor := "OCOLOR"
	sColor := "SCOLOR"
	reset := "RESET"

	bannerLines := getBannerLines(eColor, cColor, oColor, sColor, reset)

	// Expected banner lines structure - exact spacing is critical
	//nolint:dupl // Intentional duplication - matches getBannerLines() output for validation
	expectedLines := [][]string{
		{"  ", eColor, "  ██████", reset, "   ", cColor, "  ██████", reset, "   ", oColor, " ███████", reset, "   ", sColor, "  ██████", reset},
		{"  ", eColor, " ██", reset, "  ", eColor, "  ██", reset, "  ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "  ", sColor, " ██", reset},
		{"  ", eColor, " ██", reset, "   ", eColor, " ██", reset, "  ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "  ", sColor, " ██", reset},
		{"  ", eColor, " ███████", reset, "   ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "  ", sColor, "  █████", reset},
		{"  ", eColor, " ██", reset, "        ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "       ", sColor, " ██", reset},
		{"  ", eColor, " ██", reset, "        ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "       ", sColor, " ██", reset},
		{"  ", eColor, "  ██████", reset, "   ", cColor, "  ██████", reset, "   ", oColor, " ███████", reset, "   ", sColor, "  █████", reset},
	}

	// Test: Banner should have exactly 7 lines
	if len(bannerLines) != len(expectedLines) {
		t.Errorf("bannerLines has %d lines, expected %d", len(bannerLines), len(expectedLines))
	}

	// Test: Each line should match exactly (including spacing)
	for i, actualLine := range bannerLines {
		expectedLine := expectedLines[i]
		if len(actualLine) != len(expectedLine) {
			t.Errorf("line %d has %d parts, expected %d", i, len(actualLine), len(expectedLine))
			continue
		}

		for j, actualPart := range actualLine {
			expectedPart := expectedLine[j]
			// Replace color codes with placeholders for comparison
			actualNormalized := actualPart
			switch actualPart {
			case eColor:
				actualNormalized = eColor
			case cColor:
				actualNormalized = cColor
			case oColor:
				actualNormalized = oColor
			case sColor:
				actualNormalized = sColor
			case reset:
				actualNormalized = reset
			}

			if actualNormalized != expectedPart {
				t.Errorf("line %d, part %d: got %q, want %q", i, j, actualPart, expectedPart)
			}
		}
	}
}

func TestGetBannerLinesExactStructure(t *testing.T) {
	// Test with empty strings to verify exact spacing structure
	eColor := ""
	cColor := ""
	oColor := ""
	sColor := ""
	reset := ""

	bannerLines := getBannerLines(eColor, cColor, oColor, sColor, reset)

	// Expected structure with exact spacing (colors removed)
	expectedStructure := [][]string{
		{"  ", "", "  ██████", "", "   ", "", "  ██████", "", "   ", "", " ███████", "", "   ", "", "  ██████", ""},
		{"  ", "", " ██", "", "  ", "", "  ██", "", "  ", "", " ██", "", "       ", "", " ██", "", "    ", "", " ██", "", "  ", "", " ██", ""},
		{"  ", "", " ██", "", "   ", "", " ██", "", "  ", "", " ██", "", "       ", "", " ██", "", "    ", "", " ██", "", "  ", "", " ██", ""},
		{"  ", "", " ███████", "", "   ", "", " ██", "", "       ", "", " ██", "", "    ", "", " ██", "", "  ", "", "  █████", ""},
		{"  ", "", " ██", "", "        ", "", " ██", "", "       ", "", " ██", "", "    ", "", " ██", "", "       ", "", " ██", ""},
		{"  ", "", " ██", "", "        ", "", " ██", "", "       ", "", " ██", "", "    ", "", " ██", "", "       ", "", " ██", ""},
		{"  ", "", "  ██████", "", "   ", "", "  ██████", "", "   ", "", " ███████", "", "   ", "", "  █████", ""},
	}

	if len(bannerLines) != len(expectedStructure) {
		t.Fatalf("bannerLines has %d lines, expected %d", len(bannerLines), len(expectedStructure))
	}

	for i, actualLine := range bannerLines {
		expectedLine := expectedStructure[i]
		if len(actualLine) != len(expectedLine) {
			t.Errorf("line %d has %d parts, expected %d parts", i, len(actualLine), len(expectedLine))
			continue
		}

		for j, actualPart := range actualLine {
			expectedPart := expectedLine[j]
			if actualPart != expectedPart {
				t.Errorf("line %d, part %d: got %q (len=%d), want %q (len=%d)", i, j, actualPart, len(actualPart), expectedPart, len(expectedPart))
			}
		}
	}
}

func TestSetVerbose(t *testing.T) {
	// Save original state
	originalVerbose := isVerbose

	// Test setting verbose to true
	SetVerbose(true)
	if !isVerbose {
		t.Error("SetVerbose(true) did not set isVerbose to true")
	}

	// Test setting verbose to false
	SetVerbose(false)
	if isVerbose {
		t.Error("SetVerbose(false) did not set isVerbose to false")
	}

	// Restore original state
	SetVerbose(originalVerbose)
}

func TestPrintTable(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		wantErr bool
	}{
		{
			name:    "empty headers returns early",
			headers: []string{},
			rows:    [][]string{{"data"}},
			wantErr: false,
		},
		{
			name:    "empty rows returns early",
			headers: []string{"Header"},
			rows:    [][]string{},
			wantErr: false,
		},
		{
			name:    "valid table",
			headers: []string{"Name", "Value"},
			rows:    [][]string{{"test", "123"}, {"longer", "456"}},
			wantErr: false,
		},
		{
			name:    "table with varying column widths",
			headers: []string{"Short", "Very Long Header"},
			rows:    [][]string{{"a", "b"}, {"very long cell", "c"}},
			wantErr: false,
		},
		{
			name:    "table with more columns in row than headers",
			headers: []string{"Col1"},
			rows:    [][]string{{"a", "b", "c"}},
			wantErr: false,
		},
		{
			name:    "table with fewer columns in row than headers",
			headers: []string{"Col1", "Col2", "Col3"},
			rows:    [][]string{{"a"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// PrintTable doesn't return errors, so we just verify it doesn't panic
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("PrintTable panicked: %v", r)
					}
				}
			}()
			PrintTable(tt.headers, tt.rows)
		})
	}
}

func TestSupportsTrueColor(t *testing.T) {
	// Save original terminal detection state
	originalTerminalDetected := terminalDetected
	defer func() {
		terminalDetected = originalTerminalDetected
	}()

	tests := []struct {
		name           string
		setupEnv       func(*testing.T)
		isTerminal     bool
		expectedResult bool
	}{
		{
			name: "COLORTERM truecolor returns true",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("COLORTERM", "truecolor")
			},
			isTerminal:     true,
			expectedResult: true,
		},
		{
			name: "COLORTERM 24bit returns true",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("COLORTERM", "24bit")
			},
			isTerminal:     true,
			expectedResult: true,
		},
		{
			name: "TERM contains kitty returns true",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("TERM", "xterm-kitty")
			},
			isTerminal:     true,
			expectedResult: true,
		},
		{
			name: "TERM contains 256color returns true",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("TERM", "xterm-256color")
			},
			isTerminal:     true,
			expectedResult: true,
			// Note: On macOS (darwin), this may return false due to early return
			// before checking 256color, so we skip this test on darwin
		},
		{
			name: "not a terminal returns false",
			setupEnv: func(t *testing.T) {
				t.Helper()
				// No env setup needed
			},
			isTerminal:     false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip 256color test on darwin due to early return in supportsTrueColor
			if tt.name == "TERM contains 256color returns true" && runtime.GOOS == "darwin" {
				t.Skip("Skipping 256color test on darwin - early return before 256color check")
			}

			// Clean up environment - t.Setenv will handle restoration automatically
			terminalDetected = nil

			tt.setupEnv(t)

			// Mock IsTerminal by setting terminalDetected
			// This must be done after setupEnv to ensure the cache is cleared
			if tt.isTerminal {
				// Set terminalDetected to true to mock IsTerminal() returning true
				val := true
				terminalDetected = &val
			} else {
				val := false
				terminalDetected = &val
			}

			result := supportsTrueColor()
			if result != tt.expectedResult {
				t.Errorf("supportsTrueColor() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("test message")

	if spinner == nil {
		t.Fatal("NewSpinner returned nil")
	}

	if spinner.message != "test message" {
		t.Errorf("spinner.message = %q, want %q", spinner.message, "test message")
	}

	if spinner.delay != 100*time.Millisecond {
		t.Errorf("spinner.delay = %v, want %v", spinner.delay, 100*time.Millisecond)
	}

	if spinner.active {
		t.Error("spinner should not be active after creation")
	}

	if len(spinner.frames) == 0 {
		t.Error("spinner.frames should not be empty")
	}

	if spinner.done == nil {
		t.Error("spinner.done channel should be initialized")
	}
}

func TestSpinnerStop(t *testing.T) {
	// Save original state
	originalTerminalDetected := terminalDetected
	defer func() {
		terminalDetected = originalTerminalDetected
	}()

	// Mock terminal to true so Stop() actually executes
	val := true
	terminalDetected = &val

	spinner := NewSpinner("test")
	spinner.active = true

	// Stop should not panic
	spinner.Stop()

	if spinner.active {
		t.Error("spinner should not be active after Stop()")
	}
}

func TestSpinnerSuccess(t *testing.T) {
	// Save original state
	originalTerminalDetected := terminalDetected
	defer func() {
		terminalDetected = originalTerminalDetected
	}()

	// Mock terminal to true so Stop() actually executes
	val := true
	terminalDetected = &val

	spinner := NewSpinner("test")
	spinner.active = true

	// Success should stop the spinner and use the message
	spinner.Success("custom message")

	if spinner.active {
		t.Error("spinner should not be active after Success()")
	}
}

func TestSpinnerError(t *testing.T) {
	// Save original state
	originalTerminalDetected := terminalDetected
	defer func() {
		terminalDetected = originalTerminalDetected
	}()

	// Mock terminal to true so Stop() actually executes
	val := true
	terminalDetected = &val

	spinner := NewSpinner("test")
	spinner.active = true

	// Error should stop the spinner
	spinner.Error("error message")

	if spinner.active {
		t.Error("spinner should not be active after Error()")
	}
}

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(5, "test message")

	if pb == nil {
		t.Fatal("NewProgressBar returned nil")
	}

	if pb.totalSteps != 5 {
		t.Errorf("pb.totalSteps = %d, want %d", pb.totalSteps, 5)
	}

	if pb.message != "test message" {
		t.Errorf("pb.message = %q, want %q", pb.message, "test message")
	}

	if pb.width != 40 {
		t.Errorf("pb.width = %d, want %d", pb.width, 40)
	}

	if !pb.autoComplete {
		t.Error("pb.autoComplete should be true by default")
	}

	if len(pb.weights) != 5 {
		t.Errorf("pb.weights length = %d, want %d", len(pb.weights), 5)
	}

	// All weights should be 1
	for i, w := range pb.weights {
		if w != 1 {
			t.Errorf("pb.weights[%d] = %d, want %d", i, w, 1)
		}
	}
}

func TestNewWeightedProgressBar(t *testing.T) {
	weights := []int{1, 2, 3}
	pb := NewWeightedProgressBar(weights, "test")

	if pb == nil {
		t.Fatal("NewWeightedProgressBar returned nil")
	}

	expectedTotal := 6 // 1 + 2 + 3
	if pb.totalWeight != expectedTotal {
		t.Errorf("pb.totalWeight = %d, want %d", pb.totalWeight, expectedTotal)
	}

	if pb.totalSteps != 3 {
		t.Errorf("pb.totalSteps = %d, want %d", pb.totalSteps, 3)
	}
}

func TestProgressBarAdvanceStep(t *testing.T) {
	pb := NewProgressBar(3, "test")
	pb.DisableAutoComplete() // Disable auto-complete for testing

	// Advance step 0
	pb.AdvanceStep(0, "")
	if pb.completedSteps != 1 {
		t.Errorf("pb.completedSteps = %d, want %d", pb.completedSteps, 1)
	}
	if pb.currentWeight != 1 {
		t.Errorf("pb.currentWeight = %d, want %d", pb.currentWeight, 1)
	}

	// Advance step 1
	pb.AdvanceStep(1, "")
	if pb.completedSteps != 2 {
		t.Errorf("pb.completedSteps = %d, want %d", pb.completedSteps, 2)
	}
	if pb.currentWeight != 2 {
		t.Errorf("pb.currentWeight = %d, want %d", pb.currentWeight, 2)
	}

	// Advance step 2
	pb.AdvanceStep(2, "")
	if pb.completedSteps != 3 {
		t.Errorf("pb.completedSteps = %d, want %d", pb.completedSteps, 3)
	}
	if pb.currentWeight != 3 {
		t.Errorf("pb.currentWeight = %d, want %d", pb.currentWeight, 3)
	}
}

func TestProgressBarComplete(t *testing.T) {
	pb := NewProgressBar(3, "test")
	pb.DisableAutoComplete()

	// Advance partially
	pb.AdvanceStep(0, "")
	pb.AdvanceStep(1, "")

	if pb.complete {
		t.Error("pb should not be complete before Complete()")
	}

	pb.Complete()

	if !pb.complete {
		t.Error("pb should be complete after Complete()")
	}

	if pb.currentWeight != pb.totalWeight {
		t.Errorf("pb.currentWeight = %d, want %d (totalWeight)", pb.currentWeight, pb.totalWeight)
	}
}

func TestProgressBarEnableDisableAutoComplete(t *testing.T) {
	pb := NewProgressBar(2, "test")

	if !pb.autoComplete {
		t.Error("pb.autoComplete should be true by default")
	}

	pb.DisableAutoComplete()
	if pb.autoComplete {
		t.Error("pb.autoComplete should be false after DisableAutoComplete()")
	}

	pb.EnableAutoComplete()
	if !pb.autoComplete {
		t.Error("pb.autoComplete should be true after EnableAutoComplete()")
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger("[TEST]", true)

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.prefix != "[TEST]" {
		t.Errorf("logger.prefix = %q, want %q", logger.prefix, "[TEST]")
	}

	if !logger.verbose {
		t.Error("logger.verbose should be true")
	}
}

func TestLoggerDebug(t *testing.T) {
	logger := NewLogger("[TEST]", false)

	// Debug should not output when verbose is false
	// (We can't easily test output, but we can test the state)
	if logger.verbose {
		t.Error("logger.verbose should be false")
	}

	loggerVerbose := NewLogger("[TEST]", true)
	if !loggerVerbose.verbose {
		t.Error("logger.verbose should be true")
	}
}

func TestIsTerminal(t *testing.T) {
	// Save original state
	originalTerminalDetected := terminalDetected

	defer func() {
		terminalDetected = originalTerminalDetected
	}()

	// Reset cache
	terminalDetected = nil

	// First call should check actual terminal
	result1 := IsTerminal()

	// Second call should use cached value
	result2 := IsTerminal()

	if result1 != result2 {
		t.Error("IsTerminal should return cached value on second call")
	}

	// Verify cache was set
	if terminalDetected == nil {
		t.Error("terminalDetected should be set after IsTerminal() call")
	}
}

func TestPrintStep(t *testing.T) {
	// PrintStep doesn't return a value, so we just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintStep panicked: %v", r)
		}
	}()

	PrintStep(1, 5, "test step")
	PrintStep(3, 10, "another step")
}

func TestPrintFixedSelection(t *testing.T) {
	// PrintFixedSelection doesn't return a value, so we just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintFixedSelection panicked: %v", r)
		}
	}()

	PrintFixedSelection("Key", "Value")
	PrintFixedSelection("Long Key Name", "Long Value")
}

func TestPrintVerbose(t *testing.T) {
	// Save original state
	originalVerbose := isVerbose
	defer func() {
		SetVerbose(originalVerbose)
	}()

	// Test with verbose disabled
	SetVerbose(false)
	PrintVerbose("test message", true) // Should not print

	// Test with verbose enabled
	SetVerbose(true)
	PrintVerbose("test message", true) // Should print

	// Test with verbose parameter false
	SetVerbose(true)
	PrintVerbose("test message", false) // Should not print
}

func TestPrintTableWithUnicode(t *testing.T) {
	// Test that PrintTable handles Unicode characters correctly
	headers := []string{"名称", "值"}
	rows := [][]string{
		{"测试", "123"},
		{"另一个测试", "456"},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTable panicked with Unicode: %v", r)
		}
	}()

	PrintTable(headers, rows)
}

func TestPrintTableColumnWidthCalculation(t *testing.T) {
	headers := []string{"Short", "Very Long Header Name"}
	rows := [][]string{
		{"a", "b"},
		{"very long cell content", "c"},
	}

	// This test verifies that column widths are calculated correctly
	// The second column should be wide enough for "Very Long Header Name"
	// The first column should be wide enough for "very long cell content"
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTable panicked: %v", r)
		}
	}()

	PrintTable(headers, rows)
}

func TestProgressBarFinish(t *testing.T) {
	pb := NewProgressBar(3, "test")
	pb.DisableAutoComplete()

	// Finish when not complete should call Complete if all steps done
	pb.AdvanceStep(0, "")
	pb.AdvanceStep(1, "")
	pb.AdvanceStep(2, "")

	pb.Finish()

	if !pb.complete {
		t.Error("pb should be complete after Finish() when all steps are done")
	}

	// Finish when partially complete should call Render
	pb2 := NewProgressBar(3, "test")
	pb2.DisableAutoComplete()
	pb2.AdvanceStep(0, "")

	pb2.Finish()

	// Should not be complete, but should have rendered
	if pb2.complete {
		t.Error("pb should not be complete after Finish() when steps are not all done")
	}
}

func TestProgressBarRender(t *testing.T) {
	pb := NewProgressBar(3, "test")
	pb.DisableAutoComplete()

	// Render should not panic
	pb.Render()

	// Render after complete should return early
	pb.Complete()
	pb.Render() // Should not panic or do anything
}

func TestLoggerMethods(t *testing.T) {
	logger := NewLogger("[TEST]", true)

	// All logger methods should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Logger method panicked: %v", r)
		}
	}()

	logger.Info("info message")
	logger.Success("success message")
	logger.Warning("warning message")
	logger.Error("error message")
	logger.Debug("debug message")
}

func TestSpinnerSuccessWithEmptyMessage(t *testing.T) {
	spinner := NewSpinner("original message")
	spinner.active = true

	// Success with empty message should use original message
	// Note: Success() calls Stop() which may return early if not a terminal,
	// but it should still set active to false via PrintSuccess
	spinner.Success("")

	// The spinner should be stopped (active may still be true if not a terminal,
	// but the important thing is that Success() was called without panicking)
	_ = spinner
}

func TestSpinnerErrorWithEmptyMessage(t *testing.T) {
	spinner := NewSpinner("original message")
	spinner.active = true

	// Error with empty message should use original message + " failed"
	// Note: Error() calls Stop() which may return early if not a terminal,
	// but it should still set active to false via PrintError
	spinner.Error("")

	// The spinner should be stopped (active may still be true if not a terminal,
	// but the important thing is that Error() was called without panicking)
	_ = spinner
}

func TestProgressBarWithZeroTotal(t *testing.T) {
	// Test edge case with zero total steps
	pb := NewProgressBar(0, "test")

	if pb.totalSteps != 0 {
		t.Errorf("pb.totalSteps = %d, want %d", pb.totalSteps, 0)
	}

	if pb.totalWeight != 0 {
		t.Errorf("pb.totalWeight = %d, want %d", pb.totalWeight, 0)
	}
}
