FROM golang:1.17

# Tried to set Python to ignore warnings due to the instructions at this link:
# https://github.com/yaml/pyyaml/wiki/PyYAML-yaml.load(input)-Deprecation
# But this causes all the pip installs to fail, so don't do this:
# ENV PYTHONWARNINGS=ignore::yaml.YAMLLoadWarning
# ENV PYTHONWARNINGS=ignore

RUN apt-get update && apt-get install -y --no-install-recommends \
      curl \
      gettext-base \
      libyaml-dev \
      python \
      python-dev \
      python-setuptools \
      time \
      unzip \
    && rm -rf /var/lib/apt/lists/*

# pip 21+ drops support for python2 entirely, we seem to need at least 19 to build cqlsh
RUN curl https://bootstrap.pypa.io/pip/2.7/get-pip.py -o get-pip.py
RUN python get-pip.py
RUN pip install -U 'pip<21'
RUN pip install PyYAML==3.13 cqlsh==5.0.4

# verbose test output from `make`, can be disabled with VERBOSE=0
ENV VERBOSE=1

# https://github.com/docker-library/golang/blob/c1baf037d71331eb0b8d4c70cff4c29cf124c5e0/1.4/Dockerfile
RUN mkdir -p /cadence
WORKDIR /cadence
