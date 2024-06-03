# Container to run Terragrunt and Terraform
# Contains inside mise to allow users to install custom Terraform and Terragrunt versions
FROM ubuntu:22.04
LABEL maintainer="Gruntwork <info@gruntwork.io>"

ARG MISE_VERSION_INSTALL=v2024.4.0

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install -y \
    curl \
    git \
    jq \
    unzip \
    wget \
    sudo \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /root/.ssh
COPY ./known_hosts /root/.ssh/known_hosts
RUN chown -R root:root /root/.ssh
# install mise
RUN wget -q "https://github.com/jdx/mise/releases/download/${MISE_VERSION_INSTALL}/mise-${MISE_VERSION_INSTALL}-linux-x64" -O "/usr/bin/mise" \
    && chmod +x "/usr/bin/mise"

ENV PATH="~/.local/share/mise/shims:${PATH}"

ENV TF_INPUT=false
ENV TF_IN_AUTOMATION=1
