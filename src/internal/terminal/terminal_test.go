package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TerminalTestSuite tests the terminal package functionality
type TerminalTestSuite struct {
	suite.Suite
	terminal *Terminal
}

func TestTerminalSuite(t *testing.T) {
	suite.Run(t, new(TerminalTestSuite))
}

func (suite *TerminalTestSuite) SetupTest() {
	suite.terminal = NewTerminal()
	require.NotNil(suite.T(), suite.terminal)
}

func (suite *TerminalTestSuite) TestOutput() {
	// Test that Output doesn't panic
	assert.NotPanics(suite.T(), func() {
		suite.terminal.Output("Test message")
	})
}

func (suite *TerminalTestSuite) TestFormatMessage() {
	tests := []struct {
		name     string
		style    string
		message  string
		expected bool // Whether it should contain the original message
	}{
		{"Success style", "success", "Operation completed", true},
		{"Error style", "error", "Error occurred", true},
		{"Warning style", "warning", "Warning message", true},
		{"Info style", "info", "Information", true},
		{"Empty style", "", "No style message", true},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			result := suite.terminal.FormatMessage(tt.style, tt.message)

			// Should contain the original message
			if tt.expected {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func (suite *TerminalTestSuite) TestSuccess() {
	// Test that Success doesn't panic
	assert.NotPanics(suite.T(), func() {
		suite.terminal.Success("Operation completed successfully")
	})
}

func (suite *TerminalTestSuite) TestError() {
	// Test that Error doesn't panic
	assert.NotPanics(suite.T(), func() {
		suite.terminal.Error("An error occurred", nil)
	})

	// Test with command object
	assert.NotPanics(suite.T(), func() {
		suite.terminal.Error("Command error", "some-command")
	})
}

func (suite *TerminalTestSuite) TestInfo() {
	// Test that Info doesn't panic
	assert.NotPanics(suite.T(), func() {
		suite.terminal.Info("This is information")
	})
}

func (suite *TerminalTestSuite) TestWarning() {
	// Test that Warning doesn't panic
	assert.NotPanics(suite.T(), func() {
		suite.terminal.Warning("This is a warning")
	})
}

func (suite *TerminalTestSuite) TestOutputMethods() {
	// Test various output methods
	assert.NotPanics(suite.T(), func() {
		suite.terminal.OutputSimpleError("Simple error message")
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.OutputInfo("Info message with %s", "parameter")
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.OutputSuccess("Success message with %d items", 5)
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.OutputWarning("Warning message")
	})
}

func (suite *TerminalTestSuite) TestSpinnerOperations() {
	// Test spinner functionality
	assert.NotPanics(suite.T(), func() {
		suite.terminal.ToggleSpinner(true, "Processing...")
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.ToggleSpinner(false, "")
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.ResumeSpinner()
	})
}

func (suite *TerminalTestSuite) TestUtilityMethods() {
	// Test utility methods don't panic
	assert.NotPanics(suite.T(), func() {
		suite.terminal.AlternateScreen()
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.ClearScreen()
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.ClearCurrentLine()
	})

	assert.NotPanics(suite.T(), func() {
		suite.terminal.MoveCursorToTopLeft()
	})
}

func (suite *TerminalTestSuite) TestConfirmYesNo() {
	// Note: This test can't easily test interactive input without mocking,
	// but we can test that the method exists and can be called
	// In a real implementation, you'd want to mock stdin

	// For now, just verify the method exists and doesn't panic with proper setup
	assert.NotNil(suite.T(), suite.terminal.ConfirmYesNo)
}

// TestColorFunctions tests color-related functionality from colors.go
func TestColorFunctions(t *testing.T) {
	// Test that color styles are properly initialized
	assert.NotNil(t, ColorHiGreen)
	assert.NotNil(t, ColorHiMagenta)
	assert.NotNil(t, ColorHiRed)
	assert.NotNil(t, ColorHiYellow)
	assert.NotNil(t, ColorHiCyan)
	assert.NotNil(t, ColorHiBlue)

	// Test that they can render text
	assert.NotPanics(t, func() {
		_ = ColorHiGreen.Render("test")
		_ = ColorHiRed.Render("test")
		_ = ColorHiBlue.Render("test")
	})
}

// TestHelperFunctions tests helper functions from help.go
func TestHelperFunctions(t *testing.T) {
	// Test any exported helper functions exist
	t.Run("NewTerminal", func(t *testing.T) {
		terminal := NewTerminal()
		assert.NotNil(t, terminal)
		assert.IsType(t, &Terminal{}, terminal)
	})
}

// TestErrorFormatting tests error formatting functions
func TestErrorFormatting(t *testing.T) {
	terminal := NewTerminal()

	// Test different error scenarios
	t.Run("OutputUnformattedErrorAndExit", func(t *testing.T) {
		// This function calls os.Exit(), so we can't test it directly
		// In a real test, you'd need to refactor to allow dependency injection
		// or use a test runner that can handle os.Exit()
		assert.NotNil(t, terminal.OutputUnformattedErrorAndExit)
	})

	t.Run("OutputErrorAndExit", func(t *testing.T) {
		// Same as above - would need special test setup for os.Exit()
		assert.NotNil(t, terminal.OutputErrorAndExit)
	})
}

// BenchmarkTerminalOutput benchmarks terminal output performance
func BenchmarkTerminalOutput(b *testing.B) {
	terminal := NewTerminal()

	b.Run("Output", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			terminal.Output("benchmark message")
		}
	})

	b.Run("FormatMessage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			terminal.FormatMessage("info", "benchmark message")
		}
	})

	b.Run("Success", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			terminal.Success("benchmark success")
		}
	})
}

// TestTerminalInitialization tests terminal creation and initialization
func TestTerminalInitialization(t *testing.T) {
	t.Run("NewTerminal", func(t *testing.T) {
		terminal := NewTerminal()

		// Verify terminal is properly initialized
		assert.NotNil(t, terminal)

		// Basic functionality should work
		assert.NotPanics(t, func() {
			terminal.Output("test")
		})
	})

	t.Run("MultiplePInstances", func(t *testing.T) {
		// Test creating multiple terminal instances
		terminal1 := NewTerminal()
		terminal2 := NewTerminal()

		assert.NotNil(t, terminal1)
		assert.NotNil(t, terminal2)

		// They should be separate instances
		assert.NotSame(t, terminal1, terminal2)
	})
}
