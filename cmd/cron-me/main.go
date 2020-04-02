package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kfdm/cron-me/command"
)

func main() {
	// Remove 'cron-me' from args
	_, a := os.Args[0], os.Args[1:]
	cmd := &exec.Cmd{
		Path: a[0],
		Args: a,
	}

	if filepath.Base(cmd.Args[0]) == cmd.Args[0] {
		if lp, err := exec.LookPath(cmd.Args[0]); err != nil {
			//cmd.lookPathErr = err
		} else {
			cmd.Path = lp
		}
	}

	os.Exit(command.Run(cmd))
}
