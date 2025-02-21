package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

// Logging wrapper - Includes file and line number
func logPrintln(format string, v ...any) {
	// Capitalization for variables in final print statement
	Message := fmt.Sprintf(format, v...)
	pc, file, Line, ok := runtime.Caller(1)
	// This shouldn't fail, but do not swallow message if it does
	if !ok {
		log.Printf("%s:%d:%s %s\n", "UNKNOWN", 0, "UNKNOWN", Message)
	} else {
		fnWithModule := runtime.FuncForPC(pc).Name()
		fnParts := strings.Split(fnWithModule, ".")
		// Filename without path
		Name := filepath.Base(file)
		// Function name without module
		Fn := fnParts[len(fnParts)-1]
		log.Printf("%s:%d:%s %s\n", Name, Line, Fn, Message)
	}
}
