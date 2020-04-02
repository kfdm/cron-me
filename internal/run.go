package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/ShowMax/go-fqdn"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/getsentry/sentry-go"
)

// signalWatcher handles our signals
// https://github.com/Netflix/signal-wrapper/blob/master/main.go#L25
func signalWatcher(ctx context.Context, cmd *exec.Cmd) {
	signalChan := make(chan os.Signal, 100)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	signal := <-signalChan

	if err := cmd.Process.Signal(signal); err != nil {
		log.Printf("Unable to forward signal: %v", err)
	}

	for signal = range signalChan {
		log.Printf("Forwarding signal: %v", signal)
		if err := cmd.Process.Signal(signal); err != nil {
			log.Printf("Unable to forward signal %v:", err)
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
		log.Printf("Command finished with error: %v", rtn)
		returncode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return returncode
}

// Run a command
func Run(cmd *exec.Cmd) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, _ := user.Current()
	host := fqdn.Get()

	logger, _ := fluent.New(fluent.Config{Async: true})
	defer logger.Close()

	if logger != nil {
		_ = logger.Post("cron.start", map[string]string{
			"User":    user.Username,
			"Command": fmt.Sprintf("%v", cmd.Args),
			"Host":    host,
			"Test":    cmd.String(),
		})
	}

	// var bufout bytes.Buffer
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = io.MultiWriter(&bufout, os.Stdout)
	// cmd.Stderr = io.MultiWriter(&bufout, os.Stderr)

	go signalWatcher(ctx, cmd)

	start := time.Now()
	returncode := WrapReturn(cmd)
	duration := time.Since(start)

	if returncode == 0 {
		if logger != nil {
			_ = logger.Post("cron.complete", map[string]string{
				"User":       user.Username,
				"Command":    fmt.Sprintf("%v", cmd.Args),
				"Returncode": fmt.Sprintf("%d", returncode),
				"Host":       host,
				"Duration":   fmt.Sprintf("%f", duration.Seconds()),
				// "Output":     bufout.String(),
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

		err := sentry.Init(sentry.ClientOptions{})
		if err == nil {
			defer sentry.Flush(2 * time.Second)
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelWarning)
				scope.SetUser(sentry.User{Username: user.Username, ID: user.Uid})
				scope.SetExtra("Returncode", returncode)
				scope.SetExtra("Command", fmt.Sprintf("%v", cmd.Args))
			})
			sentry.CaptureMessage(cmd.String())
		}
	}

	return returncode
}
