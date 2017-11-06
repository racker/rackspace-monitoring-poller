FROM ubuntu

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        make \
        git \
        ruby \
        ruby-dev \
        ruby-ffi && \
    apt-get clean

RUN gem install --no-ri --no-rdoc fpm
