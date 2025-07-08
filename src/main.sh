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

# run terragrunt commands in specified directory
# arguments: directory and terragrunt command
# output variables:
# terragrunt_log_file path to log file
# terragrunt_exit_code exit code of terragrunt command
function run_terragrunt {
  local -r dir="$1"
  local -a command
  read -ra command <<< "$2"

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
  local -r escaped_message=$(printf '%s' "$message" | sed 's/\\/\\\\/g; s/"/\\"/g; s/$/\\n/g; s/\t/\\t/g' | tr -d '\n')
  local -r tmpfile=$(mktemp)
  echo "{\"body\": \"$escaped_message\"}" > "$tmpfile"
  curl -s -S -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" -d @"$tmpfile" "$comment_url"
  rm "$tmpfile"
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

# Check minimum supported version of Terragrunt
function check_minimum_supported_version {
  local -r min_version="0.77.22"
  local tg_version

  # Try to get terragrunt version, but fail open if we can't determine it
  if ! tg_version=$(terragrunt --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'); then
    log "Warning: Could not determine Terragrunt version, continuing anyway"
    return 0
  fi

  # If we got an empty version string, fail open
  if [[ -z "${tg_version}" ]]; then
    log "Warning: Could not parse Terragrunt version, continuing anyway"
    return 0
  fi

  # Only check version if we successfully determined it
  if [[ "$(printf '%s\n' "$min_version" "$tg_version" | sort -V | head -n1)" != "$min_version" ]]; then
    log "Terragrunt version $tg_version is less than the minimum required version $min_version"
    exit 1
  fi

  log "Terragrunt version $tg_version meets minimum requirement $min_version"
}

# Check Terragrunt is installed
function check_terragrunt_installed {
  if ! command -v terragrunt &> /dev/null; then
    log "Terragrunt is not installed"
    exit 1
  fi
}

function main {
  log "Starting Terragrunt Action"
  trap 'log "Finished Terragrunt Action execution"' EXIT
  local -r tg_command=${INPUT_TG_COMMAND}
  local -r tg_comment=${INPUT_TG_COMMENT:-0}
  local -r tg_add_approve=${INPUT_TG_ADD_APPROVE:-1}
  local -r tg_dir=${INPUT_TG_DIR:-${GITHUB_WORKSPACE}} # use GitHub workspace as default

  if [[ -z "${tg_command}" ]]; then
    log "tg_command is not set"
    exit 1
  fi

  check_terragrunt_installed
  check_minimum_supported_version

  setup_pre_exec

  # add auto approve for apply and destroy commands
  local tg_command_and_args="${tg_command}"

  if [[ "$tg_command" == "apply"* || "$tg_command" == "destroy"* ]]; then
    if [[ "${tg_add_approve}" == "1" ]]; then
      local approvePattern="^(apply|destroy)"
      # split command and arguments to insert -auto-approve
      if [[ $tg_command_and_args =~ $approvePattern ]]; then
          local matchedCommand="${BASH_REMATCH[0]}"
          local remainingArgs="${tg_command_and_args#$matchedCommand}"
          tg_command_and_args="${matchedCommand} -auto-approve ${remainingArgs}"
      fi
    fi
  fi

  run_terragrunt "${tg_dir}" "${tg_command_and_args}"
  setup_post_exec

  local -r log_file="${terragrunt_log_file}"
  trap 'rm -rf ${log_file}' EXIT

  local exit_code
  exit_code="${terragrunt_exit_code:-0}"

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
