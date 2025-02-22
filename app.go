package main

import (
	"os"
	"path/filepath"
)

func AppExecName() string { return filepath.Base(os.Args[0]) }
func AppFullName() string { return "Mindari's Word Games" }
func AppInitials() string { return "MWG" }
