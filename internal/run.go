package internal

import (
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/kfdm/cron-me/internal/logging"
)

// signalWatcher handles our signals
// https://github.com/Netflix/signal-wrapper/blob/master/main.go#L25
func signalWatcher(cmd *exec.Cmd, logger log.Logger) {
	signalChan := make(chan os.Signal, 100)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	signal := <-signalChan

	if err := cmd.Process.Signal(signal); err != nil {
		logger.Log("msg", "Unable to forward signal", "err", err)
	}

	for signal = range signalChan {
		logger.Log("msg", "Forwarding signal", "signal", signal)
		if err := cmd.Process.Signal(signal); err != nil {
			logger.Log("msg", "Unable to forward signal", "err", err)
		}
	}
}

// WrapReturn returns proper error code
func WrapReturn(cmd *exec.Cmd) int {
	returncode := 0
	rtn := cmd.Run()

	if rtn == nil {
		return returncode
	}

	if msg, ok := rtn.(*exec.ExitError); ok { // there is error code
		// log.Printf("Command finished with error: %v", rtn)
		returncode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return returncode
}

// Run a command
func Run(cmd *exec.Cmd) int {

	logger := logging.NewFluentLogger()

	user, _ := user.Current()

	logger.Log("tag", "cron.start", "User", user.Username, "Command", cmd.Args)

	go signalWatcher(cmd, logger)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	returncode := WrapReturn(cmd)
	duration := time.Since(start)

	if returncode == 0 {
		logger.Log("tag", "cron.complete", "User", user.Username, "Command", cmd.Args, "Returncode", returncode, "Duration", duration.Seconds())
	} else {
		logger.Log("tag", "cron.error", "User", user.Username, "Command", cmd.Args, "Returncode", returncode, "Duration", duration.Seconds())
		logging.Sentry(user, cmd, returncode)
	}

	return returncode
}
