package internal

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestRun(t *testing.T) {
	cmd := exec.Command("false")
	rtncode := Run(cmd)
	fmt.Print(rtncode)
}
