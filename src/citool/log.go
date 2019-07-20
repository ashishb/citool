package citool

import "fmt"

var debugMode = true

// SetDebugMode enables/disables dbeug enable mode which enables debug logging.
func SetDebugMode(newMode bool) {
	debugMode = newMode
}

// LogDebug logs if debug mode is enabled. NO-OP otherwise.
func LogDebug(msg string) {
	if !debugMode {
		return
	}
	fmt.Println(msg)
}
