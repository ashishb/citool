package citool

import "fmt"

const debugMode = true

func LogDebug(msg string) {
    if !debugMode {
        return
    }
    fmt.Println(msg)
}
