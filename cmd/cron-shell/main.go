package main

import (
	"os"
	"os/exec"

	"github.com/kfdm/cron-me/internal"
)

func main() {
	// Append /bin/bash to the start of our command so that it will actually run
	cmd := exec.Command("/bin/bash", os.Args[1:]...)

	os.Exit(internal.Run(cmd))
}
