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
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		fmt.Println(sig)
		cmd.Process.Kill()
	}()
}
