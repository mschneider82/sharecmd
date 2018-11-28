package main

import (
	"log"
	"os/exec"
	"runtime"
)

func toClip(output string) {
	var cmd *exec.Cmd
	switch os := runtime.GOOS; os {
	case "linux":
		cmd = exec.Command("xclip", "-selection", "c")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		// not yet supported os:
		// freebsd, openbsd,
		// plan9, windows...
		return
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Can't copy link to clipboard: %s", err.Error())
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Can't copy link to clipboard: %s", err.Error())
	}

	if _, err := in.Write([]byte(output)); err != nil {
		log.Fatalf("Can't copy link to clipboard: %s", err.Error())
	}

	if err := in.Close(); err != nil {
		log.Fatalf("Can't copy link to clipboard: %s", err.Error())
	}

	cmd.Wait()
	log.Println("URL copied to clipboard!")
}
