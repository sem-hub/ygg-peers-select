//go:build linux || darwin || freebsd || netbsd || openbsd

package utils

import (
	"os"
	"os/exec"
	"syscall"
)

func AsAdmin() bool {
	return os.Geteuid() == 0
}

func RusAsAdmin() error {
	cmd := exec.Command("sudo", os.Args[0])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set UID и GID root
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: 0, Gid: 0}

	// Pass arguments to program
	cmd.Args = append(cmd.Args, os.Args[1:]...)

	// Execute program as root
	err := cmd.Run()
	if err != nil {
		return err
	}

	// Exit current process
	os.Exit(0)
	// To fix build error:
	return nil
}
