package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server  Server
	Approve Approve
	Command Command

	args []string `toml:"-"`
}

type Server struct {
	Url    string `toml:"url"`
	ApiKey string `toml:"api_key"`
}

type Command struct {
	TerraformCommandPath string `toml:"terraform"`
}

type Approve struct {
	SlackChannel  string `toml:"slack_channel"`
	NeedApprovers int    `toml:"need_approvers"`
	WaitTimeout   int    `toml:"wait_timeout"`
}

func defaultConfig() Config {
	return Config{
		Server: Server{
			Url:    "wss://tfapprove.com",
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

func newConfig() (*Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	file := filepath.Join(pwd, ".tfapprove.toml")

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

func (c *Config) IsApply() bool {
	for i := range c.args {
		if c.args[i] == "apply" {
			return true
		}
	}
	return false
}

func (c *Config) IsGenerate() bool {
	for i := range c.args {
		if c.args[i] == "generate" {
			return true
		}
	}
	return false
}

var configTemplate = `### Server configuration
[Server]
  # Default url indicates shared application server.
  # If you make own application server, change this field.
  url = "wss://tfapprove.com"

  # API Key is needed for communicate with application server.
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

func generateConfig() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	file := filepath.Join(pwd, ".tfapprove.toml")
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
