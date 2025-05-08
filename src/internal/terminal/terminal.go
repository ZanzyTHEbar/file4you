package terminal

type Terminal struct{}

func NewTerminal() *Terminal {
	return &Terminal{}
}

// Style constants for terminal output formatting
const (
	StyleQuestion = "question"
	StyleDim      = "dim"
	StyleSuccess  = "success"
	StyleError    = "error"
	StyleInfo     = "info"
	StyleWarning  = "warning"
)

// SpinnerRefreshRate defines how fast the spinner should update (milliseconds)
const SpinnerRefreshRate = 100 // milliseconds

// Output prints a message to the terminal
func (t *Terminal) Output(message string) {
	// In a real implementation, this would handle proper formatting
	println(message)
}

// FormatMessage applies the specified style to a message
func (t *Terminal) FormatMessage(style string, message string) string {
	// In a real implementation, this would apply ANSI color codes based on style
	// For now, we'll just return the message unchanged
	return message
}

// Success prints a success message to the terminal
func (t *Terminal) Success(message string) {
	t.Output(t.FormatMessage(StyleSuccess, "✓ "+message))
}

// Error prints an error message to the terminal
func (t *Terminal) Error(message string, cmd interface{}) {
	t.Output(t.FormatMessage(StyleError, "✗ "+message))
}

// Info prints an informational message to the terminal
func (t *Terminal) Info(message string) {
	t.Output(t.FormatMessage(StyleInfo, "ℹ "+message))
}

// Warning prints a warning message to the terminal
func (t *Terminal) Warning(message string) {
	t.Output(t.FormatMessage(StyleWarning, "⚠ "+message))
}
