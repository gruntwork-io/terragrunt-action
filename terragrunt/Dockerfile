# Container to run Terragrunt and Terraform
# Contains inside mise to allow users to install custom Terraform and Terragrunt versions
FROM ubuntu:22.04
MAINTAINER Gruntwork <info@gruntwork.io>

ARG MISE_VERSION_INSTALL=v2024.4.0

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    curl \
    git \
    jq \
    unzip \
    wget \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Create runner user
RUN addgroup --system --gid 127 docker
RUN useradd --system -u 1001 -g 127 -ms /bin/bash runner
RUN usermod -aG sudo runner && echo 'runner ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER runner
WORKDIR /home/runner

RUN mkdir -p /home/runner/.ssh
COPY ./known_hosts /home/runner/.ssh/known_hosts

# install mise
RUN mkdir -p "${HOME}/mise"
RUN wget -q https://github.com/jdx/mise/releases/download/${MISE_VERSION_INSTALL}/mise-${MISE_VERSION_INSTALL}-linux-x64 -O /${HOME}/mise/mise
RUN chmod u+x ${HOME}/mise/mise

ENV MISE_CONFIG_DIR=~/.config/mise
ENV MISE_STATE_DIR=~/.local/state/mise
ENV MISE_DATA_DIR=~/.local/share/mise
ENV MISE_CACHE_DIR=~/.cache/mise
ENV ASDF_HASHICORP_TERRAFORM_VERSION_FILE=.terraform-version

# Running action as runner user
# https://docs.github.com/en/actions/creating-actions/dockerfile-support-for-github-actions#user
ENV PATH="/home/runner/.local/share/mise/shims:/home/runner/mise:${PATH}"

ENV TF_INPUT=false
ENV TF_IN_AUTOMATION=1
