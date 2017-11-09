FROM ubuntu:14.04

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        make \
        git \
        ruby \
        ruby-dev \
        ruby-ffi \
        binutils \
        reprepro \
        wget \
        && \
    apt-get clean

RUN gem install --no-ri --no-rdoc fpm

WORKDIR /home