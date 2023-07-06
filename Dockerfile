# Dockerfile used in execution of Github Action
FROM gruntwork/terragrunt:0.0.1
MAINTAINER Gruntwork <info@gruntwork.io>

# Avoid git permissions warnings
RUN git config --global --add safe.directory /github/workspace

COPY ["./src/main.sh", "/action/main.sh"]
ENTRYPOINT ["/action/main.sh"]
