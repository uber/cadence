#!/bin/sh

MODE=$1

if [ "$MODE" != "LDFLAG" ] && [ "$MODE" != "ECHO" ]; then
    echo "Usage: $0 <LDFLAG|ECHO>"
    exit 1
fi

export GIT_REVISION=$(git rev-parse --short HEAD)
export GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
export BUILD_DATE=$(date '+%F-%T') # outputs something in this format 2017-08-21-18:58:45
export BUILD_TS_UNIX=$(date '+%s') # second since epoch
export BASE_PACKAGE=github.com/uber/cadence/common/metrics
if [ -z ${SERVER_VERSION} ]; then
  # If not set SERVER_VERSION, then use the most recent tag.
  # If the tag is on the tip, then Version is the tag name, eg "v0.10.0",
  # otherwise version is tag-commitID: eg "v0.10.0-935-g89d99359"
  export GIT_VERSION=$(git describe --tags 2>/dev/null || echo unknown)
else
  # If passing a version explicitly, then use it
  export GIT_VERSION=${SERVER_VERSION}
fi


if [ "$MODE" = "LDFLAG" ]; then
  LD_FLAGS="-X ${BASE_PACKAGE}.Revision=${GIT_REVISION} \
  -X ${BASE_PACKAGE}.Branch=${GIT_BRANCH}               \
  -X ${BASE_PACKAGE}.Version=${GIT_VERSION}             \
  -X ${BASE_PACKAGE}.BuildDate=${BUILD_DATE}            \
  -X ${BASE_PACKAGE}.BuildTimeUnix=${BUILD_TS_UNIX}"

  echo $LD_FLAGS
fi

if [ "$MODE" = "ECHO" ]; then
  cat <<EOF
BASE_PACKAGE=${BASE_PACKAGE}
GIT_REVISION=${GIT_REVISION}
GIT_BRANCH=${GIT_BRANCH}
GIT_VERSION=${GIT_VERSION}
BUILD_DATE=${BUILD_DATE}
BUILD_TS_UNIX=${BUILD_TS_UNIX}
SERVER_VERSION=${SERVER_VERSION}
EOF
fi
