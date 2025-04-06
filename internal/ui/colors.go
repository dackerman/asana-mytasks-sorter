package ui

// Simple ANSI color codes for terminal output
const (
	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	
	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
	
	// Text styles
	Bold      = "\033[1m"
	Underline = "\033[4m"
	
	// Reset
	Reset = "\033[0m"
)

// Helper functions for consistent styling

// Header formats text as a heading
func Header(text string) string {
	return Bold + BrightWhite + text + Reset
}

// SectionTitle formats a section title 
func SectionTitle(text string) string {
	return Bold + BrightCyan + text + Reset
}

// Success formats a success message
func Success(text string) string {
	return Green + text + Reset
}

// Warning formats a warning message
func Warning(text string) string {
	return Yellow + text + Reset
}

// Error formats an error message
func Error(text string) string {
	return BrightRed + text + Reset
}

// Info formats an informational message
func Info(text string) string {
	return Cyan + text + Reset
}

// Important highlights important information
func Important(text string) string {
	return Bold + text + Reset
}

// Subtle formats text to be less prominent
func Subtle(text string) string {
	return White + text + Reset
}

// TaskName formats a task name
func TaskName(text string) string {
	return BrightWhite + text + Reset
}

// SectionName formats a section name
func SectionName(text string) string {
	return BrightYellow + text + Reset
}

// DueDate formats a due date
func DueDate(text string) string {
	return BrightGreen + text + Reset
}

// Operation formats an operation description
func Operation(text string) string {
	return Magenta + text + Reset
}