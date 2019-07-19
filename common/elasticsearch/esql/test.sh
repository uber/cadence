#!/bin/bash

dep ensure
go build
golint *.go

if [ $# -eq 0 ]; then
    go test
else
    for testName in "$@"; do
        go test -failfast -run "$testName"
    done
fi
