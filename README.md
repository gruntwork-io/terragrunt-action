# terragrunt-action

A GitHub Action for installing and running Terragrunt

## Inputs

Supported GitHub action inputs:

| Input Name     | Description                                                       | Required                                  | Example values |
|:---------------|:------------------------------------------------------------------|:-----------------------------------------:|:--------------:|
| tf_version     | Terraform version to be used in Action execution                  | `true` if `tofu_version` is not supplied |     1.4.6      | 
| tofu_version   | OpenTofu version to be used in Action execution                   | `true` if `tf_version` is not supplied   |     1.6.0      |
| tg_version     | Terragrunt version to be user in Action execution                 | `true`                                   |     0.50.8     |
| tg_dir         | Directory in which Terragrunt will be invoked                     | `false`                                  |      work      |
| tg_command     | Terragrunt command to execute                                     | `true`                                   |   plan/apply   |
| tg_comment     | Add comment to Pull request with execution output                 | `false`                                  |      0/1       |
| tg_add_approve | Automatically add "-auto-approve" to commands, enabled by default | `false`                                  |      0/1       |

## Environment Variables

Supported environment variables:

| Input Name             | Description                                                                                                  | 
|:-----------------------|:-------------------------------------------------------------------------------------------------------------|
| GITHUB_TOKEN           | GitHub token used to add comment to Pull request                                                             |
| TF_LOG                 | Log level for Terraform                                                                                      |
| TF_VAR_name            | Define custom variable name as inputs                                                                        |
| INPUT_PRE_EXEC_number  | Environment variable is utilized to provide custom commands that will be executed before running Terragrunt  |
| INPUT_POST_EXEC_number | Environment variable is utilized to provide custom commands that will be executed *after* running Terragrunt |

## Outputs

Outputs of GitHub action:

| Input Name          | Description                     |
|:--------------------|:--------------------------------|
| tg_action_exit_code | Terragrunt exit code            |
| tg_action_output    | Terragrunt output as plain text |
| tg_output           | Terragrunt output in JSON format|

## Usage

Example definition of Github Action workflow:

```yaml
name: 'Terragrunt GitHub Actions'
on:
  - pull_request

env:
  tofu_version: '1.8.1'
  tg_version: '0.67.0'
  working_dir: 'project'

jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - name: 'Checkout'
        uses: actions/checkout@main

      - name: Check terragrunt HCL
        uses: gruntwork-io/terragrunt-action@v2
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'hclfmt --terragrunt-check --terragrunt-diff'

  plan:
    runs-on: ubuntu-latest
    needs: [ checks ]
    steps:
      - name: 'Checkout'
        uses: actions/checkout@main

      - name: Plan
        uses: gruntwork-io/terragrunt-action@v2
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'plan'

  deploy:
    runs-on: ubuntu-latest
    needs: [ plan ]
    environment: 'prod'
    if: github.ref == 'refs/heads/main'
    steps:
      - name: 'Checkout'
        uses: actions/checkout@main

      - name: Deploy
        uses: gruntwork-io/terragrunt-action@v2
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'apply'
```

Example of passing custom code before running Terragrunt:

```yaml
...
- name: Plan
  uses: gruntwork-io/terragrunt-action@v2
  env:
    # configure git to use custom token to clone repository.
    INPUT_PRE_EXEC_1: |
      git config --global url."https://user:${{secrets.PAT_TOKEN}}@github.com".insteadOf "https://github.com"
    # print git configuration
    INPUT_PRE_EXEC_2: |
      git config --global --list
  with:
    tg_command: 'plan'
...
```

Example of using GitHub cache for Terraform plugins (providers):

```yaml
...
env:
  tofu_version: 1.5.7
  tg_version: 0.53.2
  working_dir: project
  TF_PLUGIN_CACHE_DIR: ${{ github.workspace }}/.terraform.d/plugin-cache

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@main

      - name: Create Terraform Plugin Cache Dir
        run: mkdir -p $TF_PLUGIN_CACHE_DIR

      - name: Terraform Plugin Cache
        uses: actions/cache@v4.0.1
        with:
          path: ${{ env.TF_PLUGIN_CACHE_DIR }}
          key: ${{ runner.os }}-terraform-plugin-cache-${{ hashFiles('**/.terraform.lock.hcl') }}

      - name: Plan
        uses: gruntwork-io/terragrunt-action@v2
        env:
          TF_PLUGIN_CACHE_DIR: /github/workspace/.terraform.d/plugin-cache
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: plan
...
```
