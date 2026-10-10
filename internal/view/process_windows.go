//go:build windows

package view

import (
	"os"
	"os/exec"
	"syscall"
)

// detachedProcess desacopla el proceso de fondo de la consola padre.
const detachedProcess = 0x00000008

// configureDetach arranca el proceso de fondo sin consola (DETACHED_PROCESS).
func configureDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detachedProcess}
}

// pidAlive indica si el proceso pid sigue vivo.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.FindProcess(pid)
	return err == nil
}

// terminateProcess termina el proceso.
func terminateProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
