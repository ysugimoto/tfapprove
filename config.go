package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configFileName = ".tfapprove.toml"
)

// TOML configuration struct
// Some field values are injected via .tfapprove.toml
type Config struct {
	Server  Server
	Approve Approve
	Command Command

	// Stack CLI arguments
	args []string `toml:"-"`
}

// Server struct
type Server struct {
	// URL always use fixed value so the user could not change this field.
	Url    string
	ApiKey string `toml:"api_key"`
}

// Command setting struct
type Command struct {
	TerraformCommandPath string `toml:"terraform"`
}

// Approvement configuration struct
type Approve struct {
	SlackChannel  string `toml:"slack_channel"`
	NeedApprovers int    `toml:"need_approvers"`
	WaitTimeout   int    `toml:"wait_timeout"`
}

// defaultConfig() generates default configuration with default value.
func defaultConfig() Config {
	return Config{
		Server: Server{
			Url:    aggregateServer,
			ApiKey: "",
		},
		Command: Command{
			TerraformCommandPath: "terraform",
		},
		Approve: Approve{
			NeedApprovers: 1,
			WaitTimeout:   1,
		},
	}
}

// Create new configuration pointer.
// Find .tfapprove.toml file on current directory and decode it
func newConfig() (*Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	file := filepath.Join(pwd, configFileName)

	var c Config = defaultConfig()
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"-h"}
	}
	c.args = args

	if _, err := toml.DecodeFile(file, &c); err != nil {
		if _, ok := err.(*fs.PathError); ok {
			return &c, nil
		}
		return nil, err
	}

	if v := os.Getenv("TFAPPROVE_API_KEY"); v != "" {
		c.Server.ApiKey = v
	}
	return &c, nil
}

// Check apply command because we need to wrap apply command
func (c *Config) IsApply() bool {
	for i := range c.args {
		if c.args[i] == "apply" {
			return true
		}
	}
	return false
}

// Check generate command because generate subcommand is only enable on this command
func (c *Config) IsGenerate() bool {
	for i := range c.args {
		if c.args[i] == "generate" {
			return true
		}
	}
	return false
}

// Check version subcommand
func (c *Config) IsVersion() bool {
	for i := range c.args {
		if c.args[i] == "version" {
			return true
		}
	}
	return false
}

var configTemplate = `### Server configuration
[Server]
  # API Key is needed for communicating with application server.
  # For secret reason, you can speficy this value via envrionment variable of "TFAPPROVE_API_KEY".
  api_key = ""

### Approval configuration
[Approve]
  # Slack channel ID like CXXXXXXXX that post approval message to.
  slack_channel = ""

  # Minimum approvers to continue apply.
  need_approvers = 1

  # Maximum wait time to get approval (minute order)
  wait_timeout = 1

### Terraform command setting
[Command]
  # Specify "terraform" command path
  terraform = "terraform"
`

// Generate configuration file
func generateConfig() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	file := filepath.Join(pwd, configFileName)
	if _, err := os.Stat(file); err == nil {
		return errors.New("Configuration file already exists")
	}

	fp, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err := io.WriteString(fp, configTemplate); err != nil {
		return err
	}
	return nil
}
