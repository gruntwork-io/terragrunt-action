# Dockerfile used in execution of Github Action
FROM clarify-published-docker-image

COPY ["./src/main.sh", "/action/main.sh"]
ENTRYPOINT ["/action/main.sh"]