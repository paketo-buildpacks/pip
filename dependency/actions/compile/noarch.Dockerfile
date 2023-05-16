FROM ubuntu:jammy

ENV DEBIAN_FRONTEND noninteractive
ENV LC_CTYPE = 'en_US.UTF'

RUN apt-get update && apt install python3-pip -y

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
