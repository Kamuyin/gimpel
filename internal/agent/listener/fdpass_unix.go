//go:build linux || darwin || freebsd || openbsd || netbsd

package listener

import "os"

func fdToFile(fd int) *os.File {
	return os.NewFile(uintptr(fd), "")
}
