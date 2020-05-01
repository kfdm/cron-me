package main

import (
	"os"
	"os/exec"

	"github.com/kfdm/cron-me/command"
)

func main() {
	// Append /bin/bash to the start of our command so that it will actually run
	cmd := exec.Command("/bin/sh", os.Args[1:]...)

	os.Exit(command.Run(cmd))
}
