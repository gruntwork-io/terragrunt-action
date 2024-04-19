# Dockerfile used in execution of Github Action
FROM gruntwork/terragrunt:0.0.2
LABEL maintainer "Gruntwork <info@gruntwork.io>"

ARG MISE_VERSION_INSTALL=v2024.4.0

RUN mkdir -p "${HOME}/mise"
RUN wget -q https://github.com/jdx/mise/releases/download/${MISE_VERSION_INSTALL}/mise-${MISE_VERSION_INSTALL}-linux-x64 -O /${HOME}/mise/mise
RUN chmod u+x ${HOME}/mise/mise

ENV MISE_CONFIG_DIR=~/.config/mise
ENV MISE_STATE_DIR=~/.local/state/mise
ENV MISE_DATA_DIR=~/.local/share/mise
ENV MISE_CACHE_DIR=~/.cache/mise
ENV ASDF_HASHICORP_TERRAFORM_VERSION_FILE=.terraform-version

ENV PATH="/home/runner/.local/share/mise/shims:/home/runner/mise:${PATH}"

COPY ["./src/main.sh", "/action/main.sh"]

ENTRYPOINT ["/action/main.sh"]
