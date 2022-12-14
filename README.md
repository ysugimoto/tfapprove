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

Press `Add to Slack` button below and install Slack App to your Slack workspace.

<a href="https://slack.com/oauth/v2/authorize?client_id=1860443096256.4277783553521&scope=chat:write,files:write,im:write&user_scope=" target="_blank" rel="noreferrer noopener">
  <img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
  </a>

After installed, Slack app will tell you _API Key_ on DM, you need to remember its value.

<img width="586" alt="Screen Shot 2022-10-28 at 21 28 47" src="https://user-images.githubusercontent.com/1000401/198689264-9978ccb3-08e5-46b9-a727-6afe5252f976.png">

### Configuration

`tfapprove` wants configuration file names `.tfapprove.toml`, you can generate skeleton configuration via `tfapprove generate`.

The configuration example is following

```toml
[Server]
  # API Key is needed for communicating with application server.
  # For secret reason, you can speficy this value via envrionment variable of "TFAPPROVE_API_KEY".
  api_key = ""

[Approve]
  # Slack channel ID or channel name that send approval message to.
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

### Wait for approvals

Once you run `tfapprove apply` with valid configurations, the command waits for other member's approval from Slack, and show plan result to the thread message.

<img width="493" alt="Screen Shot 2022-10-24 at 22 34 16" src="https://user-images.githubusercontent.com/1000401/198689311-e208c123-d29f-4bc2-82c2-a6bf155c4c26.png">

Then members can choose appove or reject this plan, press the button which they want to do.
After member has approved, terraform apply will continue.
