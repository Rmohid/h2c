package cli

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func runDaemonShellCommand() error {
	h2d := os.Args[0]
	cmd := exec.Command(h2d, "start")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("Failed to run '%v start': %v", h2d, err)
	}
	for i := 0; i < 3; i++ {
		cmd = exec.Command(h2d, "pid")
		err = cmd.Run()
		if err == nil {
			return nil // If 'h2d pid' returns no error, the process should be running successfully.
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return fmt.Errorf("Failed to run '%v start'", h2d)
}
