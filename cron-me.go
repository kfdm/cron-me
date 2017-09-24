package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/getsentry/raven-go"
)

func main() {
	cmd := exec.Command(os.Args[1])

	var bufout bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(&bufout, os.Stdout)
	cmd.Stderr = io.MultiWriter(&bufout, os.Stderr)

	err := cmd.Run()
	if err != nil {
		raven.CaptureMessageAndWait(cmd.Path, map[string]string{"Args": fmt.Sprintf("%v", cmd.Args)})
		log.Fatal(err)
	}
}
