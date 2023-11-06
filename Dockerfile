# Dockerfile used in execution of Github Action
FROM gruntwork/terragrunt:0.0.2
MAINTAINER Gruntwork <info@gruntwork.io>

COPY ["./src/main.sh", "/action/main.sh"]

ENTRYPOINT ["/action/main.sh"]
