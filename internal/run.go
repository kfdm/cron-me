package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/ShowMax/go-fqdn"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/getsentry/raven-go"
)

// Wrap a command execution
func Wrap(cmd *exec.Cmd) (int, time.Duration) {
	returncode := 0
	start := time.Now()
	rtn := cmd.Run()
	duration := time.Since(start)

	if rtn == nil {
		return returncode, duration
	}

	if msg, ok := rtn.(*exec.ExitError); ok { // there is error code
		log.Printf("Command finished with error: %v", rtn)
		returncode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	return returncode, duration
}

// Run a command
func Run(cmd *exec.Cmd) int {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	user, _ := user.Current()
	host := fqdn.Get()

	logger, _ := fluent.New(fluent.Config{})
	defer logger.Close()

	if logger != nil {
		_ = logger.Post("cron.start", map[string]string{
			"User":    user.Username,
			"Command": fmt.Sprintf("%v", cmd.Args),
			"Host":    host,
		})
	}

	var bufout bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(&bufout, os.Stdout)
	cmd.Stderr = io.MultiWriter(&bufout, os.Stderr)

	go func() {
		s := <-c
		if cmd.ProcessState == nil {
			cmd.Process.Signal(s)
		}
	}()

	returncode, duration := Wrap(cmd)

	if returncode == 0 {
		if logger != nil {
			_ = logger.Post("cron.complete", map[string]string{
				"User":       user.Username,
				"Command":    fmt.Sprintf("%v", cmd.Args),
				"Returncode": fmt.Sprintf("%d", returncode),
				"Host":       host,
				"Duration":   fmt.Sprintf("%f", duration.Seconds()),
				//"Output":     bufout.String(),
			})
		}
	} else {
		if logger != nil {
			_ = logger.Post("cron.error", map[string]string{
				"User":       user.Username,
				"Command":    fmt.Sprintf("%v", cmd.Args),
				"Returncode": fmt.Sprintf("%d", returncode),
				"Host":       host,
				"Duration":   fmt.Sprintf("%f", duration.Seconds()),
			})
		}
		raven.CaptureMessageAndWait(cmd.Path, map[string]string{
			"User":       user.Username,
			"Command":    fmt.Sprintf("%v", cmd.Args),
			"Returncode": fmt.Sprintf("%d", returncode),
			"Host":       host,
		})
	}

	return returncode
}
