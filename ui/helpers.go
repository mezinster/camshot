package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// showError displays an error dialog
func showError(parent fyne.Window, message string) {
	dialog.ShowError(errorf(message), parent)
}

// showInfo displays an information dialog
func showInfo(parent fyne.Window, message string) {
	dialog.ShowInformation("Info", message, parent)
}

// errorf creates a simple error from string
type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}

func errorf(msg string) error {
	return &simpleError{msg: msg}
}
