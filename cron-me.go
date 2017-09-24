package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"os/exec"

	"github.com/getsentry/raven-go"
	"github.com/fluent/fluent-logger-golang/fluent"
)

func main() {
	user, _ := user.Current()

	logger, _ := fluent.New(fluent.Config{})
	defer logger.Close()

	cmd := exec.Command(os.Args[1])

	if logger != nil {
		_ = logger.Post("cron.start", map[string]string{
			"Path":  cmd.Path,
			"User": user.Username,
		})
	}

	var bufout bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(&bufout, os.Stdout)
	cmd.Stderr = io.MultiWriter(&bufout, os.Stderr)

	err := cmd.Run()
	if err != nil {
		if logger != nil {
			_ = logger.Post("cron.error", map[string]string{
		    	"Path":  cmd.Path,
				"User": user.Username,
			})
		}
		raven.CaptureMessageAndWait(cmd.Path, map[string]string{
			"Args": fmt.Sprintf("%v", cmd.Args),
			"User": user.Username,
		})
		log.Fatal(err)
	} else {
		if logger != nil {
			_ = logger.Post("cron.complete", map[string]string{
				"Path":  cmd.Path,
				"User": user.Username,
				"Output": bufout.String(),
			})
		}
	}
}
