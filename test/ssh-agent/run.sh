#!/usr/bin/env bash
set -x

cleanup() {
    echo "Caught SIGTERM signal! Cleaning up..."
    ssh-agent -k
    exit 0
}

trap cleanup SIGTERM SIGINT

mkdir -p ~/.ssh
cp -rfv ~/keys/* ~/.ssh/
chown -R $(whoami) ~/.ssh

# delete socket if exists
if [[ -S "$SSH_AUTH_SOCK" ]]; then
  rm -f "$SSH_AUTH_SOCK"
fi

eval "$(ssh-agent -s -a "$SSH_AUTH_SOCK")"
ssh-add ~/.ssh/*

echo "SSH agent running on: $SSH_AUTH_SOCK"
tail -f /dev/null
