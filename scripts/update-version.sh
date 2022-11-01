#!/bin/bash -e
cd "$(dirname $0)"
# 0.25

majorVersion="$(grep MajorVersion ../common/version.go | awk '{print $4}' | sed 's/"//g')"
versionString="$majorVersion.$(date +%s)"
goLiteral="const VersionString = \"$versionString\""

sed -i "" "s/^const VersionString = \".*\"$/$goLiteral/" ../common/version.go