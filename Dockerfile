# Dockerfile used in execution of Github Action
FROM gruntwork/terragrunt:0.0.1
MAINTAINER Gruntwork <info@gruntwork.io>

COPY ["./src/main.sh", "/action/main.sh"]

RUN addgroup --system --gid 1001 runner
RUN adduser --system --uid 1001 runner
USER runner

ENTRYPOINT ["/action/main.sh"]
