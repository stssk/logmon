//go:build windows
// +build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func openFileReadOnly(filePath string) (*os.File, error) {
	path, err := windows.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(path, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(handle), filePath), nil
}
