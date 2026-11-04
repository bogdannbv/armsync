//go:build !windows

package steam

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func suppressStderr(fn func()) (err error) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, devNull.Close())
	}()

	savedStderr, err := dupFd(int(os.Stderr.Fd()))
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, dup2Fd(savedStderr, int(os.Stderr.Fd())))
		err = errors.Join(err, closeFd(savedStderr))
	}()

	if err = dup2Fd(int(devNull.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}

	fn()
	return nil
}

func dupFd(fd int) (int, error) {
	return unix.Dup(fd)
}

func dup2Fd(oldfd, newfd int) error {
	if oldfd == newfd {
		return nil
	}
	return unix.Dup3(oldfd, newfd, 0)
}

func closeFd(fd int) error {
	return unix.Close(fd)
}
