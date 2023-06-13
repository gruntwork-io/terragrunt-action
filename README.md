# terragrunt-action
A GitHub Action for installing and running Terragrunt

## Inputs

## Outputs

## Environment Variables

## Usage

Example definition of Github Action workflow:
```yaml
name: 'Terragrunt GitHub Actions'
on:
  - pull_request

env:
  tf_version: '1.4.6'
  tg_version: '0.46.3'
  working_dir: 'project'

jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - name: 'Checkout'
        uses: actions/checkout@master

      - name: Check terragrunt HCL
        uses: gruntwork-io/terragrunt-action@v1
        with:
          tf_version: ${{ env.tf_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'hclfmt --terragrunt-check --terragrunt-diff'

  plan:
    runs-on: ubuntu-latest
    needs: [checks]
    steps:
      - name: 'Checkout'
        uses: actions/checkout@master

      - name: Plan
        uses: gruntwork-io/terragrunt-action@v1
        with:
          tf_version: ${{ env.tf_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'plan'

  deploy:
    runs-on: ubuntu-latest
    needs: [plan]
    environment: 'prod'
    if: github.ref == 'refs/heads/master'
    steps:
      - name: 'Checkout'
        uses: actions/checkout@master

      - name: Deploy
        uses: gruntwork-io/terragrunt-action@v1
        with:
          tf_version: ${{ env.tf_version }}
          tg_version: ${{ env.tg_version }}
          tg_dir: ${{ env.working_dir }}
          tg_command: 'apply'
```

