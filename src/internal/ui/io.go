package ui

// Interactor defines an interface for all user interactions,
// allowing for different implementations (e.g., CLI, GUI).
type Interactor interface {
	// Prompt asks the user for textual input.
	Prompt(message string, defaultValue string) (string, error)
	// Confirm asks the user for a yes/no confirmation.
	Confirm(message string, defaultValue bool) (bool, error)
	// Select presents a list of options for the user to choose from.
	Select(message string, options []string, defaultValue string) (string, error)

	// Output displays a standard message to the user.
	Output(message string)
	// Outputf displays a formatted standard message to the user.
	Outputf(format string, a ...interface{})
	// Success displays a success message.
	Success(message string)
	// Successf displays a formatted success message.
	Successf(format string, a ...interface{})
	// Info displays an informational message.
	Info(message string)
	// Infof displays a formatted informational message.
	Infof(format string, a ...interface{})
	// Warning displays a warning message.
	Warning(message string)
	// Warningf displays a formatted warning message.
	Warningf(format string, a ...interface{})
	// Error displays an error message. If err is not nil, its message is included.
	Error(message string, err error)
	// Errorf displays a formatted error message. If err is not nil, its message is included.
	Errorf(format string, err error, a ...interface{})
	// Fatal displays an error message and exits. If err is not nil, its message is included.
	Fatal(message string, err error)
	// Fatalf displays a formatted error message and exits. If err is not nil, its message is included.
	Fatalf(format string, err error, a ...interface{})

	// StartSpinner starts a terminal spinner with a given message.
	StartSpinner(message string)
	// StopSpinner stops the current spinner.
	// success indicates whether the operation succeeded, influencing the spinner's final symbol.
	// message is an optional final message for the spinner.
	StopSpinner(success bool, message string)

	// ShowCustomHelp displays a custom help message.
	// It might take parameters to customize the help output, e.g., showAll, command context.
	ShowCustomHelp(showAll bool, commandPath string) // commandPath could be cmd.CommandPath()

	// TODO: Add methods for tables, progress bars, etc. as needed.
}
