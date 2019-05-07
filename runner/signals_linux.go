package runner

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func signalWatch(cmd *exec.Cmd) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		syscall.Kill(cmd.Process.Pid, sig.(syscall.Signal))
	}()
}
