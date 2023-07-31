FROM golang:1.20

# Tried to set Python to ignore warnings due to the instructions at this link:
# https://github.com/yaml/pyyaml/wiki/PyYAML-yaml.load(input)-Deprecation
# But this causes all the pip installs to fail, so don't do this:
# ENV PYTHONWARNINGS=ignore::yaml.YAMLLoadWarning
# ENV PYTHONWARNINGS=ignore

RUN rm -rf /var/lib/apt/lists/* && apt-get update && apt-get install -y --no-install-recommends \
      curl \
      gettext-base \
      libyaml-dev \
      python3-pip \
      time \
      unzip \
    && rm -rf /var/lib/apt/lists/*

# We want to install cqlsh to system.
# After python 3.11 that is restricted and advised to use use specific versions, but we are using a single user setup, so there is no point of it.
ENV PIP_BREAK_SYSTEM_PACKAGES 1

RUN pip install --user cqlsh
# verbose test output from `make`, can be disabled with V=0
ENV V=1

# https://github.com/docker-library/golang/blob/c1baf037d71331eb0b8d4c70cff4c29cf124c5e0/1.4/Dockerfile
RUN mkdir -p /cadence
WORKDIR /cadence
