# Copyright (c) 2017 Uber Technologies, Inc.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

FROM cassandra:3.9

# get golang 1.8.1
RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
		curl \
		git libsnappy-dev zlib1g-dev libbz2-dev libgflags-dev gettext \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.8.1
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 a579ab19d5237e263254f1eac5352efcf1d70b9dacadb6d6bb12b0911ede8994

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# get and compile cadence-server
ENV CADENCE_HOME $GOPATH/src/github.com/uber/cadence

RUN go get -u github.com/Masterminds/glide
RUN go get -u github.com/golang/lint/golint
RUN git clone --depth=50 https://github.com/uber/cadence.git $CADENCE_HOME
RUN cd $CADENCE_HOME; make bins_nothrift

EXPOSE 7933 7934 7935

COPY ./start.sh $CADENCE_HOME/start.sh
COPY ./config_template.yaml $CADENCE_HOME/config/docker_template.yaml
RUN chmod a+x $CADENCE_HOME/start.sh

WORKDIR $CADENCE_HOME
CMD ./start.sh $CADENCE_HOME