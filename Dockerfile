# Dockerfile used in execution of Github Action
FROM clarify-published-docker-image
MAINTAINER Gruntwork <info@gruntwork.io>

COPY ["./src/main.sh", "/action/main.sh"]
ENTRYPOINT ["/action/main.sh"]
