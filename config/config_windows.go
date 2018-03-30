// +build windows

package config

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	shell         = syscall.MustLoadDLL("Shell32.dll")
	getFolderPath = shell.MustFindProc("SHGetFolderPathW")
)

const CSIDL_APPDATA = 26

func configFile() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "forgerock-aws-sts-rc"), nil
}

func configDir() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", nil
	}

	return filepath.Join(dir, "forgerock-aws-sts.d"), nil
}

func homeDir() (string, error) {
	b := make([]uint16, syscall.MAX_PATH)

	r, _, err := getFolderPath.Call(0, CSIDL_APPDATA, 0, 0, uintptr(unsafe.Pointer(&b[0])))
	if uint32(r) != 0 {
		return "", err
	}

	return syscall.UTF16ToString(b), nil
}
