package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"file4you/internal/terminal" // Assuming your existing terminal functions are here
	"file4you/internal/ui"

	"github.com/briandowns/spinner" // Assuming you use this or similar
	"github.com/spf13/cobra"
)

// CobraInteractor implements the ui.Interactor interface for Cobra CLI interactions.
type CobraInteractor struct {
	term         *terminal.Terminal // Changed to pointer
	activeSpinner *spinner.Spinner
	// Add other necessary fields, e.g., for styling
}

// NewCobraInteractor creates a new instance of CobraInteractor.
func NewCobraInteractor(term *terminal.Terminal) ui.Interactor { // Changed to pointer
	return &CobraInteractor{term: term}
}

func (ci *CobraInteractor) Prompt(message string, defaultValue string) (string, error) {
	ci.term.Output(fmt.Sprintf("%s ", ci.term.FormatMessage(terminal.StyleQuestion, message)))
	if defaultValue != "" {
		ci.term.Output(ci.term.FormatMessage(terminal.StyleDim, fmt.Sprintf("(default: %s) ", defaultValue)))
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func (ci *CobraInteractor) Confirm(message string, defaultValue bool) (bool, error) {
	promptMessage := message
	if defaultValue {
		promptMessage = fmt.Sprintf("%s (Y/n)", message)
	} else {
		promptMessage = fmt.Sprintf("%s (y/N)", message)
	}

	val, err := ci.Prompt(promptMessage, "")
	if err != nil {
		return false, err
	}

	val = strings.ToLower(strings.TrimSpace(val))
	if val == "" {
		return defaultValue, nil
	}
	if val == "y" || val == "yes" {
		return true, nil
	}
	if val == "n" || val == "no" {
		return false, nil
	}
	// Re-prompt if invalid input, or return error - for simplicity, returning default
	ci.Warningf("Invalid input. Defaulting to %v.", defaultValue)
	return defaultValue, nil
}

func (ci *CobraInteractor) Select(message string, options []string, defaultValue string) (string, error) {
	ci.Output(message)
	for i, opt := range options {
		prefix := "  "
		if opt == defaultValue {
			prefix = "* "
		}
		ci.Outputf("%s%d. %s", prefix, i+1, opt)
	}

	choiceMsg := "Enter your choice"
	if defaultValue != "" {
		choiceMsg = fmt.Sprintf("Enter your choice (default: %s)", defaultValue)
	}

	inputStr, err := ci.Prompt(choiceMsg, "")
	if err != nil {
		return defaultValue, err // Or handle error more gracefully
	}

	if inputStr == "" && defaultValue != "" {
		return defaultValue, nil
	}

	// TODO: Add proper input validation for selection (number or direct string match)
	// For now, assume direct string match or return default if empty
	for _, opt := range options {
		if strings.EqualFold(inputStr, opt) {
			return opt, nil
		}
	}
	if defaultValue != "" {
		ci.Warningf("Invalid selection. Defaulting to %s.", defaultValue)
		return defaultValue, nil
	}
	return "", fmt.Errorf("invalid selection: %s", inputStr)
}

func (ci *CobraInteractor) Output(message string) {
	ci.term.Output(message)
}

func (ci *CobraInteractor) Outputf(format string, a ...interface{}) {
	ci.term.Output(fmt.Sprintf(format, a...))
}

func (ci *CobraInteractor) Success(message string) {
	ci.term.Success(message)
}

func (ci *CobraInteractor) Successf(format string, a ...interface{}) {
	ci.term.Success(fmt.Sprintf(format, a...))
}

func (ci *CobraInteractor) Info(message string) {
	ci.term.Info(message)
}

func (ci *CobraInteractor) Infof(format string, a ...interface{}) {
	ci.term.Info(fmt.Sprintf(format, a...))
}

func (ci *CobraInteractor) Warning(message string) {
	ci.term.Warning(message)
}

func (ci *CobraInteractor) Warningf(format string, a ...interface{}) {
	ci.term.Warning(fmt.Sprintf(format, a...))
}

func (ci *CobraInteractor) Error(message string, err error) {
	fullMessage := message
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", message, err)
	}
	ci.term.Error(fullMessage, nil) // Assuming term.Error takes an optional command
}

func (ci *CobraInteractor) Errorf(format string, err error, a ...interface{}) {
	baseMessage := fmt.Sprintf(format, a...)
	fullMessage := baseMessage
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", baseMessage, err)
	}
	ci.term.Error(fullMessage, nil)
}

func (ci *CobraInteractor) Fatal(message string, err error) {
	fullMessage := message
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", message, err)
	}
	ci.term.Error(fullMessage, nil) // Assuming term.Error can handle this
	os.Exit(1)
}

func (ci *CobraInteractor) Fatalf(format string, err error, a ...interface{}) {
	baseMessage := fmt.Sprintf(format, a...)
	fullMessage := baseMessage
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", baseMessage, err)
	}
	ci.term.Error(fullMessage, nil)
	os.Exit(1)
}

func (ci *CobraInteractor) StartSpinner(message string) {
	if ci.activeSpinner != nil {
		ci.activeSpinner.Stop()
	}
	// Assuming terminal.StartSpinner returns a spinner instance and takes a message
	// This part needs to align with how your terminal.StartSpinner works.
	// For now, using github.com/briandowns/spinner as a placeholder.
	// You'll need to integrate this with your existing terminal.Terminal
	// For example, your terminal.Terminal might have StartSpinner/StopSpinner methods.
	// If terminal.Spinner is a type:
	// ci.activeSpinner = ci.term.NewSpinner(message) // Or however it's created
	// ci.activeSpinner.Start()

	// Placeholder if using briandowns/spinner directly and your term doesn't manage it:
	s := spinner.New(spinner.CharSets[9], terminal.SpinnerRefreshRate)
	s.Suffix = " " + message
	s.Start()
	ci.activeSpinner = s
}

func (ci *CobraInteractor) StopSpinner(success bool, message string) {
	if ci.activeSpinner == nil {
		if message != "" { // If there was no spinner but a message is provided, just print it.
			if success {
				ci.Success(message)
			} else {
				ci.Error(message, nil)
			}
		}
		return
	}

	originalSuffix := ci.activeSpinner.Suffix
	if success {
		ci.activeSpinner.FinalMSG = ci.term.FormatMessage(terminal.StyleSuccess, "✔"+originalSuffix+" "+message+"
")
	} else {
		ci.activeSpinner.FinalMSG = ci.term.FormatMessage(terminal.StyleError, "✖"+originalSuffix+" "+message+"
")
	}
	ci.activeSpinner.Stop()
	ci.activeSpinner = nil
}

// Ensure CobraInteractor implements ui.Interactor
var _ ui.Interactor = (*CobraInteractor)(nil)
