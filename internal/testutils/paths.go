package testutils

import (
	"path/filepath"
	"runtime"
)

// GetTestDirPath returns the absolute path to the testdata directory
func GetTestDirPath() string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	return filepath.Join(projectRoot, "testdata")
}
