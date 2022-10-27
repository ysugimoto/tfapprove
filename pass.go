package main

import (
	"os"
	"os/exec"
)

func passTerraform(c *Config) error {
	cmd := exec.Command(c.Command.TerraformCommandPath, c.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
