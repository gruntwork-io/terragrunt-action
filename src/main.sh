#!/usr/bin/env bash
set -e
[[ "${TRACE}" == "1" ]] && set -x

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
  output="${output//$'<'/'%3C'}"
  echo "${output}"
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
  sudo git config --global --add safe.directory /github/workspace
  # Also trust any subfolder within workspace
  sudo git config --global --add safe.directory "*"
}

function setup_permissions {
  local -r dir="${1}"
  # Set permissions for the working directory
  if [[ -f "${dir}" ]]; then
    sudo chown -R $(whoami) "${dir}"
    sudo chmod -R o+rw "${dir}"
  fi
  # Set permissions for the output file
  if [[ -f "${GITHUB_OUTPUT}" ]]; then
    sudo chown -R $(whoami) "${GITHUB_OUTPUT}"
  fi
  # set permissions for .terraform directories, if any
  find . -name ".terraform" -exec sudo chmod -R 777 {} \;
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

# Run INPUT_POST_EXEC_* environment variables as Bash code
function setup_post_exec {
  # Get all environment variables that match the pattern INPUT_POST_EXEC_*
  local -r post_exec_vars=$(env | grep -o '^INPUT_POST_EXEC_[0-9]\+' | sort)
  # Loop through each pre-execution variable and execute its value (Bash code)
  local post_exec_command
  while IFS= read -r post_exec_var; do
    if [[ -n "${post_exec_var}" ]]; then
      log "Evaluating ${post_exec_var}"
      post_exec_command="${!post_exec_var}"
      eval "$post_exec_command"
    fi
  done <<< "$post_exec_vars"
}

function main {
  log "Starting Terragrunt Action"
  trap 'log "Finished Terragrunt Action execution"' EXIT
  local -r tf_version=${INPUT_TF_VERSION}
  local -r tg_version=${INPUT_TG_VERSION}
  local -r tg_command=${INPUT_TG_COMMAND}
  local -r tg_comment=${INPUT_TG_COMMENT:-0}
  local -r tg_add_approve=${INPUT_TG_ADD_APPROVE:-1}
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
  setup_permissions "${tg_dir}"
  trap 'setup_permissions $tg_dir ' EXIT
  setup_pre_exec

  install_terraform "${tf_version}"
  install_terragrunt "${tg_version}"

  # add auto approve for apply and destroy commands
  local tg_arg_and_commands="${tg_command}"
  if [[ "$tg_command" == "apply"* || "$tg_command" == "destroy"* || "$tg_command" == "run-all apply"* || "$tg_command" == "run-all destroy"* ]]; then
    export TERRAGRUNT_NON_INTERACTIVE=true
    export TF_INPUT=false
    export TF_IN_AUTOMATION=1

    if [[ "${tg_add_approve}" == "1" ]]; then
      local approvePattern="^(apply|destroy|run-all apply|run-all destroy)"
      # split command and arguments to insert -auto-approve
      if [[ $tg_arg_and_commands =~ $approvePattern ]]; then
          local matchedCommand="${BASH_REMATCH[0]}"
          local remainingArgs="${tg_arg_and_commands#$matchedCommand}"
          tg_arg_and_commands="${matchedCommand} -auto-approve ${remainingArgs}"
      fi
    fi
  fi
  run_terragrunt "${tg_dir}" "${tg_arg_and_commands}"
  setup_permissions "${tg_dir}"
  # setup permissions for the output files
  setup_post_exec

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
    comment "<details>
<summary>Execution result of \"$tg_command\" in \"${tg_dir}\"</summary>

\`\`\`terraform
${terragrunt_output}
\`\`\`

</details>
    "
  fi

  echo "tg_action_exit_code=${exit_code}" >> "${GITHUB_OUTPUT}"

  local tg_action_output
  tg_action_output=$(clean_multiline_text "${terragrunt_output}")
  echo "tg_action_output=${tg_action_output}" >> "${GITHUB_OUTPUT}"

  exit $exit_code
}

main "$@"
