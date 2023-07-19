# Dockerfile used in execution of Github Action
FROM gruntwork/terragrunt:0.0.1
MAINTAINER Gruntwork <info@gruntwork.io>

COPY ["./src/main.sh", "/action/main.sh"]
COPY /home/runner/.gitconfi[g] /root/
ENTRYPOINT ["/action/main.sh"]
