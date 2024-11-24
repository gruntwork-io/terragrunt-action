# Dockerfile used in execution of Github Action
FROM nixos/nix:2.25.2

# NOTE(pete): Pin to a specific version for reproducibility.  This SHA was latest as of 20241122
RUN nix-channel --update nixpkgs https://github.com/NixOS/nixpkgs/archive/23e89b7da85c3640bbc2173fe04f4bd114342367.tar.gz

RUN nix-env -i terragrunt opentofu sudo curl

COPY ["./src/main.sh", "/action/main.sh"]
ENTRYPOINT ["/action/main.sh"]
