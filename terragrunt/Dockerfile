# Container to run Terragrunt and Terraform
# Contains inside TFenv and TGSwitch to allow users to install custom versions
FROM ubuntu:22.04
MAINTAINER Gruntwork <info@gruntwork.io>

ARG TF_ENV_VERSION=v3.0.0
ARG TGSWITCH_VERSION=0.6.0

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    curl \
    git \
    jq \
    unzip \
    wget \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -ms /bin/bash terragrunt
USER terragrunt

# clone tfenv
RUN git clone --depth=1 --branch $TF_ENV_VERSION https://github.com/tfutils/tfenv.git ~/.tfenv
RUN echo 'export PATH="$HOME/.tfenv/bin:$PATH"' >> ~/.bash_profile

# install tgswitch
RUN mkdir -p "$HOME/tgswitch"
RUN wget -q https://github.com/warrensbox/tgswitch/releases/download/${TGSWITCH_VERSION}/tgswitch_${TGSWITCH_VERSION}_linux_amd64.tar.gz -O /tmp/tgswitch_${TGSWITCH_VERSION}_linux_amd64.tar.gz
RUN tar -xzf /tmp/tgswitch_${TGSWITCH_VERSION}_linux_amd64.tar.gz -C $HOME/tgswitch
RUN chmod u+x $HOME/tgswitch/tgswitch
RUN rm -rf /tmp/tgswitch_${TGSWITCH_VERSION}_linux_amd64.tar.gz

ENV PATH="/home/terragrunt/.tfenv/bin:/home/terragrunt/tgswitch:/home/terragrunt/bin:${PATH}"

ENV TF_INPUT=false
ENV TF_IN_AUTOMATION=1
