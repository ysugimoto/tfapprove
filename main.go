package main

import (
	"io"
	"os"
	"os/exec"
)

// These values are injected via compilation time
var aggregateServer string = ""
var version string = "dev"

func main() {
	if err := _main(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			os.Exit(e.ExitCode())
		} else {
			_, _ = io.WriteString(os.Stderr, err.Error())
		}
	}
}

func _main() error {
	c, err := newConfig()
	if err != nil {
		return err
	}

	switch {
	case c.IsApply():
		return wrapTerraformApply(c)
	case c.IsGenerate():
		if err := generateConfig(); err != nil {
			return err
		}
		_, _ = io.WriteString(os.Stdout, "Configuration file has generated!\n")
		return nil
	case c.IsVersion():
		_, _ = io.WriteString(os.Stdout, "TFApprove  "+version+"\n")
		_, _ = io.WriteString(os.Stdout, "---------\n")
		fallthrough
	default:
		// Simply pass arguments to the "terraform" command
		cmd := exec.Command(c.Command.TerraformCommandPath, c.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		return cmd.Run()
	}
}
