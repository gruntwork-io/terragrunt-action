FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y openssh-client && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir /ssh-agent

ENV SSH_AUTH_SOCK /tmp/ssh-agent.sock
ADD run.sh /run.sh
CMD ["/run.sh"]
