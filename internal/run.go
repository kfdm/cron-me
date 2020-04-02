package internal

import (
	"context"

	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/ShowMax/go-fqdn"
	"github.com/getsentry/sentry-go"
	"github.com/go-kit/kit/log"
	"github.com/kfdm/cron-me/internal/logger"
)

// signalWatcher handles our signals
// https://github.com/Netflix/signal-wrapper/blob/master/main.go#L25
func signalWatcher(ctx context.Context, cmd *exec.Cmd, logger log.Logger) {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logger.NewFluentLogger()

	user, _ := user.Current()
	host := fqdn.Get()

	logger.Log("tag", "cron.start", "User", user.Username, "Command", cmd.Args, "Host", host)

	go signalWatcher(ctx, cmd, logger)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	returncode := WrapReturn(cmd)
	duration := time.Since(start)

	if returncode == 0 {
		logger.Log("tag", "cron.complete", "User", user.Username, "Command", cmd.Args, "Returncode", returncode, "Host", host, "Duration", duration.Seconds())
	} else {
		logger.Log("tag", "cron.error", "User", user.Username, "Command", cmd.Args, "Returncode", returncode, "Host", host, "Duration", duration.Seconds())

		err := sentry.Init(sentry.ClientOptions{})
		if err == nil {
			defer sentry.Flush(2 * time.Second)
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelWarning)
				scope.SetUser(sentry.User{Username: user.Username, ID: user.Uid})
				scope.SetExtra("Returncode", returncode)
			})
			sentry.CaptureMessage(cmd.String())
		}
	}

	return returncode
}
