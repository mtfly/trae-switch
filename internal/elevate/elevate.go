//go:build windows

package elevate

import (
	"os"
	"syscall"
	"unsafe"
)

const SW_SHOWNORMAL = 1

func Elevate() error {
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}

	ret, _, _ := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("runas"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(exePath))),
		0,
		0,
		SW_SHOWNORMAL,
	)

	if ret <= 32 {
		return os.NewSyscallError("ShellExecute", syscall.Errno(ret))
	}
	return nil
}