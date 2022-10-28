# tfapprove

`tfapprove` is a terrraform wrapper tool that requires approvals before run `terraform apply`.

### Requirements

- [Terraform](https://www.terraform.io/)

## Hacking `terraform apply` subcommand

To apply your infrastructure change, usually you can type `terraform apply` command. Then you can use `tfapprove apply` instead.
`tfapprove apply` command prevents user confirmation of typing `yes` or `no` to apply changes, but wait for other people's approval on Slack.
The other subcommands (e.g plan, init, etc...) runs as `terraform` subcommands.

Entirely, you can use `tfapprove` command instead of `terraform` command.

## Installation

Press `Install to Slack` button on `README.md` and install tfqapprove Slack App to your Slack workspace.
After installed, Slack app will tell you _API Key_ on DM, you need to remmember its value.

### Configuration

`tfapprove` wants configuration file names `.tfapprove.toml`, you can generate skeleton configuration via `tfapprove generate`.

The configuration example is following

```toml
[Server]
  # API Key is needed for communicating with application server.
  # For secret reason, you can speficy this value via envrionment variable of "TFAPPROVE_API_KEY".
  api_key = ""

[Approve]
  # Slack channel ID like CXXXXXXXX format that post approval message to.
  slack_channel = ""

  # Minimum approvers to continue apply.
  need_approvers = 1

  # Maximum wait time to get approval (minute order)
  wait_timeout = 1

[Command]
  # Specify "terraform" command path
  terraform = "terraform"
```

We describe as following table:

| Section | name           | required | description                                                    |
|:-------:|:--------------:|:--------:|:---------------------------------------------------------------|
| Server  | -              | yes      | Server Setting                                                 |
|         | api_key        | yes      | Authentication Key, Put API Key which you got on installation  |
| Approve | -              | yes      | Approve Setting                                                |
|         | slacK_channel  | yes      | Slack channel ID that post approval request message            |
|         | need_approvers | no       | Minimum approvers count, default is `1`                        |
|         | wait_timeout   | no       | Timeout minutes to give up approval comes , default is 1       |
| Command | -              | yes      | Command Setting                                                |
|         | terraform      | yes      | `terraform` command path                                       |

