package main

import (
	"os"
	"path/filepath"
)

// Shown in usage
func AppExecName() string { return filepath.Base(os.Args[0]) }

// Shown in usage and web page titles
func AppFullName() string { return "Mindari's Word Games" }

// Used by logger
func AppInitials() string { return "MWG" }
