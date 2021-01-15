package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func signalWatch(cmd *exec.Cmd) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		fmt.Println("Received Signal", sig, "sending SIGTERM to child", cmd.Process.Pid)
		syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	}()
}
