//go:build windows

package steam

import (
	"errors"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type restoreFunc func() error

const (
	crtStderrFD = 2
	crtOWronly  = 0x0001
)

func suppressStderr(fn func()) (err error) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, devNull.Close())
	}()

	restoreStdHandle, err := redirectWindowsStderr(windows.Handle(devNull.Fd()))
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, restoreStdHandle())
	}()

	oldStderr := os.Stderr
	os.Stderr = devNull
	defer func() {
		os.Stderr = oldStderr
	}()

	var restoreCRT []restoreFunc
	for _, dll := range []string{"ucrtbase.dll", "msvcrt.dll"} {
		restore, err := redirectCRTStderr(dll)
		if err == nil {
			restoreCRT = append(restoreCRT, restore)
		}
	}
	defer func() {
		for i := len(restoreCRT) - 1; i >= 0; i-- {
			err = errors.Join(err, restoreCRT[i]())
		}
	}()

	fn()
	return nil
}

func redirectWindowsStderr(to windows.Handle) (restoreFunc, error) {
	oldHandle, err := windows.GetStdHandle(windows.STD_ERROR_HANDLE)
	if err != nil {
		return nil, err
	}

	if err := windows.SetStdHandle(windows.STD_ERROR_HANDLE, to); err != nil {
		return nil, err
	}

	return func() error {
		return windows.SetStdHandle(windows.STD_ERROR_HANDLE, oldHandle)
	}, nil
}

func redirectCRTStderr(dll string) (restoreFunc, error) {
	crt := windows.NewLazySystemDLL(dll)
	open := crt.NewProc("_open")
	dup := crt.NewProc("_dup")
	dup2 := crt.NewProc("_dup2")
	clse := crt.NewProc("_close")

	nulPath, err := windows.BytePtrFromString("NUL")
	if err != nil {
		return nil, err
	}

	nullFD, err := crtOpen(open, nulPath)
	if err != nil {
		return nil, err
	}

	savedFD, err := crtDup(dup, crtStderrFD)
	if err != nil {
		_ = crtClose(clse, nullFD)
		return nil, err
	}

	if err = crtDup2(dup2, nullFD, crtStderrFD); err != nil {
		_ = crtClose(clse, savedFD)
		_ = crtClose(clse, nullFD)
		return nil, err
	}

	return func() error {
		err = crtDup2(dup2, savedFD, crtStderrFD)
		err = errors.Join(err, crtClose(clse, savedFD))
		err = errors.Join(err, crtClose(clse, nullFD))
		return err
	}, nil
}

func crtOpen(proc *windows.LazyProc, path *byte) (int, error) {
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(path)), uintptr(crtOWronly))
	if r == ^uintptr(0) {
		return -1, syscall.EINVAL
	}
	return int(r), nil
}

func crtDup(proc *windows.LazyProc, fd int) (int, error) {
	r, _, _ := proc.Call(uintptr(fd))
	if r == ^uintptr(0) {
		return -1, syscall.EINVAL
	}
	return int(r), nil
}

func crtDup2(proc *windows.LazyProc, oldfd, newfd int) error {
	r, _, _ := proc.Call(uintptr(oldfd), uintptr(newfd))
	if r == ^uintptr(0) {
		return syscall.EINVAL
	}
	return nil
}

func crtClose(proc *windows.LazyProc, fd int) error {
	r, _, _ := proc.Call(uintptr(fd))
	if r == ^uintptr(0) {
		return syscall.EINVAL
	}
	return nil
}
