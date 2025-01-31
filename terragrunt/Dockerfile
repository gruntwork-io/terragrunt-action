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
RUN curl https://mise.run | MISE_VERSION=${MISE_VERSION_INSTALL} MISE_INSTALL_PATH=/usr/bin/mise sh

ENV PATH="~/.local/share/mise/shims:${PATH}"

ENV TF_INPUT=false
ENV TF_IN_AUTOMATION=1
