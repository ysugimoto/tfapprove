package main

import (
	"io"
	"os"
	"os/exec"
)

func main() {
	if err := _main(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			os.Exit(e.ExitCode())
		} else {
			io.WriteString(os.Stderr, err.Error())
		}
	}
}

func _main() error {
	c, err := newConfig()
	if err != nil {
		return err
	}

	if c.IsApply() {
		return wrapTerraformApply(c)
	} else if c.IsGenerate() {
		if err := generateConfig(); err != nil {
			return err
		}
		io.WriteString(os.Stdout, "Configuration file has generated!\n")
		return nil
	} else {
		return passTerraform(c)
	}
}
