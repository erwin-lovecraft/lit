package ioutil

import (
	"os"
	"strings"
	"sync"
)

var (
	// Path to the resources directory
	resourceDir = "resources"
	once        = sync.Once{}
)

// SetResourceDir sets the path to the resources directory
// This function is safe to call multiple times; it will only set the path once
func SetResourceDir(path string) {
	once.Do(func() {
		resourceDir = path
	})
}

// ReadFile reads a file from the resources directory
func ReadFile(fileName string) ([]byte, error) {
	// Make sure the file path is relative to the resources directory
	fileName = strings.Trim(fileName, "/")
	fileName = resourceDir + "/" + fileName

	return os.ReadFile(fileName)
}
