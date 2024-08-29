//go:build !windows
// +build !windows

package main

import (
	"os"
)

func openFileReadOnly(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_RDONLY, 0)
}
