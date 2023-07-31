#!/usr/bin/env bash
set -e

# write log message with timestamp
function log {
  local -r message="$1"
  local -r timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  >&2 echo -e "${timestamp} ${message}"
}

# remove ANSI color codes from argument variable
function clean_colors {
  local -r input="$1"
  echo "${input}" | sed -E 's/\x1B\[[0-9;]*[mGK]//g'
}

# clean multiline text to be passed to Github API
function clean_multiline_text {
  local -r input="$1"
  local output
  output="${input//'%'/'%25'}"
  output="${output//$'\n'/'%0A'}"
  output="${output//$'\r'/'%0D'}"
  echo "${output}"
}

# configure custom git authentication token based on provided private/internal org/repo path (supports GitHub App or PAT)
function configurePrivateGitPath(){
  if [ -n "${INPUT_GITHUB_TOKEN}" ] && [ -n "${INPUT_GITHUB_PRIVATE_PATH}" ]; then
    if [ -n "${INPUT_GITHUB_AUTH_TYPE}" ] && [ "${INPUT_GITHUB_AUTH_TYPE}" =~ "app" ]; then
      git config --global url."https://x-access-token:${INPUT_GITHUB_TOKEN}@${INPUT_GITHUB_PRIVATE_PATH}".insteadOf "https://${INPUT_GITHUB_PRIVATE_PATH}"
    fi
  fi
}

# install and switch particular terraform version
function install_terraform {
  local -r version="$1"
  if [[ "${version}" == "none" ]]; then
    return
  fi
  tfenv install "${version}"
  tfenv use "${version}"
}

# install passed terragrunt version
function install_terragrunt {
  local -r version="$1"
  if [[ "${version}" == "none" ]]; then
    return
  fi
  TG_VERSION="${version}" tgswitch
}

# run terragrunt commands in specified directory
# arguments: directory and terragrunt command
# output variables:
# terragrunt_log_file path to log file
# terragrunt_exit_code exit code of terragrunt command
function run_terragrunt {
  local -r dir="$1"
  local -r command=($2)

  # terragrunt_log_file can be used later as file with execution output
  terragrunt_log_file=$(mktemp)

  cd "${dir}"
  terragrunt "${command[@]}" 2>&1 | tee "${terragrunt_log_file}"
  # terragrunt_exit_code can be used later to determine if execution was successful
  terragrunt_exit_code=${PIPESTATUS[0]}
}

# post comment to pull request
function comment {
  local -r message="$1"
  local comment_url
  comment_url=$(jq -r '.pull_request.comments_url' "$GITHUB_EVENT_PATH")
  # may be getting called from something like branch deploy
  if [[ "${comment_url}" == "" || "${comment_url}" == "null" ]]; then
    comment_url=$(jq -r '.issue.comments_url' "$GITHUB_EVENT_PATH")
  fi
  if [[ "${comment_url}" == "" || "${comment_url}" == "null" ]]; then
    log "Skipping comment as there is not comment url"
    return
  fi
  local messagePayload
  messagePayload=$(jq -n --arg body "$message" '{ "body": $body }')
  curl -s -S -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" -d "$messagePayload" "$comment_url"
}

function setup_git {
  # Avoid git permissions warnings
  git config --global --add safe.directory /github/workspace
}

# Run INPUT_PRE_EXEC_* environment variables as Bash code
function setup_pre_exec {
  # Get all environment variables that match the pattern INPUT_PRE_EXEC_*
  local -r pre_exec_vars=$(env | grep -o '^INPUT_PRE_EXEC_[0-9]\+' | sort)
  # Loop through each pre-execution variable and execute its value (Bash code)
  local pre_exec_command
  while IFS= read -r pre_exec_var; do
    if [[ -n "${pre_exec_var}" ]]; then
      log "Evaluating ${pre_exec_var}"
      pre_exec_command="${!pre_exec_var}"
      eval "$pre_exec_command"
    fi
  done <<< "$pre_exec_vars"
}

function main {
  log "Starting Terragrunt Action"
  trap 'log "Finished Terragrunt Action execution"' EXIT
  local -r tf_version=${INPUT_TF_VERSION}
  local -r tg_version=${INPUT_TG_VERSION}
  local -r tg_command=${INPUT_TG_COMMAND}
  local -r tg_comment=${INPUT_TG_COMMENT:-0}
  local -r tg_dir=${INPUT_TG_DIR:-.}

  if [[ -z "${tf_version}" ]]; then
    log "tf_version is not set"
    exit 1
  fi

  if [[ -z "${tg_version}" ]]; then
    log "tg_version is not set"
    exit 1
  fi

  if [[ -z "${tg_command}" ]]; then
    log "tg_command is not set"
    exit 1
  fi
  setup_git
  setup_pre_exec

  install_terraform "${tf_version}"
  install_terragrunt "${tg_version}"
  
  configurePrivateGitPath

  # add auto approve for apply and destroy commands
  if [[ "$tg_command" == "apply"* || "$tg_command" == "destroy"* || "$tg_command" == "run-all apply"* || "$tg_command" == "run-all destroy"* ]]; then
    local -r tg_arg_and_commands="${tg_command} -auto-approve --terragrunt-non-interactive"
  else
    local -r tg_arg_and_commands="${tg_command}"
  fi
  run_terragrunt "${tg_dir}" "${tg_arg_and_commands}"

  local -r log_file="${terragrunt_log_file}"
  trap 'rm -rf ${log_file}' EXIT

  local exit_code
  exit_code=$(("${terragrunt_exit_code}"))

  local terragrunt_log_content
  terragrunt_log_content=$(cat "${log_file}")
  # output without colors
  local terragrunt_output
  terragrunt_output=$(clean_colors "${terragrunt_log_content}")

  if [[ "${tg_comment}" == "1" ]]; then
    comment "Execution result of \`$tg_command\` in \`${tg_dir}\` :
\`\`\`
${terragrunt_output}
\`\`\`
    "
  fi

  echo "tg_action_exit_code=${exit_code}" >> "${GITHUB_OUTPUT}"

  local tg_action_output
  tg_action_output=$(clean_multiline_text "${terragrunt_output}")
  echo "tg_action_output=${tg_action_output}" >> "${GITHUB_OUTPUT}"

  exit $exit_code
}

main "$@"
