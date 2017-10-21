package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ShowMax/go-fqdn"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/getsentry/raven-go"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Remove 'cron-me' from args
	_, a := os.Args[0], os.Args[1:]

	val, _ := os.LookupEnv("SHELL")
	if val == os.Args[0] {
		a = os.Args
		a[0] = "/bin/sh"
	}

	if len(a) == 0 {
		fmt.Printf("No arguments?")
		os.Exit(1)
	}

	user, _ := user.Current()
	host := fqdn.Get()

	logger, _ := fluent.New(fluent.Config{})
	defer logger.Close()

	cmd := &exec.Cmd{
		Path: a[0],
		Args: a,
	}
	if filepath.Base(a[0]) == a[0] {
		if lp, err := exec.LookPath(a[0]); err != nil {
			//cmd.lookPathErr = err
		} else {
			cmd.Path = lp
		}
	}

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

	start := time.Now()
	rtn := cmd.Run()
	duration := time.Since(start)

	returncode := 0
	if rtn != nil {
		if msg, ok := rtn.(*exec.ExitError); ok { // there is error code
			log.Printf("Command finished with error: %v", rtn)
			returncode = msg.Sys().(syscall.WaitStatus).ExitStatus()
		}

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
	} else {
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
	}
	os.Exit(returncode)
}
