package citool

import "fmt"

var debugMode = true

func SetDebugMode(newMode bool) {
	debugMode = newMode
}

func LogDebug(msg string) {
	if !debugMode {
		return
	}
	fmt.Println(msg)
}
