package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var browserCommands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

func OpenBrowser(uri string) {
	run, ok := browserCommands[runtime.GOOS]
	if !ok {
		fmt.Fprintf(os.Stdout, "don't know how to open things on %s platform\n", runtime.GOOS)
		fmt.Fprintf(os.Stdout, "Click this link to authorize repository access: %s\n", uri)
	}
	cmd := exec.Command(run, uri)
	cmd.Start()
}
