# Dockerfile used in execution of Github Action
FROM nixos/nix:2.26.3

RUN nix-env -i terragrunt opentofu sudo curl gnused jq git

COPY ["./src/main.sh", "/action/main.sh"]
ENTRYPOINT ["/action/main.sh"]
