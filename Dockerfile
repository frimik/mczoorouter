FROM ubuntu:18.04

ENV LANG=C.UTF-8

RUN apt update && \
    apt dist-upgrade -y && \
    apt install -y --no-install-recommends ca-certificates wget gnupg && \
    wget -O - https://facebook.github.io/mcrouter/debrepo/bionic/PUBLIC.KEY | apt-key add && \
    echo "deb https://facebook.github.io/mcrouter/debrepo/bionic bionic contrib" >> /etc/apt/sources.list && \
    apt update && \
    apt install -y mcrouter && \
    apt remove -y wget gnupg && \
    apt autoremove -y && \
    apt clean all

USER 1001
