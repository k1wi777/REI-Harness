//go:build !windows

package view

import (
	"os"
	"os/exec"
	"syscall"
)

// configureDetach desacopla el proceso de fondo del terminal (setsid).
func configureDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// pidAlive indica si el proceso pid sigue vivo (kill 0).
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

// terminateProcess envía SIGTERM al proceso.
func terminateProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}
