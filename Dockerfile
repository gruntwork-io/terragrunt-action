# Dockerfile used in execution of Github Action
FROM nixos/nix:2.25.2

RUN nix-channel --update
RUN nix-env -i terragrunt opentofu sudo

COPY ["./src/main.sh", "/action/main.sh"]
ENTRYPOINT ["/action/main.sh"]
