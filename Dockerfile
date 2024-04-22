# Dockerfile used in execution of Github Action
FROM arsci/tg:v0.1
LABEL maintainer "Gruntwork <info@gruntwork.io>"

COPY ["./src/main.sh", "/action/main.sh"]

ENTRYPOINT ["/action/main.sh"]
