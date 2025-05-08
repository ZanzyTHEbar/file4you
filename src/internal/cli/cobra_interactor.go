package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"file4you/internal/terminal" // Assuming your existing terminal functions are here
	"file4you/internal/ui"

	"github.com/briandowns/spinner" // Assuming you use this or similar
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
	input, err := reader.ReadString('\n')
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
	// TODO: Assuming terminal.StartSpinner returns a spinner instance and takes a message
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
		// If there's a message, print it even if spinner wasn't formally started by this interactor
		if message != "" {
			if success {
				ci.Success(message)
			} else {
				ci.Error(message, nil) // Or Warning, depending on context
			}
		}
		return
	}

	if message != "" {
		if success {
			ci.activeSpinner.FinalMSG = ci.term.FormatMessage(terminal.StyleSuccess, "✔ ") + message + "\n"
		} else {
			ci.activeSpinner.FinalMSG = ci.term.FormatMessage(terminal.StyleError, "✖ ") + message + "\n"
		}
	}

	if success {
		// TODO: Optional: Change color or symbol on success before stopping
		// ci.activeSpinner.Color("green") // Example if spinner supports color changes
	}

	ci.activeSpinner.Stop()
	ci.activeSpinner = nil
}

// ShowCustomHelp implements the corresponding method in the Interactor interface.
// It calls the terminal's PrintCustomHelp method.
func (ci *CobraInteractor) ShowCustomHelp(showAll bool, commandPath string) {
	// TODO: Assuming terminal.PrintCustomHelp is the function we want to call.
	// It might need the root command or specific command to display relevant help.
	// The original call was params.term.PrintCustomHelp(helpShowAll)
	// We need to ensure ci.term has access to the necessary cobra.Command object
	// or that PrintCustomHelp can function with just `showAll` and perhaps a command path string.

	// TODO: If your terminal.PrintCustomHelp function is defined as:
	// func (t *Terminal) PrintCustomHelp(showAll bool, cmd *cobra.Command, palette []*cobra.Command)
	// then CobraInteractor needs access to the command and palette.
	// This is a limitation of the current Interactor design if it doesn't pass commands through.

	// For now, let's assume PrintCustomHelp can be called on the term instance
	// and it handles how to get command info internally or via global state if necessary (less ideal).
	// This is a simplification.
	// A more robust way would be for `NewHelp` to pass the `*cobra.Command` to this method,
	// requiring a change in the `ShowCustomHelp` signature in `ui.Interactor` and here.

	// Placeholder: This simulates calling a method on the terminal instance.
	// You will need to replace this with the actual call to your terminal's help function.
	// For example, if your terminal package has a function `PrintHelp(term *terminal.Terminal, showAll bool, commandPath string)`:
	// terminal.PrintHelp(ci.term, showAll, commandPath)

	// Or if PrintCustomHelp is a method of ci.term:
	// ci.term.PrintCustomHelp(showAll, commandPath) // Adjust signature as needed

	// Based on the previous help.go, it seemed to use a global `CmdDesc` and `PrintCustomCmd`
	// and a `PrintCustomHelp` that might iterate through commands.
	// Let's try to replicate a simplified version of what `terminal.PrintCustomHelp` might do.

	ci.Output("--- Custom Help ---")
	if showAll {
		ci.Info("Displaying all available commands (simulated):")
		// TODO: In a real scenario, you'd iterate over registered commands.
		// For example, if you have access to the root command:
		// for _, cmd := range rootCmd.Commands() {
		//    ci.Outputf("  %s - %s", cmd.Name(), cmd.Short)
		// }
		// Using the CmdDesc from terminal package as an example of what might be shown:
		for cmd, details := range terminal.CmdDesc {
			alias := ""
			if details[0] != "" {
				alias = fmt.Sprintf(" (alias: %s)", details[0])
			}
			ci.Outputf("  %s%s: %s", cmd, alias, details[1])
		}
	} else {
		ci.Info(fmt.Sprintf("Displaying help for command: %s (simulated)", commandPath))
		// Logic to display specific help for `commandPath`
		// This might involve looking up the command and printing its Long field and usage.
		details, ok := terminal.CmdDesc[commandPath] // This is a simplification
		if ok {
			ci.Outputf("  Description: %s", details[1])
		} else {
			ci.Warningf("No specific custom help available for %s. Try 'detailed_help --all'.", commandPath)
		}
	}
	ci.Output("--- End Custom Help ---")
	// The original call in help.go was: term.PrintCustomHelp(helpShowAll, cmd.Root())
	// This indicates that PrintCustomHelp in the terminal package likely expects a *cobra.Command.
	// To truly replicate this, CobraInteractor would need access to the command, or PrintCustomHelp
	// would need to be refactored. The current implementation is a placeholder.
}

// Ensure CobraInteractor implements ui.Interactor
var _ ui.Interactor = (*CobraInteractor)(nil)
