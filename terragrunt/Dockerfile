# Container to run Terragrunt and Terraform
# Contains inside mise to allow users to install custom Terraform and Terragrunt versions
FROM ubuntu:22.04
LABEL maintainer "Gruntwork <info@gruntwork.io>"

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

# # Create runner user
# RUN addgroup --system --gid 127 docker \
#     && useradd --system -u 1001 -g 127 -ms /bin/bash runner \
#     && usermod -aG sudo runner && echo 'runner ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
# USER runner
# WORKDIR /home/runner

RUN mkdir -p /home/runner/.ssh
COPY ./known_hosts /home/runner/.ssh/known_hosts

# install mise
RUN wget -q "https://github.com/jdx/mise/releases/download/${MISE_VERSION_INSTALL}/mise-${MISE_VERSION_INSTALL}-linux-x64" -O "/usr/bin/mise" \
    && chmod u+x "/usr/bin/mise"

# Running action as runner user
# https://docs.github.com/en/actions/creating-actions/dockerfile-support-for-github-actions#user
ENV PATH="~/.local/share/mise/shims:~/mise:${PATH}"

ENV TF_INPUT=false
ENV TF_IN_AUTOMATION=1
