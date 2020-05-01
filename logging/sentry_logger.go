package logging

import (
	"os/exec"
	"os/user"
	"time"

	"github.com/getsentry/sentry-go"
)

// Sentry to send to fluentd
func Sentry(user *user.User, cmd *exec.Cmd, rtn int) *sentry.EventID {
	err := sentry.Init(sentry.ClientOptions{})
	if err == nil {
		defer sentry.Flush(2 * time.Second)
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelWarning)
			scope.SetUser(sentry.User{Username: user.Username, ID: user.Uid})
			scope.SetExtra("Returncode", rtn)
		})
		return sentry.CaptureMessage(cmd.String())
	}
	return nil
}
