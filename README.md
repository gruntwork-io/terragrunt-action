# terragrunt-action

A GitHub Action for installing and running Terragrunt and OpenTofu.

## Minimum Supported Version

The current minimum supported version of Terragrunt is 0.77.22.

## Inputs

Supported GitHub action inputs:

| Input Name     | Description                                                        | Required                                                                    | Example values      |
|:---------------|:-------------------------------------------------------------------|:---------------------------------------------------------------------------:|:-------------------:|
| tg_version     | Terragrunt version to be used in Action execution                  | `true` if no `mise.toml` file present                                       |     0.50.8          |
| tofu_version   | OpenTofu version to be used in Action execution                    | `true` if `tf_path` is not provided and the file `mise.toml` is not present |     1.6.0           |
| tf_path        | Path to Terraform binary (use to explicitly choose tofu/terraform) | `false`                                                                     | /usr/bin/tofu       |
| tg_dir         | Directory in which Terragrunt will be invoked                      | `false`                                                                     |      work           |
| tg_command     | Terragrunt command to execute                                      | `false`                                                                     |   plan/apply        |
| tg_comment     | Add comment to Pull request with execution output                  | `false`                                                                     |      0/1            |
| tg_add_approve | Automatically add "-auto-approve" to commands, enabled by default  | `false`                                                                     |      0/1            |
| github_token   | GitHub token for API authentication to avoid rate limits           | `false`                                                                     | ${{ github.token }} |

## Tool Version Management

This action supports two ways to specify tool versions:

1. **Using `mise.toml` file** (recommended): Create a `mise.toml` or `.mise.toml` file in your repository root to configure [mise](https://github.com/jdx/mise):

   ```toml
   [tools]
   terragrunt = "0.82.3"
   opentofu = "1.10.1"
   ```

2. **Using action inputs**: Specify versions directly in the action inputs when no `mise.toml` file is present.

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

## Usage

### Example 1: As an installation action (without command execution)

You can use this action to simply install Terragrunt and OpenTofu/Terraform, then run terragrunt commands in subsequent steps:

```yaml
name: 'Terragrunt GitHub Actions'
on:
  - pull_request

jobs:
  terragrunt:
    runs-on: ubuntu-latest
    steps:
      - name: 'Checkout'
        uses: actions/checkout@v4

      # Assuming you have a mise.toml file in your repository.
      - name: Install Terragrunt and OpenTofu
        uses: gruntwork-io/terragrunt-action@v3
        # Note: no tg_command specified, so terragrunt won't be executed

      # If you don't have a mise.toml file in your repository, you can specify tool version directly.
      # - name: Install Terragrunt and OpenTofu
      #   uses: gruntwork-io/terragrunt-action@v3
      #   with:
      #     tg_version: '0.82.2'
      #     tofu_version: '1.10.1'

      - name: Run terragrunt plan
        run: |
          cd infrastructure/
          terragrunt plan

      - name: Run terragrunt apply
        if: github.ref == 'refs/heads/main'
        run: |
          cd infrastructure/
          terragrunt run --all --non-interactive apply
```

### Example 2: Using mise.toml file (recommended)

Create a `mise.toml` file in your repository root:

```toml
[tools]
terragrunt = "0.82.2"
opentofu = "1.10.1"
```

Then use the action without specifying versions:

```yaml
name: 'Terragrunt GitHub Actions'
on:
  - pull_request

env:
  working_dir: 'project'

jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - name: 'Checkout'
        uses: actions/checkout@v4

      - name: Check terragrunt HCL
        uses: gruntwork-io/terragrunt-action@v3
        with:
          tg_dir: ${{ env.working_dir }}
          tg_command: 'hcl fmt --check --diff'

  plan:
    runs-on: ubuntu-latest
    needs: [ checks ]
    steps:
      - name: 'Checkout'
        uses: actions/checkout@v4

      - name: Plan
        uses: gruntwork-io/terragrunt-action@v3
        with:
          tg_dir: ${{ env.working_dir }}
          tg_command: 'plan'

  deploy:
    runs-on: ubuntu-latest
    needs: [ plan ]
    environment: 'prod'
    if: github.ref == 'refs/heads/main'
    steps:
      - name: 'Checkout'
        uses: actions/checkout@v4

      - name: Deploy
        uses: gruntwork-io/terragrunt-action@v3
        with:
          tg_dir: ${{ env.working_dir }}
          tg_command: 'apply'
```

### Example 2: Using action inputs for tool versions

```yaml
name: 'Terragrunt GitHub Actions'
on:
  - pull_request

env:
  tg_version: '0.82.2'
  tofu_version: '1.10.1'
  working_dir: 'project'

jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - name: 'Checkout'
        uses: actions/checkout@v4

      - name: Check terragrunt HCL
        uses: gruntwork-io/terragrunt-action@v3
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'hcl fmt --check --diff'

  plan:
    runs-on: ubuntu-latest
    needs: [ checks ]
    steps:
      - name: 'Checkout'
        uses: actions/checkout@main

      - name: Plan
        uses: gruntwork-io/terragrunt-action@v3
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
        uses: actions/checkout@v4

      - name: Deploy
        uses: gruntwork-io/terragrunt-action@v3
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'apply'
```

### Example 4: Using tf_path to specify Terraform binary

```yaml
- name: Deploy with explicit terraform binary
  uses: gruntwork-io/terragrunt-action@v3
  with:
    tf_path: "terraform"  # Explicitly use terraform even if both tofu and terraform are installed.
    tg_dir: ${{ env.working_dir }}
    tg_command: 'apply'
```

### Example 5: Passing custom code before running Terragrunt

```yaml
...
- name: Plan
  uses: gruntwork-io/terragrunt-action@v3
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

### Example 6: Using GitHub cache for OpenTofu plugins (providers)

```yaml
...
env:
  tg_version: 0.82.2
  tofu_version: 1.10.1
  working_dir: project
  TF_PLUGIN_CACHE_DIR: ${{ github.workspace }}/.terraform.d/plugin-cache

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Create OpenTofu Plugin Cache Dir
        run: mkdir -p $TF_PLUGIN_CACHE_DIR

      - name: OpenTofu Plugin Cache
        uses: actions/cache@v4
        with:
          path: ${{ env.TF_PLUGIN_CACHE_DIR }}
          key: ${{ runner.os }}-terraform-plugin-cache-${{ hashFiles('**/.terraform.lock.hcl') }}

      - name: Plan
        uses: gruntwork-io/terragrunt-action@v3
        env:
          TF_PLUGIN_CACHE_DIR: ${{ env.TF_PLUGIN_CACHE_DIR }}
        with:
          tofu_version: ${{ env.tofu_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: plan
...
```
